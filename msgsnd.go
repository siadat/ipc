package ipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
)

// Msgsnd calls the msgsnd() syscall.
func Msgsnd(qid uint, msg *Msgbuf, flags uint) error {
	if len(msg.Mtext) > msgmax {
		return fmt.Errorf("mtext is too large, %d > %d", len(msg.Mtext), msgmax)
	}

	buf := make([]byte, uintSize+msgmax)
	buffer := bytes.NewBuffer(buf)
	buffer.Reset()
	var err error
	switch uintSize {
	case 4:
		err = binary.Write(buffer, binary.LittleEndian, uint32(msg.Mtype))
	case 8:
		err = binary.Write(buffer, binary.LittleEndian, uint64(msg.Mtype))
	}
	if err != nil {
		return fmt.Errorf("Can't write binary: %v", err)
	}
	buffer.Write(msg.Mtext)
	_, _, errno := syscall.Syscall6(syscall.SYS_MSGSND,
		uintptr(qid),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(msg.Mtext)),
		uintptr(flags),
		0,
		0,
	)
	if errno != 0 {
		return errno
	}
	return nil
}
