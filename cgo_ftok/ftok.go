package cgo_ftok

/*
#include <stdio.h>
#include <stdlib.h>
#include <sys/ipc.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

func Ftok(path string, id uint64) (uint64, error) {
	cs := C.CString(path)

	if cs == nil {
		return 0, errors.New("malloc failed to allocate memory")
	}

	defer C.free(unsafe.Pointer(cs))

	res, err := C.ftok(cs, C.int(id))

	if err != nil {
		return 0, err
	}

	return uint64(res), nil
}
