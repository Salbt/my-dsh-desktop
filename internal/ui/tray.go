package ui

import (
	_ "embed"

	"fyne.io/systray"
)

//go:embed assets/tray.png
var trayIcon []byte

func (u *UI) RunTray(onLogs func(), onUpdate func(), onQuit func()) {
	systray.Run(func() {
		systray.SetTitle("DSH Desktop")
		systray.SetTooltip("DeepSeek Harness Desktop")
		systray.SetIcon(trayIcon)
		mLogs := systray.AddMenuItem("打开日志目录", "")
		mUpdate := systray.AddMenuItem("更新 Harness", "")
		mQuit := systray.AddMenuItem("退出", "")
		go func() {
			for {
				select {
				case <-mLogs.ClickedCh:
					onLogs()
				case <-mUpdate.ClickedCh:
					onUpdate()
				case <-mQuit.ClickedCh:
					onQuit()
				}
			}
		}()
	}, func() {})
}

func QuitTray() {
	systray.Quit()
}
