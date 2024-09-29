package main

import (
	_ "net/http/pprof"

	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gorilla/websocket"
)

/*
高并发下进程的开销还是很大
*/

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 4,
	WriteBufferSize: 1024 * 4,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type serverTaskReq struct {
	Home           string   `p:"home" v:"required"`
	Name           string   `p:"name" v:"required"`
	Args           []string `p:"args" v:"required"`
	Envs           []string `p:"envs" v:"required"`
	Client         string   `p:"client"`
	ClientProtocol string   `p:"client_protocol"`
}

type serverTaskExec struct {
	cmd *exec.Cmd
}

type serverMessage struct {
	msgType int
	msgData []byte
}

func serverTaskHandler(ch chan serverMessage, data serverTaskReq, chCmd chan serverTaskExec) {
	select {
	case ch <- serverMessage{websocket.TextMessage, gconv.Bytes(gjson.New(data).String())}:
	default:
		fmt.Println("serverTaskHandler send err, ", "len: ", cap(ch))
		return
	}
	shellTask := AppService{Home: data.Home, Name: data.Name, Args: data.Args, Env: data.Envs}
	go shellTask.StartCommandPipeCh(ch, chCmd)
}

type AppService struct {
	Home string
	Name string
	Args []string
	Env  []string
	Cmd  *exec.Cmd
}

func (c *AppService) StartCommandPipeCh(ch chan serverMessage, chCmd chan serverTaskExec) {
	shell := exec.Command(c.Name, c.Args...)
	shell.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	shell.Dir = c.Home
	shell.Env = c.Env
	c.Cmd = shell
	select {
	case chCmd <- serverTaskExec{cmd: shell}:
	default:
		fmt.Println("chCmd send err, ", "len: ", cap(chCmd))
		return
	}

	// stdout / stderr
	pipeOut, pipeOutErr := c.Cmd.StdoutPipe()
	if pipeOutErr != nil {
		fmt.Println("pipeOutErr, ", pipeOutErr)
		select {
		case ch <- serverMessage{websocket.TextMessage, gconv.Bytes(pipeOutErr)}:
		default:
			fmt.Println("pipeOutErr ch send err, ", "len: ", cap(ch))
			return
		}
		return
	}
	pipeErr, pipeErrErr := c.Cmd.StderrPipe()
	if pipeErrErr != nil {
		select {
		case ch <- serverMessage{websocket.TextMessage, gconv.Bytes(pipeErrErr)}:
		default:
			fmt.Println("pipeErrErr ch send err, ", "len: ", cap(ch))
			return
		}
		fmt.Println("pipeErrErr, ", pipeErrErr)
		return
	}

	// start
	startErr := shell.Start()
	if startErr != nil {
		fmt.Println("startErr, ", startErr)
		select {
		case ch <- serverMessage{websocket.TextMessage, gconv.Bytes(startErr)}:
		default:
			fmt.Println("startErr ch send err, ", "len: ", cap(ch))
			return
		}
		return
	}

	// stdout
	go func() {
		reader := bufio.NewReader(pipeOut)
		for {
			line, err := reader.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			fmt.Println("pipeOutReader, ", line)
			select {
			case ch <- serverMessage{websocket.TextMessage, gconv.Bytes(line)}:
			default:
				fmt.Println("pipeOutReader ch send err, ", "len: ", cap(ch))
				return
			}
		}
	}()

	// stderr
	go func() {
		reader := bufio.NewReader(pipeErr)
		for {
			line, err := reader.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			fmt.Println("pipeErrReader, ", line)
			select {
			case ch <- serverMessage{websocket.TextMessage, gconv.Bytes(line)}:
			default:
				fmt.Println("pipeErrReader ch send err, ", "len: ", cap(ch))
				return
			}
		}
	}()

	if waitErr := shell.Wait(); waitErr != nil {
		fmt.Println("waitErr, ", waitErr)
		select {
		case ch <- serverMessage{websocket.TextMessage, gconv.Bytes(waitErr)}:
		default:
			fmt.Println("waitErr ch send err, ", "len: ", cap(ch))
			return
		}
		return
	}
}

