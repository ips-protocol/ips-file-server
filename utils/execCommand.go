package utils

import (
	"os/exec"
)

// 执行命令
// 执行失败时返回 err
// 成功时返回对应的控制台输出
func ExecCommand(commandName string) (result string, err error) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command("/bin/bash", "-c", commandName)

	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	result = string(bytes)
	return
}
