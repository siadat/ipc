package ipc

import (
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

// Permission define the Linux permission for a specific queue,
// private data is required to read/write buffer of constant length
type Permission struct {
	Key     uint32
	Uid     uint32
	Gid     uint32
	Cuid    uint32
	Cgid    uint32
	Mode    uint16
	pad1    uint16
	Seq     uint16
	pad2    uint16
	unused1 uint64
	unused2 uint64
}

// MsqidDS struct type define one of the buffer for MsgctlExtend
type MsqidDS struct {
	MsgPerm   Permission
	MsgStime  time.Time
	MsgRtime  time.Time
	MsgCtime  time.Time
	MsgCbytes uint64
	MsgQnum   uint64
	MsgQbytes uint64
	MsglSpid  uint32
	MsglRpid  uint32
}

// Msginfo struct type define one of the buffer for MsgctlExtend
type Msginfo struct {
	Msgpool int32
	Msgmap  int32
	Msgmax  int32
	Msgmnb  int32
	Msgmni  int32
	Msgssz  int32
	Msgtql  int32
	Msgseg  uint16
}

type ipctime uint64

// msqidds internal data for buffer conversion
type msqidds struct {
	msgPerm   Permission
	msgStime  ipctime
	msgRtime  ipctime
	msgCtime  ipctime
	msgCbytes uint64
	msgQnum   uint64
	msgQbytes uint64
	msglSpid  uint32
	msglRpid  uint32
	unused    [8]uint8
}

// convertBuffer from private to public data conversion
func (m *msqidds) convertBuffer(r *MsqidDS) {
	r.MsgPerm = m.msgPerm
	r.MsgStime = time.Unix(int64(0), int64(m.msgStime))
	r.MsgRtime = time.Unix(int64(0), int64(m.msgRtime))
	r.MsgCtime = time.Unix(int64(0), int64(m.msgCtime))
	r.MsgCbytes = m.msgCbytes
	r.MsgQnum = m.msgQnum
	r.MsgQbytes = m.msgQbytes
	r.MsglSpid = m.msglSpid
	r.MsglRpid = m.msglRpid
}

// convertBuffer from public to private data conversion
func (m *MsqidDS) convertBuffer(r *msqidds) {
	r.msgPerm = m.MsgPerm
	r.msgStime = ipctime(m.MsgStime.Nanosecond())
	r.msgRtime = ipctime(m.MsgRtime.Nanosecond())
	r.msgCtime = ipctime(m.MsgCtime.Nanosecond())
	r.msgCbytes = m.MsgCbytes
	r.msgQnum = m.MsgQnum
	r.msgQbytes = m.MsgQbytes
	r.msglSpid = m.MsglSpid
	r.msglRpid = m.MsglRpid
}

// Msgctl calls the msgctl() syscall.
// FIXME: we are not passing the buf argument, see msgctl(2).
func Msgctl(qid uint, cmd int) error {
	if cmd != IPC_RMID {
		return fmt.Errorf("only IPC_RMID is supported at the moment")
	}
	var buf uintptr = 0
	_, _, err := syscall.Syscall(syscall.SYS_MSGCTL, uintptr(qid), buf, 0)
	if err != 0 {
		return err
	}
	return nil
}

// MsgctlExtend calls the msgctl() syscall including all behaviour.
// MSG_STAT and MSG_STAT_ANY returns the identifier of the queue whose index was given in msqid.
func MsgctlExtend(qid uint, cmd int, buf interface{}) (int, error) {
	var ret uintptr
	var errno syscall.Errno
	switch cmd {
	case IPC_RMID:
		return 0, Msgctl(qid, cmd)
	case IPC_SET, IPC_STAT, MSG_STAT, MSG_STAT_ANY:
		convertedBuffer, ok := buf.(*MsqidDS)
		if !ok {
			return 0, errors.New("Wrong buffer type: required *MsqidDS type")
		}
		var realBuffer msqidds
		if cmd == IPC_SET {
			convertedBuffer.convertBuffer(&realBuffer)
		}
		ret, _, errno = syscall.Syscall(
			syscall.SYS_MSGCTL,
			uintptr(qid),
			uintptr(cmd),
			uintptr(unsafe.Pointer(&realBuffer)))
		if errno != 0 {
			return 0, syscall.Errno(errno)
		}
		realBuffer.convertBuffer(buf.(*MsqidDS))
	case IPC_INFO, MSG_INFO:
		_, ok := buf.(*Msginfo)
		if !ok {
			return 0, errors.New("Wrong buffer type: required *Msginfo type")
		}
		ret, _, errno = syscall.Syscall(
			syscall.SYS_MSGCTL,
			0,
			uintptr(cmd),
			uintptr(unsafe.Pointer(buf.(*Msginfo))))
		if errno != 0 {
			return 0, syscall.Errno(errno)
		}
	default:
		return 0, fmt.Errorf("cmd [%d] is not a valid control command", cmd)
	}
	return int(ret), nil
}
