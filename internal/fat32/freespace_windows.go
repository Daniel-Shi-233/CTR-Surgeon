//go:build windows

package fat32

import (
	"syscall"
	"unsafe"
)

func getFreeSpace(path string) (uint64, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable uint64
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	ret, _, err := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		0,
		0,
	)
	if ret == 0 {
		return 0, err
	}

	return freeBytesAvailable, nil
}
