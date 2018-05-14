package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"testshell"
)

func ScanLine() string {
	var c byte
	var err error
	var b []byte
	for err == nil {
		_, err = fmt.Scanf("%c", &c)

		if c != '\n' {
			b = append(b, c)
		} else {
			break
		}
	}

	return string(b)
}

func execFromFile(fileName string) {
	//从文件中读
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
		if err != nil || io.EOF == err {
			break
		}
		fmt.Print(line)
		if testshell.Exec(line) {
			break
		}
	}
}

func execFromStdIO() {
	for {
		fmt.Print("TestShell> ")
		line := ScanLine()
		if testshell.Exec(line) {
			break
		}
	}
}

func main() {

	switch len(os.Args) {
	case 1:
		execFromStdIO()
	case 2:
		execFromFile(os.Args[1])
	default:
		fmt.Println("参数错误")
		os.Exit(1)
	}
}
