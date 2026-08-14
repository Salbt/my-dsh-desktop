package ui

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32DLL        = windows.NewLazySystemDLL("user32.dll")
	kernel32DLL      = windows.NewLazySystemDLL("kernel32.dll")
	procSendMessageW = user32DLL.NewProc("SendMessageW")
	procLoadIconW    = user32DLL.NewProc("LoadIconW")
	procGetModuleW   = kernel32DLL.NewProc("GetModuleHandleW")
)

const (
	wmSetIcon = 0x0080
	iconSmall = 0
	iconBig   = 1
)

func applyWindowIcon(hwnd unsafe.Pointer) {
	if hwnd == nil {
		return
	}
	hinst, _, _ := procGetModuleW.Call(0)
	if hinst == 0 {
		return
	}
	icon, _, _ := procLoadIconW.Call(hinst, 1)
	if icon == 0 {
		return
	}
	procSendMessageW.Call(uintptr(hwnd), wmSetIcon, iconSmall, icon)
	procSendMessageW.Call(uintptr(hwnd), wmSetIcon, iconBig, icon)
}
