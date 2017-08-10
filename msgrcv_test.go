package ipc_test

import (
	"log"
	"syscall"
	"testing"
	"time"

	"github.com/siadat/ipc"
)

func TestMsgrcvBlocks(t *testing.T) {
	keyFunc := func(path string, id uint64) uint64 {
		key, err := ipc.Ftok(path, id)
		if err != nil {
			t.Fatal(err)
		}
		return key
	}

	cases := []struct {
		key  uint64
		perm int
	}{
		{keyFunc("/dev/null", uint64('m')), 0600},
	}

	for _, tt := range cases {
		qid, err := ipc.Msgget(tt.key, ipc.IPC_CREAT|ipc.IPC_EXCL|tt.perm)
		if err == syscall.EEXIST {
			t.Errorf("queue with key 0x%x exists", tt.key)
		}
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err = ipc.Msgctl(qid, ipc.IPC_RMID)
			if err != nil {
				t.Fatal(err)
			}
		}()

		qbuf := &ipc.Msgbuf{Mtype: 12}
		ch := make(chan struct{})
		go func() {
			err = ipc.Msgrcv(qid, qbuf, 0)
			if err == syscall.EIDRM {
				// OK, queue was removed
			} else if err != nil {
				log.Fatalf("syscall error: %v", err)
			}
			ch <- struct{}{}
		}()

		select {
		case <-time.After(100 * time.Millisecond):
			// OK, Msgrvc should block
		case <-ch:
			t.Fatal("msgrcv did not block")
		}
	}
}
