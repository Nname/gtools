package command

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type AppService struct {
	Home string
	Name string
	Args []string
	Env  []string
	Cmd  *exec.Cmd
}

func (c *AppService) StdOutPipe() {
	pipeOut, err := c.Cmd.StdoutPipe()
	if err != nil {
		return
	}
	pipeOutScan := bufio.NewScanner(pipeOut)
	pipeOutScanBuf := make([]byte, 0, bufio.MaxScanTokenSize*10)
	pipeOutScan.Buffer(pipeOutScanBuf, cap(pipeOutScanBuf)*10)
	for pipeOutScan.Scan() {
		fmt.Println("StdoutPipe, ", pipeOutScan.Text())
	}
	pipeOutScanErr := pipeOutScan.Err()
	if pipeOutScanErr != nil {
		fmt.Println("StdoutPipeScanErr, ", pipeOutScanErr)
	}
}

func (c *AppService) StdErrPipe() {
	pipe, err := c.Cmd.StderrPipe()
	if err != nil {
		return
	}
	pipeScan := bufio.NewScanner(pipe)
	pipeScanBuf := make([]byte, 0, bufio.MaxScanTokenSize*10)
	pipeScan.Buffer(pipeScanBuf, cap(pipeScanBuf)*10)
	for pipeScan.Scan() {
		fmt.Println("StderrPipe, ", pipeScan.Text())
	}
	pipeScanErr := pipeScan.Err()
	if pipeScanErr != nil {
		fmt.Println("StderrPipeScanErr, ", pipeScanErr)
	}
}

func (c *AppService) StartCommandStd() {
	shell := exec.Command(c.Name, c.Args...)
	shell.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	shell.Dir = c.Home
	shell.Env = c.Env
	c.Cmd = shell
	shell.Stdout = os.Stdout
	shell.Stderr = os.Stderr
	shell.Run()
}

func (c *AppService) StartCommandPipe() {
	shell := exec.Command(c.Name, c.Args...)
	shell.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	shell.Dir = c.Home
	shell.Env = c.Env
	c.Cmd = shell

	// 异步
	var wg sync.WaitGroup
	wg.Add(2)

	// stdout
	pipeOut, pipeOutErr := c.Cmd.StdoutPipe()
	if pipeOutErr != nil {
		fmt.Println("pipeOutErr, ", pipeOutErr)
		return
	}
	pipeOutScan := bufio.NewScanner(pipeOut)
	pipeOutScanBuf := make([]byte, 0, bufio.MaxScanTokenSize*10)
	pipeOutScan.Buffer(pipeOutScanBuf, cap(pipeOutScanBuf)*10)
	go func() {
		defer wg.Done()
		for pipeOutScan.Scan() {
			fmt.Println("pipeOutScan, ", pipeOutScan.Text())
		}
	}()
	pipeOutScanErr := pipeOutScan.Err()
	if pipeOutScanErr != nil {
		fmt.Println("pipeOutScanErr, ", pipeOutScanErr)
		return
	}

	// stderr
	pipeErr, pipeErrErr := c.Cmd.StderrPipe()
	if pipeErrErr != nil {
		fmt.Println("pipeErrErr, ", pipeErrErr)
		return
	}
	pipeErrScan := bufio.NewScanner(pipeErr)
	pipeErrScanBuf := make([]byte, 0, bufio.MaxScanTokenSize*10)
	pipeErrScan.Buffer(pipeErrScanBuf, cap(pipeErrScanBuf)*10)
	go func() {
		defer wg.Done()
		for pipeErrScan.Scan() {
			fmt.Println("pipeErrScan, ", pipeErrScan.Text())
		}
	}()
	pipeErrScanErr := pipeErrScan.Err()
	if pipeErrScanErr != nil {
		fmt.Println("pipeErrScanErr, ", pipeErrScanErr)
		return
	}

	// run
	runErr := shell.Run()
	if runErr != nil {
		fmt.Println("runErr, ", runErr)
		return
	}
	wg.Wait()

	defer c.stopCommand()
}

func (c *AppService) StartCommandPipeCh() {
	shell := exec.Command(c.Name, c.Args...)
	shell.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	shell.Dir = c.Home
	shell.Env = c.Env
	c.Cmd = shell

	// 异步
	var wg sync.WaitGroup
	wg.Add(2)

	// stdout
	pipeOut, pipeOutErr := c.Cmd.StdoutPipe()
	if pipeOutErr != nil {
		fmt.Println("pipeOutErr, ", pipeOutErr)
		return
	}
	pipeOutScan := bufio.NewScanner(pipeOut)
	pipeOutScanBuf := make([]byte, 0, bufio.MaxScanTokenSize*10)
	pipeOutScan.Buffer(pipeOutScanBuf, cap(pipeOutScanBuf)*10)
	go func() {
		defer wg.Done()
		for pipeOutScan.Scan() {
			fmt.Println("pipeOutScan, ", pipeOutScan.Text())
		}
	}()
	pipeOutScanErr := pipeOutScan.Err()
	if pipeOutScanErr != nil {
		fmt.Println("pipeOutScanErr, ", pipeOutScanErr)
		return
	}

	// stderr
	pipeErr, pipeErrErr := c.Cmd.StderrPipe()
	if pipeErrErr != nil {
		fmt.Println("pipeErrErr, ", pipeErrErr)
		return
	}
	pipeErrScan := bufio.NewScanner(pipeErr)
	pipeErrScanBuf := make([]byte, 0, bufio.MaxScanTokenSize*10)
	pipeErrScan.Buffer(pipeErrScanBuf, cap(pipeErrScanBuf)*10)
	go func() {
		defer wg.Done()
		for pipeErrScan.Scan() {
			fmt.Println("pipeErrScan, ", pipeErrScan.Text())
		}
	}()
	pipeErrScanErr := pipeErrScan.Err()
	if pipeErrScanErr != nil {
		fmt.Println("pipeErrScanErr, ", pipeErrScanErr)
		return
	}

	// run
	runErr := shell.Run()
	if runErr != nil {
		fmt.Println("runErr, ", runErr)
		return
	}
	wg.Wait()

	defer c.stopCommand()
}

func (c *AppService) stopCommand() {
	fmt.Println("内部退出, ", c.Cmd.ProcessState.String(), c.Cmd.ProcessState.ExitCode())
	syscall.Kill(-c.Cmd.Process.Pid, syscall.SIGKILL)
}

func (c *AppService) StopCommand() {
	fmt.Println("外部退出, ", c.Cmd.ProcessState.String(), c.Cmd.ProcessState.ExitCode())
	syscall.Kill(-c.Cmd.Process.Pid, syscall.SIGKILL)
}

func (c *AppService) Demo() {
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>")
	go c.StartCommandPipe()
	time.Sleep(time.Second * 10)
	c.StopCommand()
}
