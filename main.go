package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"my-dsh-desktop/internal/appcfg"
	"my-dsh-desktop/internal/dsh"
	"my-dsh-desktop/internal/envx"
	"my-dsh-desktop/internal/logx"
	"my-dsh-desktop/internal/nodejs"
	"my-dsh-desktop/internal/port"
	"my-dsh-desktop/internal/proc"
	"my-dsh-desktop/internal/single"
	"my-dsh-desktop/internal/ui"
	"my-dsh-desktop/internal/winutil"
)

var (
	mu          sync.Mutex
	currentNode *nodejs.Node
	currentMgr  *proc.Manager
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--self-update" {
		os.Exit(selfUpdateMain(os.Args[2:]))
	}

	cfg, err := appcfg.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "初始化配置失败:", err)
		os.Exit(1)
	}
	if err := cfg.EnsureDirs(); err != nil {
		fmt.Fprintln(os.Stderr, "初始化目录失败:", err)
		os.Exit(1)
	}
	log, err := logx.New(filepath.Join(cfg.Dirs.Logs, "app.log"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "初始化日志失败:", err)
		os.Exit(1)
	}
	log.Printf("运行模式: portable=%v runtime=%s", cfg.Portable, cfg.Dirs.Runtime)
	if !single.Acquire("my-dsh-desktop-singleton") {
		log.Printf("检测到已有实例在运行，本次启动退出")
		os.Exit(0)
	}
	defer single.Release()

	win := ui.New("DeepSeek Harness Desktop", 1440, 900)
	win.SetLoading("正在启动 DeepSeek Harness...")

	go win.RunTray(func() {
		ui.OpenLogsDir(cfg.Dirs.Logs)
	}, func() {
		startUpdate(cfg, log, win)
	}, func() {
		log.Printf("从托盘退出")
		win.Terminate()
	})

	go func() {
		url, m, err := runPipeline(cfg, log, win.SetLoading)
		if err != nil {
			log.Printf("启动失败: %v", err)
			win.SetError(err.Error(), cfg.Dirs.Logs)
			return
		}
		mu.Lock()
		currentMgr = m
		mu.Unlock()
		win.Navigate(url)
	}()

	win.Run()
	log.Printf("窗口已关闭，正在停止 Harness")
	mu.Lock()
	mgr := currentMgr
	mu.Unlock()
	if mgr != nil {
		mgr.Stop()
	}
	ui.QuitTray()
	os.Exit(0)
}

func localAppDataDir() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "my-dsh-desktop")
}

func runPipeline(cfg *appcfg.Config, log *logx.Logger, status func(string)) (string, *proc.Manager, error) {
	status("检测 Node.js 运行环境...")
	node, err := nodejs.Ensure(cfg.Dirs.Node, filepath.Join(localAppDataDir(), "runtime", "node"), log.Printf)
	if err != nil {
		return "", nil, fmt.Errorf("Node.js 准备失败: %w", err)
	}
	mu.Lock()
	currentNode = node
	mu.Unlock()

	status("检查 DeepSeek Harness 安装...")
	d, err := dsh.Ensure(node, cfg.Dirs.Dsh, cfg.DSHVersion, cfg.Registry, cfg.Dirs.NpmCache, log.Printf)
	if err != nil {
		return "", nil, fmt.Errorf("DeepSeek Harness 安装失败: %w", err)
	}

	status("启动 DeepSeek Harness 服务...")
	return runDsh(cfg, node, d, log, status)
}

func runDsh(cfg *appcfg.Config, node *nodejs.Node, d *dsh.Dsh, log *logx.Logger, status func(string)) (string, *proc.Manager, error) {
	p, err := port.Free()
	if err != nil {
		return "", nil, fmt.Errorf("获取空闲端口失败: %w", err)
	}
	env := dshEnvironment(cfg, node)
	mgr := proc.Start(node.Exe, d.LaunchArgs(p), env, log.Printf)

	status("等待服务就绪...")
	url := fmt.Sprintf("http://127.0.0.1:%d", p)
	if err := port.WaitReady(url+"/", 300*time.Second, log.Printf); err != nil {
		select {
		case <-mgr.Exited():
			return "", nil, fmt.Errorf("Harness 进程已提前退出，请查看日志目录: %s", cfg.Dirs.Logs)
		default:
		}
		mgr.Stop()
		return "", nil, err
	}
	log.Printf("DeepSeek Harness 已就绪: %s", url)
	return url, mgr, nil
}

func dshEnvironment(cfg *appcfg.Config, node *nodejs.Node) []string {
	pathParts := []string{
		filepath.Dir(node.Exe),
		filepath.Join(filepath.Dir(cfg.Dirs.Runtime), "bin"),
		filepath.Join(cfg.Dirs.Pnpm, "node_modules", ".bin"),
	}
	if current := os.Getenv("PATH"); current != "" {
		pathParts = append(pathParts, current)
	}
	extra := []string{
		"DSH_HOME=" + cfg.Dirs.Home,
		"PATH=" + strings.Join(pathParts, string(os.PathListSeparator)),
		"NPM_CONFIG_CACHE=" + cfg.Dirs.NpmCache,
		"NPM_CONFIG_STORE_DIR=" + filepath.Join(cfg.Dirs.NpmCache, "pnpm-store"),
	}
	if cfg.Registry != "" {
		extra = append(extra, "NPM_CONFIG_REGISTRY="+cfg.Registry)
	}
	return envx.Merge(extra...)
}

