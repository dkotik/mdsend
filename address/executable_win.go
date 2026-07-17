//go:build windows

package address

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func isFileExecutable(path string) (ok bool, bool error) {
	utf16Path, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getBinaryType := kernel32.NewProc("GetBinaryTypeW")
	var binaryType uint32
	// GetBinaryTypeW returns true if the file is an executable
	// err = windows.GetBinaryType(utf16Path, &binaryType)
	_, _, err = getBinaryType.Call(
		uintptr(unsafe.Pointer(utf16Path)),
		uintptr(unsafe.Pointer(&binaryType)),
	)
	return err == nil, err
}
