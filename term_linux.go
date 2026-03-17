//go:build linux

package bezel

import (
	"fmt"
	"syscall"
	"unsafe"
)

// termios flag constants (linux).
const (
	// Input flags.
	brkint = 0x0002
	icrnl  = 0x0100
	inpck  = 0x0010
	istrip = 0x0020
	ixon   = 0x0400

	// Output flags.
	opost = 0x0001

	// Control flags.
	cs8 = 0x0030

	// Local flags.
	echo   = 0x0008
	icanon = 0x0002
	iexten = 0x8000
	isig   = 0x0001

	// Control character indices.
	vmin  = 6
	vtime = 5

	// ioctl commands.
	ioctlGetAttr = 0x5401 // TCGETS
	ioctlSetAttr = 0x5402 // TCSETS
	ioctlGetSize = 0x5413 // TIOCGWINSZ
)

// termios matches the linux kernel struct layout.
// Fields are uint32 because linux uses unsigned int (4 bytes).
type termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Line   uint8
	Cc     [32]byte
	Ispeed uint32
	Ospeed uint32
}

type winsize struct {
	Rows   uint16
	Cols   uint16
	Xpixel uint16
	Ypixel uint16
}

func getTermios(fd uintptr) (termios, error) {
	var t termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, ioctlGetAttr, uintptr(unsafe.Pointer(&t)))
	if errno != 0 {
		return t, fmt.Errorf("TCGETS: %w", errno)
	}
	return t, nil
}

func setTermios(fd uintptr, t *termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, ioctlSetAttr, uintptr(unsafe.Pointer(t)))
	if errno != 0 {
		return fmt.Errorf("TCSETS: %w", errno)
	}
	return nil
}

func getWinsize(fd uintptr) (winsize, error) {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, ioctlGetSize, uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return ws, fmt.Errorf("TIOCGWINSZ: %w", errno)
	}
	return ws, nil
}
