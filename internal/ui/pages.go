package ui

func loadingPage(msg string) string {
	return `<!doctype html><html><head><meta charset="utf-8"><style>
body{margin:0;height:100vh;display:flex;align-items:center;justify-content:center;background:#0f1115;color:#c9d1d9;font-family:"Segoe UI","Microsoft YaHei",sans-serif}
.box{text-align:center}
h1{font-size:20px;font-weight:600;color:#e6edf3;margin:0 0 16px}
.dots span{display:inline-block;width:8px;height:8px;margin:0 4px;border-radius:50%;background:#4d6bfe;animation:b 1.2s infinite}
.dots span:nth-child(2){animation-delay:.2s}
.dots span:nth-child(3){animation-delay:.4s}
@keyframes b{0%,80%,100%{opacity:.2}40%{opacity:1}}
p{font-size:13px;color:#8b949e;margin:18px 0 0}
</style></head><body><div class="box"><h1>DeepSeek Harness</h1><div class="dots"><span></span><span></span><span></span></div><p>` + msg + `</p></div></body></html>`
}

func errorPage(msg, logsDir string) string {
	return `<!doctype html><html><head><meta charset="utf-8"><style>
body{margin:0;height:100vh;display:flex;align-items:center;justify-content:center;background:#0f1115;color:#c9d1d9;font-family:"Segoe UI","Microsoft YaHei",sans-serif}
.box{max-width:640px;text-align:center}
h1{font-size:20px;font-weight:600;color:#f85149;margin:0 0 16px}
p{font-size:13px;color:#c9d1d9;margin:12px 0;line-height:1.6}
.hint{font-size:12px;color:#8b949e}
</style></head><body><div class="box"><h1>启动失败</h1><p>` + msg + `</p><p class="hint">日志目录: ` + logsDir + `</p><p class="hint">关闭本窗口后所有相关进程将自动结束，重新运行程序即可重试。</p></div></body></html>`
}
