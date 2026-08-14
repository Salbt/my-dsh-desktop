package ui

import (
	"html"
	"os/exec"
	"sync"

	"github.com/webview/webview_go"
)

type UI struct {
	w   webview.WebView
	mu  sync.Mutex
	run bool
}

func New(title string, width, height int) *UI {
	w := webview.New(false)
	w.SetTitle(title)
	w.SetSize(width, height, webview.HintNone)
	applyWindowIcon(w.Window())
	return &UI{w: w}
}

func (u *UI) Run() {
	u.mu.Lock()
	u.run = true
	u.mu.Unlock()
	u.w.Run()
	u.w.Destroy()
}

func (u *UI) Terminate() {
	u.w.Terminate()
}

func (u *UI) post(f func()) {
	u.mu.Lock()
	running := u.run
	u.mu.Unlock()
	if running {
		u.w.Dispatch(f)
		return
	}
	f()
}

func (u *UI) SetLoading(msg string) {
	u.post(func() { u.w.SetHtml(loadingPage(html.EscapeString(msg))) })
}

func (u *UI) SetError(msg, logsDir string) {
	u.post(func() { u.w.SetHtml(errorPage(html.EscapeString(msg), html.EscapeString(logsDir))) })
}

func (u *UI) Navigate(url string) {
	u.post(func() { u.w.Navigate(url) })
}

func OpenLogsDir(dir string) {
	_ = exec.Command("explorer", dir).Start()
}
