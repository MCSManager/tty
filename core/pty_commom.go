package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"
)

var PtySize = ""
var Color = false

type DataProtocol struct {
	Type int    `json:"type"`
	Data string `json:"data"`
}

type cmdjson struct {
	Cmd []string `json:"cmd"`
}

func (pty *Pty) HandleStdIO() {
	go pty.handleStdIn()
	pty.handleStdOut()
}

func (pty *Pty) handleStdIn() {
	if PtySize == "" {
		pty.noSizeFlag()
	} else {
		pty.existSizeFlag()
	}
}

func (pty *Pty) noSizeFlag() {
	var err error
	var protocol DataProtocol
	var bufferText string
	inputReader := bufio.NewReader(os.Stdin)
	for {
		bufferText, _ = inputReader.ReadString('\n')
		err = json.Unmarshal([]byte(bufferText), &protocol)
		if err != nil {
			fmt.Printf("[MCSMANAGER-TTY] Unmarshall json err: %v\noriginal data: %#v\n", err, bufferText)
			fmt.Println("[MCSMANAGER-TTY] 正在使用 json 格式命令,具体格式请查看 github.com/MCSManager/pty")
			continue
		}
		switch protocol.Type {
		case 1:
			pty.StdIn.Write([]byte(protocol.Data))
		case 2:
			pty.ResizeWindow(&protocol.Data)
		case 3:
			pty.StdIn.Write([]byte{3})
		default:
		}
	}
}

func (pty *Pty) existSizeFlag() {
	// 删除 stdin cache，达到系统信号直接传递到 pty
	// 此方法操作到了文件描述符，不适用于父子进程操作
	// oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	// if err != nil {
	// 	panic(err)
	// }
	// defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	io.Copy(pty.StdIn, os.Stdin)
}

func (pty *Pty) handleStdOut() {
	var stdout io.Writer
	if Color {
		stdout = colorable.NewColorableStdout()
	} else {
		stdout = colorable.NewNonColorable(os.Stdout)
	}
	io.Copy(stdout, pty.StdOut)
}

// Set the PTY window size based on the text
func (pty *Pty) ResizeWindow(sizeText *string) {
	arr := strings.Split(*sizeText, ",")
	if len(arr) != 2 {
		fmt.Printf("[MCSMANAGER-TTY] Set tty size data failed,original data:%#v\n", *sizeText)
		return
	}
	cols, err1 := strconv.Atoi(arr[0])
	rows, err2 := strconv.Atoi(arr[1])
	if err1 != nil || err2 != nil {
		fmt.Printf("[MCSMANAGER-TTY] Failed to set window size,original data:%#v\n", *sizeText)
		return
	}
	pty.Setsize(uint32(cols), uint32(rows))
}
