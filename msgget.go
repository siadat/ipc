package ipc

import (
	"syscall"
)

// Msgget calls the msgget() syscall.
// Calling Msgget() with the same key twice will generate a different qid,
// whenever a queue is created
func Msgget(key uint, mode int) (uint, error) {
	qid, _, err := syscall.Syscall(syscall.SYS_MSGGET, uintptr(key), uintptr(mode), 0)
	if err != 0 {
		return 0, err
	}
	return uint(qid), nil
}
