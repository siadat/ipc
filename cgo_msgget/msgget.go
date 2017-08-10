package cgo_msgget

/*
#cgo CFLAGS: -O2
#include <stdlib.h>
#include <sys/types.h>
#include <sys/ipc.h>
#include <sys/msg.h>
#include <fcntl.h>
#include <errno.h>
#include <string.h>
#include <sys/msg.h>
*/
import "C"

const (
	IPC_CREAT = C.IPC_CREAT
)

func Msgget(key uint64, mode int) (uint64, error) {
	res, err := C.msgget(C.key_t(key), C.int(mode))

	if err != nil {
		return 0, err
	}

	return uint64(res), nil
}
