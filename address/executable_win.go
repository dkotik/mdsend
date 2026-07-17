//go:build windows

package address

import (
	"golang.org/x/sys/windows"
)

func isFileExecutable(path string) (ok bool, bool error) {
	utf16Path, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}

	var binaryType uint32
	// GetBinaryTypeW returns true if the file is an executable
	err = windows.GetBinaryType(utf16Path, &binaryType)
	return err == nil, err
}