func startUpdate(cfg *appcfg.Config, log *logx.Logger, win *ui.UI) {
	r := winutil.MessageBox(0, "将下载并安装最新版 DeepSeek Harness，期间服务会短暂中断。是否继续？", "更新 Harness", winutil.MB_YESNO|winutil.MB_ICONINFO)
	if r != winutil.IDYES {
		return
	}
	win.SetLoading("正在更新 DeepSeek Harness...")
	go func() {
		mu.Lock()
		node := currentNode
		mu.Unlock()
		if node == nil {
			n, err := nodejs.Ensure(cfg.Dirs.Node, filepath.Join(localAppDataDir(), "runtime", "node"), log.Printf)
			if err != nil {
				updateFailed(cfg, log, win, fmt.Errorf("Node.js 准备失败: %w", err))
				return
			}
			node = n
			mu.Lock()
			currentNode = n
			mu.Unlock()
		}

		mu.Lock()
		mgr := currentMgr
		mu.Unlock()
		if mgr != nil {
			log.Printf("停止当前服务以进行更新")
			mgr.Stop()
		}

		if winutil.IsDirWritable(cfg.Dirs.Dsh) {
			if _, err := dsh.Update(node, cfg.Dirs.Dsh, cfg.Registry, cfg.Dirs.NpmCache, log.Printf); err != nil {
				updateFailed(cfg, log, win, fmt.Errorf("更新失败: %w", err))
				return
			}
		} else {
			if err := elevatedUpdate(cfg, node, log); err != nil {
				updateFailed(cfg, log, win, err)
				return
			}
		}
		log.Printf("Harness 更新完成 (%s)，重启服务", dsh.InstalledVersion(cfg.Dirs.Dsh))
		restartAfterUpdate(cfg, log, win)
	}()
}

func elevatedUpdate(cfg *appcfg.Config, node *nodejs.Node, log *logx.Logger) error {
	log.Printf("runtime 目录不可写，请求管理员权限执行更新...")
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取自身路径失败: %w", err)
	}
	okPath := filepath.Join(cfg.Dirs.Dsh, "update.ok")
	errPath := filepath.Join(cfg.Dirs.Dsh, "update.err")
	_ = os.Remove(okPath)
	_ = os.Remove(errPath)
	args := []string{
		"--self-update",
		"--node", node.Exe,
		"--prefix", cfg.Dirs.Dsh,
		"--cache", cfg.Dirs.NpmCache,
		"--registry", cfg.Registry,
		"--log", filepath.Join(cfg.Dirs.Logs, "update.log"),
	}
	if err := winutil.RunElevated(exe, args); err != nil {
		return err
	}
	deadline := time.Now().Add(30 * time.Minute)
	for time.Now().Before(deadline) {
		if b, err := os.ReadFile(errPath); err == nil {
			return fmt.Errorf("提权更新失败: %s", string(b))
		}
		if _, err := os.Stat(okPath); err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("更新超时（30 分钟），请查看日志目录: %s", cfg.Dirs.Logs)
}

func restartAfterUpdate(cfg *appcfg.Config, log *logx.Logger, win *ui.UI) {
	mu.Lock()
	node := currentNode
	mu.Unlock()
	d := &dsh.Dsh{BinJS: dsh.BinPath(cfg.Dirs.Dsh), Dir: cfg.Dirs.Dsh}
	url, m, err := runDsh(cfg, node, d, log, win.SetLoading)
	if err != nil {
		log.Printf("更新后启动失败: %v", err)
		win.SetError("更新完成但服务启动失败: "+err.Error(), cfg.Dirs.Logs)
		return
	}
	mu.Lock()
	currentMgr = m
	mu.Unlock()
	win.Navigate(url)
}

func updateFailed(cfg *appcfg.Config, log *logx.Logger, win *ui.UI, err error) {
	log.Printf("%v", err)
	win.SetError(err.Error(), cfg.Dirs.Logs)
	go func() {
		time.Sleep(6 * time.Second)
		restartAfterUpdate(cfg, log, win)
	}()
}

func selfUpdateMain(args []string) int {
	var nodeExe, prefix, cache, registry, logPath string
	for i := 0; i < len(args); i++ {
		if i+1 >= len(args) {
			break
		}
		switch args[i] {
		case "--node":
			i++
			nodeExe = args[i]
		case "--prefix":
			i++
			prefix = args[i]
		case "--cache":
			i++
			cache = args[i]
		case "--registry":
			i++
			registry = args[i]
		case "--log":
			i++
			logPath = args[i]
		}
	}
	logf := func(format string, a ...any) { fmt.Printf(format+"\n", a...) }
	if logPath != "" {
		if l, err := logx.New(logPath); err == nil {
			logf = l.Printf
		}
	}
	node := &nodejs.Node{Exe: nodeExe, Bundled: true}
	_, err := dsh.Update(node, prefix, registry, cache, logf)
	if err != nil {
		logf("更新失败: %v", err)
		_ = os.WriteFile(filepath.Join(prefix, "update.err"), []byte(err.Error()), 0o644)
		return 1
	}
	logf("更新完成")
	_ = os.WriteFile(filepath.Join(prefix, "update.ok"), []byte(time.Now().Format(time.RFC3339)), 0o644)
	return 0
}
