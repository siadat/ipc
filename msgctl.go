package ipc

import (
	"syscall"
)

// Msgctl calls the msgctl() syscall.
// FIXME: we are not passing the buf argument, see msgctl(2).
func Msgctl(qid uint64, cmd int) error {
	var buf uintptr = 0
	_, _, err := syscall.Syscall(syscall.SYS_MSGCTL, uintptr(qid), buf, 0)
	if err != 0 {
		return err
	}
	return nil
}
