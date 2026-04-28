//go:build !windows

package output

import (
	"syscall"
	"unsafe"
)

func getTermSize() (int, int, error) {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}
	var ws winsize
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, 1, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	if err != 0 {
		return 80, 24, nil
	}
	return int(ws.Col), int(ws.Row), nil
}
