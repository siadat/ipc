package ipc

import (
	"syscall"
	"unsafe"
)

func init() {
	var buf msginfo
	_, _, err := syscall.Syscall(
		syscall.SYS_MSGCTL, 
		0, 
		uintptr(IPC_INFO), 
		uintptr(unsafe.Pointer(&buf)))
	if err != 0 {
		return // can't read current msg info, leave defaults
	}

	msgmax = int(buf.msgmax)
}

type msginfo struct {
    msgpool int32
    msgmap int32
    msgmax int32
    msgmnb int32
    msgmni int32
    msgssz int32
    msgtql int32
    msgseg uint16
}