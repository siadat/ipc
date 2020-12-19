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
	buffer, err := PrepareMsg(msg)
	if err != nil {
		return err
	}

	return MsgsndPrepared(qid, len(msg.Mtext), buffer, flags)
}

// PrepareMsg creates a buffer containing the message to send
func PrepareMsg(msg *Msgbuf) (*bytes.Buffer, error) {
	if len(msg.Mtext) > msgmax {
		return nil, fmt.Errorf("mtext is too large, %d > %d", len(msg.Mtext), msgmax)
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
		return nil, fmt.Errorf("Can't write binary: %v", err)
	}
	buffer.Write(msg.Mtext)

	return buffer, nil

}

//MsgsndPrepared sends a prepared message using the msgsnd() syscall
func MsgsndPrepared(qid uint, len int, buffer *bytes.Buffer, flags uint) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_MSGSND,
		uintptr(qid),
		uintptr(unsafe.Pointer(&buffer.Bytes()[0])),
		uintptr(len),
		uintptr(flags),
		0,
		0,
	)
	if errno != 0 {
		return errno
	}
	return nil

}