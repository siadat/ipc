package ipc

import (
	"bytes"
	"syscall"
	"unsafe"
)

// Msgrcv calls the msgrcv() syscall.
func Msgrcv(qid uint64, msg *Msgbuf, flags uint64) error {
	qbuf := msgbufInternal{
		Mtype: msg.Mtype,
	}
	_, _, err := syscall.Syscall6(syscall.SYS_MSGRCV,
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
	sz := bytes.Index(qbuf.Mtext[:], []byte{0})
	msg.Mtype = qbuf.Mtype
	msg.Mtext = qbuf.Mtext[:sz]
	return nil
}
