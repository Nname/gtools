package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

// 正常退出
func main() {
	for i := 0; i < 100; i++ {
		fmt.Println("loop, ", i)
		go Loop()
		time.Sleep(time.Second * 60)
	}
}

func Loop() {
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		fmt.Println("client, ", i)
		wg.Add(1)
		go ClientSend(i, &wg)
	}
	wg.Wait()
	fmt.Println("done...............")
}

func ClientSend(i int, wg *sync.WaitGroup) {
	defer wg.Done()
	// 创建WebSocket连接
	url := "ws://192.168.2.177:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("failed to connect to WebSocket server:", err)
		return
	}
	defer conn.Close()

	writeErr := conn.WriteMessage(websocket.TextMessage, []byte(`{"home":"/tmp","name":"./t.sh","envs":["a=1"],"args":["x=2"]}`))
	if writeErr != nil {
		fmt.Println("err, ", writeErr)
		return
	}
	time.Sleep(time.Second * 60)
	fmt.Println("client close, ", i)
}
