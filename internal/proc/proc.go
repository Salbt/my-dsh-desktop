package proc

import (
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"

	"my-dsh-desktop/internal/logx"
	"my-dsh-desktop/internal/nodejs"
)

type Manager struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	exited chan struct{}
}

func Start(exe string, args []string, env []string, logf nodejs.Logf) *Manager {
	m := &Manager{exited: make(chan struct{})}
	cmd := exec.Command(exe, args...)
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000200}
	cmd.Stdout = logx.Writer(logf)
	cmd.Stderr = logx.Writer(logf)
	if err := cmd.Start(); err != nil {
		logf("启动子进程失败: %v", err)
		close(m.exited)
		return m
	}
	m.cmd = cmd
	logf("子进程已启动: %s (PID=%d)", exe, cmd.Process.Pid)
	go func() {
		err := cmd.Wait()
		if err != nil {
			logf("子进程退出: %v", err)
		} else {
			logf("子进程正常退出")
		}
		m.mu.Lock()
		m.cmd = nil
		m.mu.Unlock()
		close(m.exited)
	}()
	return m
}

func (m *Manager) Exited() <-chan struct{} { return m.exited }

func (m *Manager) Stop() {
	m.mu.Lock()
	cmd := m.cmd
	m.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return
	}
	pid := strconv.Itoa(cmd.Process.Pid)
	kill := exec.Command("taskkill", "/T", "/F", "/PID", pid)
	kill.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000200}
	_ = kill.Run()
	select {
	case <-m.exited:
	case <-time.After(5 * time.Second):
	}
}
