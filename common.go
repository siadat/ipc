package ipc

const (
	// https://code.woboq.org/userspace/glibc/sysdeps/unix/sysv/linux/bits/ipc.h.html
	// Mode bits for `msgget', `semget', and `shmget'.
	IPC_CREAT  = 01000 // Create key if key does not exist.
	IPC_EXCL   = 02000 // Fail if key exists.
	IPC_NOWAIT = 04000 // Return error on wait.

	// Control commands for `msgctl', `semctl', and `shmctl'.
	IPC_RMID = 0 // Remove identifier.
	IPC_SET  = 1 // Set `ipc_perm' options.
	IPC_STAT = 2 // Get `ipc_perm' options.
	IPC_INFO = 3 // See ipcs.

	// Special key values.
	IPC_PRIVATE = 0 // Private key. NOTE: this value is of type __key_t, i.e., ((__key_t) 0)
)

const (
	bufSize = 128
)

type Msgbuf struct {
	Mtype uint64
	Mtext []byte
}

type msgbufInternal struct {
	Mtype uint64
	Mtext [bufSize]byte
}
