package ipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
)

// Msgrcv calls the msgrcv() syscall.
func Msgrcv(qid uint, msg *Msgbuf, flags uint) error {
	var buf = make([]byte, uintSize+msgmax)

	lengthRead, _, errno := syscall.Syscall6(syscall.SYS_MSGRCV,
		uintptr(qid),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(msgmax),
		uintptr(msg.Mtype),
		uintptr(flags),
		0,
	)
	if errno != 0 {
		return errno
	}
	buffer := bytes.NewBuffer(buf)
	switch uintSize {
	case 4:
		var mtype uint32
		err := binary.Read(buffer, binary.LittleEndian, &mtype)
		if err != nil {
			return fmt.Errorf("Can't write binary: %v", err)
		}
		msg.Mtype = uint(mtype)
	case 8:
		var mtype uint64
		err := binary.Read(buffer, binary.LittleEndian, &mtype)
		if err != nil {
			return fmt.Errorf("Can't write binary: %v", err)
		}
		msg.Mtype = uint(mtype)
	}
	msg.Mtext = buf[uintSize : uintSize+int(lengthRead)]
	return nil
}