func ExistPidGroup(ch chan serverMessage, chCmd chan serverTaskExec) {
	close(chCmd)
	for taskExec := range chCmd {
		if taskExec.cmd != nil && taskExec.cmd.Process != nil {
			syscall.Kill(-taskExec.cmd.Process.Pid, syscall.SIGKILL)
		}
	}
	time.Sleep(time.Second * 3)
	close(ch)
}

func serverWrite(ch chan serverMessage, conn *websocket.Conn, wsLock *sync.Mutex) {
	for {
		data, ok := <-ch
		if !ok {
			return
		}
		switch data.msgType {
		case websocket.PingMessage, websocket.PongMessage:
			fmt.Println("PingMessage/PongMessage, ", data.msgType, string(data.msgData), time.Now())
			wsLock.Lock()
			err := conn.WriteMessage(data.msgType, data.msgData)
			wsLock.Unlock()
			if err != nil {
				fmt.Println("PingMessage/PongMessage write err, ", err)
				return
			}
		case websocket.TextMessage, websocket.BinaryMessage:
			fmt.Println("Message, ", data.msgType, string(data.msgData), time.Now())
			wsLock.Lock()
			err := conn.WriteMessage(data.msgType, data.msgData)
			wsLock.Unlock()
			if err != nil {
				fmt.Println("Message write err, ", err)
				return
			}
		default:
			fmt.Println("Other, ", data.msgType, string(data.msgData), time.Now())
			return
		}
	}
}

func serverPingHandler(conn *websocket.Conn, ch chan serverMessage, chCmd chan serverTaskExec, wsLock *sync.Mutex) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	defer ExistPidGroup(ch, chCmd)
	for {
		select {
		case <-ticker.C:
			wsLock.Lock()
			err := conn.WriteMessage(websocket.PingMessage, nil)
			wsLock.Unlock()
			if err != nil {
				fmt.Println("serverPingHandler err, ", err, conn.RemoteAddr().String(), conn.RemoteAddr().Network())
				return
			}
			fmt.Println("heatBeat ping success, ", conn.RemoteAddr().String(), conn.RemoteAddr().Network())
		}
	}
}

func serverTaskWebsocket(w http.ResponseWriter, r *http.Request) {
	// ws bug
	// defer gorillaContext.Clear(r)

	// 升级 HTTP 连接为 WebSocket 连接
	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	// websocket connect lock
	var wsLock sync.Mutex

	// 缓冲通道
	ch := make(chan serverMessage, 100)
	chCmd := make(chan serverTaskExec, 10)

	// 异步写
	go serverWrite(ch, conn, &wsLock)

	// 主动探测
	go serverPingHandler(conn, ch, chCmd, &wsLock)

	for {
		_, msgData, readErr := conn.ReadMessage()
		if readErr != nil {
			return
		}
		var taskData serverTaskReq
		if scanErr := gconv.Scan(msgData, &taskData); scanErr != nil {
			fmt.Println("scanErr, ", scanErr, "err client, ", conn.RemoteAddr().String(), conn.RemoteAddr().Network())
			continue
		}
		if validatorErr := g.Validator().Assoc(msgData).Data(taskData).Run(context.Background()); validatorErr != nil {
			fmt.Println("validatorErr, ", validatorErr, "err client, ", conn.RemoteAddr().String(), conn.RemoteAddr().Network())
			continue
		}
		taskData.Client = conn.RemoteAddr().String()
		taskData.ClientProtocol = conn.RemoteAddr().Network()
		go serverTaskHandler(ch, taskData, chCmd)
	}
}

func main() {
	//go clearStaleRequest()
	server := &http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/ws", serverTaskWebsocket)

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("start server err, ", err)
	}
	fmt.Println("start server ...")
}

//func clearStaleRequest() {
//	for {
//		time.Sleep(3 * time.Minute)
//		gorillaContext.Purge(45)
//	}
//}
