package ipc_test

import (
	"syscall"
	"testing"
	"time"

	"github.com/siadat/ipc"
)

func TestMsgsnd(t *testing.T) {
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
			t.Errorf("queue(key=0x%x) exists", tt.key)
		}
		if err != nil {
			t.Fatal(err)
		}
		defer func(qid uint64) {
			err := ipc.Msgctl(qid, ipc.IPC_RMID)
			if err != nil {
				t.Fatal(err)
			}
		}(qid)

		mtext := "hello"
		done := make(chan struct{})
		go func() {
			qbuf := &ipc.Msgbuf{Mtype: 12}
			err := ipc.Msgrcv(qid, qbuf, 0)
			if err != nil {
				t.Fatal(err)
			}
			if want, got := mtext, string(qbuf.Mtext); want != got {
				t.Fatalf("want %#v, got %#v", want, got)
			}
			done <- struct{}{}
		}()

		m := &ipc.Msgbuf{Mtype: 12, Mtext: []byte(mtext)}
		err = ipc.Msgsnd(qid, m, 0)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("blocked for too long")
		}
	}
}
