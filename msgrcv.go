package ipc

import (
	"syscall"
	"unsafe"
)

// Msgrcv calls the msgrcv() syscall.
func Msgrcv(qid uint64, msg *Msgbuf, flags uint64) error {
	qbuf := msgbufInternal{
		Mtype: msg.Mtype,
	}
	lengthRead, _, err := syscall.Syscall6(syscall.SYS_MSGRCV,
		uintptr(qid),
		uintptr(unsafe.Pointer(&qbuf)),
		uintptr(bufSize),
		uintptr(msg.Mtype),
		uintptr(flags),
		0,
	)
	if err != 0 {
		return err
	}

	msg.Mtype = qbuf.Mtype
	msg.Mtext = qbuf.Mtext[:lengthRead]
	return nil
}
