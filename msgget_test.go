package ipc_test

import (
	"fmt"
	"syscall"
	"testing"

	"github.com/siadat/ipc"
	"github.com/siadat/ipc/cgo_msgget"
)

func TestMsgget(t *testing.T) {
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
		{keyFunc("msgget.go", 123), 0600},
		{keyFunc("ftok.go", 123), 0600},
	}

	for _, tt := range cases {
		_, err := ipc.Msgget(tt.key, tt.perm)
		if want, got := syscall.ENOENT, err; want != got {
			t.Fatalf("msgget for non-existing queue should fail without IPC_CREAT, want %q, got %v", want, got)
		}
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
		qidAgain, err := cgo_msgget.Msgget(tt.key, tt.perm)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := fmt.Sprintf("0x%x", qid), fmt.Sprintf("0x%x", qidAgain); got != want {
			t.Errorf("want %s, got %s", want, got)
		}
	}
}
