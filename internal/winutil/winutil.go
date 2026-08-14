package winutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32DLL       = windows.NewLazySystemDLL("user32.dll")
	shell32DLL      = windows.NewLazySystemDLL("shell32.dll")
	procMessageBoxW = user32DLL.NewProc("MessageBoxW")
	procShellExecW  = shell32DLL.NewProc("ShellExecuteW")
)

const (
	MB_OK       = 0x00000000
	MB_YESNO    = 0x00000004
	MB_ICONINFO = 0x00000040
	IDYES       = 6
)

func MessageBox(hwnd uintptr, text, title string, flags uint32) uint32 {
	t, _ := windows.UTF16PtrFromString(text)
	ti, _ := windows.UTF16PtrFromString(title)
	r, _, _ := procMessageBoxW.Call(hwnd, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(ti)), uintptr(flags))
	return uint32(r)
}

func IsDirWritable(dir string) bool {
	f, err := os.CreateTemp(dir, ".wtest-*")
	if err != nil {
		return false
	}
	name := f.Name()
	f.Close()
	_ = os.Remove(name)
	return true
}

func RunElevated(exe string, args []string) error {
	exeW, err := windows.UTF16PtrFromString(exe)
	if err != nil {
		return err
	}
	paramsW, err := windows.UTF16PtrFromString(strings.Join(args, " "))
	if err != nil {
		return err
	}
	verbW, _ := windows.UTF16PtrFromString("runas")
	dirW, _ := windows.UTF16PtrFromString(filepath.Dir(exe))
	r, _, callErr := procShellExecW.Call(
		0,
		uintptr(unsafe.Pointer(verbW)),
		uintptr(unsafe.Pointer(exeW)),
		uintptr(unsafe.Pointer(paramsW)),
		uintptr(unsafe.Pointer(dirW)),
		1,
	)
	if r <= 32 {
		return fmt.Errorf("ShellExecuteW runas 失败: %v", callErr)
	}
	return nil
}
