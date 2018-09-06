package ipc

import (
	"syscall"
)

// Ftok returns a probably-unique key that can be used by System V IPC
// syscalls, e.g. msgget().
// See ftok(3) and https://code.woboq.org/userspace/glibc/sysvipc/ftok.c.html
func Ftok(path string, id uint64) (uint64, error) {
	st := &syscall.Stat_t{}
	if err := syscall.Stat(path, st); err != nil {
		return 0, err
	}
	return uint64((st.Ino & 0xffff) | uint64((st.Dev & 0xff) << 16) | 
		((id & 0xff) << 24)), nil
}
