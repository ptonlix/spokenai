package cmdexec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

type CmdExec struct {
	logger *zap.Logger
}

func NewCmdExec(logger *zap.Logger) (*CmdExec, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}
	return &CmdExec{logger: logger}, nil
}

// ExecuteCommand 执行命令并返回进程ID
func (c *CmdExec) ExecuteCommand(command string, args ...string) (int, error) {
	cmd := exec.Command(command, args...)
	err := cmd.Start()
	if err != nil {
		c.logger.Error("Execute command error:", zap.String("Error", fmt.Sprintf("%v+", err)))
		return 0, err
	}
	pid := cmd.Process.Pid
	return pid, nil
}

// TerminateProcess 根据进程ID终止进程
func (c *CmdExec) TerminateProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		c.logger.Error("Found not process", zap.String("Error", fmt.Sprintf("%v+", err)))
		return err
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		c.logger.Error("Terminate process Failed", zap.String("Error", fmt.Sprintf("%v+", err)))
		return err
	}

	// 等待进程退出
	_, err = process.Wait()
	if err != nil {
		c.logger.Error("Waiting process exit Failed", zap.String("Error", fmt.Sprintf("%v+", err)))
		return err
	}

	return nil
}

// ListRunningProcesses 获取所有运行中的进程
func (c *CmdExec) ListRunningProcesses() ([]string, error) {
	processes := []string{}

	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		c.logger.Error("Get processes list Failed", zap.String("Error", fmt.Sprintf("%v+", err)))
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for i := range lines {
		processes = append(processes, lines[i])
	}

	return processes, nil
}
