//go:build darwin

package bezel

import (
	"fmt"
	"syscall"
	"unsafe"
)

// termios flag constants (darwin).
const (
	// Input flags.
	brkint = 0x00000002
	icrnl  = 0x00000100
	inpck  = 0x00000010
	istrip = 0x00000020
	ixon   = 0x00000200

	// Output flags.
	opost = 0x00000001

	// Control flags.
	cs8 = 0x00000300

	// Local flags.
	echo   = 0x00000008
	icanon = 0x00000100
	iexten = 0x00000400
	isig   = 0x00000002

	// Control character indices.
	vmin  = 16
	vtime = 17

	// ioctl commands.
	ioctlGetAttr = 0x40487413 // TIOCGETA
	ioctlSetAttr = 0x80487414 // TIOCSETA
	ioctlGetSize = 0x40087468 // TIOCGWINSZ
)

// termios matches the darwin kernel struct layout.
// Fields are uint64 because darwin uses unsigned long (8 bytes on LP64).
type termios struct {
	Iflag  uint64
	Oflag  uint64
	Cflag  uint64
	Lflag  uint64
	Cc     [20]byte
	Ispeed uint64
	Ospeed uint64
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
		return t, fmt.Errorf("TIOCGETA: %w", errno)
	}
	return t, nil
}

func setTermios(fd uintptr, t *termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, ioctlSetAttr, uintptr(unsafe.Pointer(t)))
	if errno != 0 {
		return fmt.Errorf("TIOCSETA: %w", errno)
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
