package main

import (
	"os/exec"
	"os"
	"runtime"
)

var clearFuncMap = map[string]func(){}

func init() {
	clearFuncMap["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clearFuncMap["darwin"] = clearFuncMap["linux"]
	clearFuncMap["windows"] = func() {
		// TODO: 檢查是否有 cls 命令，若沒有提示用 Windows 自帶的 cmd.exe 打開助手
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func clearConsole() {
	if clearFunc, ok := clearFuncMap[runtime.GOOS]; ok {
		clearFunc()
	}
}
