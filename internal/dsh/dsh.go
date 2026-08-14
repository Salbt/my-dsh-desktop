package dsh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"my-dsh-desktop/internal/logx"
	"my-dsh-desktop/internal/nodejs"
)

const (
	ctrlCExit       = 0xc000013a
	childFlags      = 0x08000200
	installAttempts = 3
)

type Dsh struct {
	BinJS string
	Dir   string
}

func (d *Dsh) LaunchArgs(port int) []string {
	return []string{d.BinJS, "web", "--port", strconv.Itoa(port)}
}

func BinPath(prefixDir string) string {
	return filepath.Join(prefixDir, "node_modules", "@deepseek-ai", "dsh", "lib", "bin.js")
}

func Ensure(node *nodejs.Node, prefixDir, version, registry, cacheDir string, logf nodejs.Logf) (*Dsh, error) {
	bin := BinPath(prefixDir)
	if _, err := os.Stat(bin); err == nil {
		logf("DeepSeek Harness 已安装 (%s)", installedVersion(prefixDir))
		return &Dsh{BinJS: bin, Dir: prefixDir}, nil
	}
	return install(node, prefixDir, version, registry, cacheDir, logf)
}

func Update(node *nodejs.Node, prefixDir, registry, cacheDir string, logf nodejs.Logf) (*Dsh, error) {
	return install(node, prefixDir, "latest", registry, cacheDir, logf)
}

func install(node *nodejs.Node, prefixDir, version, registry, cacheDir string, logf nodejs.Logf) (*Dsh, error) {
	logf("安装 DeepSeek Harness @deepseek-ai/dsh@%s ...", version)
	args := []string{"install", "--prefix", prefixDir, "--no-audit", "--no-fund", "--prefer-offline"}
	if cacheDir != "" {
		args = append(args, "--cache", cacheDir)
	}
	if registry != "" {
		args = append(args, "--registry", registry)
	}
	args = append(args, "@deepseek-ai/dsh@"+version)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	var runErr error
	for attempt := 1; attempt <= installAttempts; attempt++ {
		var cmd *exec.Cmd
		npmCli := node.NpmCliJS()
		if _, err := os.Stat(npmCli); err == nil {
			full := append([]string{npmCli}, args...)
			cmd = exec.CommandContext(ctx, node.Exe, full...)
		} else {
			cmd = exec.CommandContext(ctx, "npm", args...)
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: childFlags}
		cmd.Stdout = logx.Writer(logf)
		cmd.Stderr = logx.Writer(logf)
		runErr = cmd.Run()
		if runErr == nil {
			break
		}
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) && exitErr.ExitCode() == ctrlCExit && attempt < installAttempts {
			logf("npm 进程被中断，第 %d 次重试...", attempt+1)
			continue
		}
		return nil, fmt.Errorf("npm install 失败: %w", runErr)
	}
	if runErr != nil {
		return nil, fmt.Errorf("npm install 失败: %w", runErr)
	}
	bin := BinPath(prefixDir)
	if _, err := os.Stat(bin); err != nil {
		return nil, fmt.Errorf("安装完成但未找到 %s", bin)
	}
	return &Dsh{BinJS: bin, Dir: prefixDir}, nil
}

func InstalledVersion(prefixDir string) string {
	return installedVersion(prefixDir)
}

func installedVersion(prefixDir string) string {
	b, err := os.ReadFile(filepath.Join(prefixDir, "node_modules", "@deepseek-ai", "dsh", "package.json"))
	if err != nil {
		return "?"
	}
	var meta struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(b, &meta) != nil || meta.Version == "" {
		return "?"
	}
	return meta.Version
}
