package nodejs

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type Logf func(format string, args ...any)

type Node struct {
	Exe     string
	Bundled bool
}

func (n *Node) NpmCliJS() string {
	return filepath.Join(filepath.Dir(n.Exe), "node_modules", "npm", "bin", "npm-cli.js")
}

func Ensure(dir, fallbackDir string, logf Logf) (*Node, error) {
	if exe, ok := findBundled(dir); ok {
		logf("使用内置 Node: %s", exe)
		return &Node{Exe: exe, Bundled: true}, nil
	}
	if n, ok := DetectSystem(logf); ok {
		logf("使用系统 Node: %s", n.Exe)
		return n, nil
	}
	logf("下载便携版 Node.js 24...")
	if n, err := ensureBundled(dir, logf); err == nil {
		return n, nil
	} else {
		logf("下载到 %s 失败: %v，尝试回退目录 %s", dir, err, fallbackDir)
	}
	return ensureBundled(fallbackDir, logf)
}

func findBundled(dir string) (string, bool) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}
	for _, e := range ents {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "node-") || !strings.HasSuffix(e.Name(), "-win-x64") {
			continue
		}
		exe := filepath.Join(dir, e.Name(), "node.exe")
		if st, err := os.Stat(exe); err == nil && !st.IsDir() {
			return exe, true
		}
	}
	return "", false
}

func DetectSystem(logf Logf) (*Node, bool) {
	cmd := exec.Command("node", "-v")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000200}
	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	v := strings.TrimSpace(string(out))
	if !SupportedVersion(v) {
		logf("系统 Node %s 不满足要求 (需 ^22.19 或 >=24)", v)
		return nil, false
	}
	p, err := exec.LookPath("node")
	if err != nil {
		return nil, false
	}
	return &Node{Exe: p}, true
}

func SupportedVersion(v string) bool {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return false
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	if major >= 24 {
		return true
	}
	return major == 22 && minor >= 19
}

type distEntry struct {
	Version string   `json:"version"`
	Files   []string `json:"files"`
}

func ensureBundled(dir string, logf Logf) (*Node, error) {
	body, err := fetchAll("https://nodejs.org/dist/index.json")
	if err != nil {
		logf("获取 nodejs.org 索引失败: %v，改用 npmmirror 镜像", err)
		body, err = fetchAll("https://npmmirror.com/mirrors/node/index.json")
		if err != nil {
			return nil, fmt.Errorf("获取 Node 版本索引失败: %w", err)
		}
	}
	var list []distEntry
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("解析 Node 版本索引失败: %w", err)
	}
	var ver, zipName string
	for _, e := range list {
		if !strings.HasPrefix(e.Version, "v24.") {
			continue
		}
		name := "node-" + e.Version + "-win-x64.zip"
		for _, f := range e.Files {
			if f == "win-x64-zip" {
				ver = e.Version
				zipName = name
				break
			}
		}
		if ver != "" {
			break
		}
	}
	if ver == "" {
		return nil, fmt.Errorf("未找到可用的 Node 24 win-x64 发行包")
	}
	exe := filepath.Join(dir, "node-"+ver+"-win-x64", "node.exe")
	if _, err := os.Stat(exe); err == nil {
		logf("便携 Node 已存在: %s", exe)
		return &Node{Exe: exe, Bundled: true}, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	zipPath := filepath.Join(dir, zipName)
	urls := []string{
		"https://nodejs.org/dist/" + ver + "/" + zipName,
		"https://npmmirror.com/mirrors/node/" + ver + "/" + zipName,
	}
	var lastErr error
	for _, u := range urls {
		_ = os.Remove(zipPath)
		if err := downloadFile(u, zipPath, logf); err != nil {
			lastErr = err
			logf("下载失败: %s: %v", u, err)
			continue
		}
		lastErr = nil
		break
	}
	if lastErr != nil {
		return nil, fmt.Errorf("Node 下载失败: %w", lastErr)
	}
	logf("解压 Node.js...")
	if err := unzip(zipPath, dir); err != nil {
		return nil, fmt.Errorf("解压失败: %w", err)
	}
	_ = os.Remove(zipPath)
	if _, err := os.Stat(exe); err != nil {
		return nil, fmt.Errorf("解压后未找到 %s", exe)
	}
	logf("便携 Node 就绪: %s", exe)
	return &Node{Exe: exe, Bundled: true}, nil
}

func fetchAll(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "my-dsh-desktop")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func downloadFile(url, dest string, logf Logf) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "my-dsh-desktop")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	buf := make([]byte, 1<<20)
	var total, lastMB int64
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
			total += int64(n)
			if mb := total >> 20; mb > lastMB {
				lastMB = mb
				logf("下载 %s: %d MB", filepath.Base(url), mb)
			}
		}
		if rerr == io.EOF {
			return nil
		}
		if rerr != nil {
			return rerr
		}
	}
}

func unzip(src, destDir string) error {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		target := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, werr := io.Copy(out, rc)
		rc.Close()
		out.Close()
		if werr != nil {
			return werr
		}
	}
	return nil
}
