package single

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32DLL  = windows.NewLazySystemDLL("kernel32.dll")
	createMutexW = kernel32DLL.NewProc("CreateMutexW")
	handle       uintptr
)

func Acquire(name string) bool {
	h, _, callErr := createMutexW.Call(0, 0, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(name))))
	if h == 0 {
		return false
	}
	handle = h
	errno, ok := callErr.(syscall.Errno)
	if !ok {
		return true
	}
	return errno != syscall.Errno(windows.ERROR_ALREADY_EXISTS)
}

func Release() {
	if handle != 0 {
		windows.CloseHandle(windows.Handle(handle))
		handle = 0
	}
}
