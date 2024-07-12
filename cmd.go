package hubur

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func Cmd(cmd *exec.Cmd) (exitCode int, exitData string) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ex, ok := err.(*exec.ExitError); ok {
			exitCode = ex.Sys().(syscall.WaitStatus).ExitStatus()
		}
	}
	return exitCode, string(out)
}

// 守护进程的命令执行这个
// cmd 携带context 对象
// @params exitConditions 响应内容里包含指定内容就退出
// @params verbose 是否实时打印响应内容
// @params duration 持续的时长
func CmdPipe(cmd *exec.Cmd, exitConditions string, verbose bool, duration int) (out string, err error) {
	var flag bool
	var builder strings.Builder
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
	}

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	// 从管道中实时获取输出并打印到终端
	reader := bufio.NewReader(stdout)
	after := time.After(time.Duration(duration) * time.Second)
	for {
		select {
		case <-after:
			flag = true
			cmd.Process.Signal(syscall.SIGKILL)
			break
		case <-done:
			flag = true
			break
		default:
			line, err := reader.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			if verbose {
				fmt.Print(line)
			}
			builder.WriteString(line)
			if strings.Contains(line, exitConditions) {
				if err = cmd.Process.Signal(syscall.SIGKILL); err != nil {
					fmt.Println("退出程序进程错误", err)
				}
				break
			}
		}
		if flag {
			break
		}
	}
	if runtime.GOOS == "windows" {
		processName := FindWindowsProcessName(cmd.Args)
		if len(processName) > 0 {
			pid, _ := FindWindowsProcessID(processName)
			if pid > 0 {
				fmt.Println("发现残留进程", pid)
				process, _ := os.FindProcess(pid)
				process.Signal(syscall.SIGKILL)
			}
		}
	}

	return builder.String(), nil
}

// 查找windows下指定程序名的进程号
func FindWindowsProcessID(processName string) (int, error) {
	buf := bytes.Buffer{}
	cmd := exec.Command("wmic", "process", "get", "name,processid")
	cmd.Stdout = &buf
	cmd.Run()

	cmd2 := exec.Command("findstr", processName)
	cmd2.Stdin = &buf
	data, _ := cmd2.CombinedOutput()
	if len(data) == 0 {
		return -1, errors.New("not find")
	}
	info := string(data)
	reg := regexp.MustCompile(`[0-9]+`)
	pid := reg.FindString(info)
	return strconv.Atoi(pid)
}

func FindWindowsProcessName(args []string) string {
	for _, arg := range args {
		if strings.Contains(arg, ".exe") {
			return arg
		}
	}
	return ""
}
