package ipc_test

import (
	"fmt"
	"syscall"
	"testing"

	"github.com/siadat/ipc"
	"github.com/siadat/ipc/cgo_msgget"
)

func TestMsgget(t *testing.T) {
	keyFunc := func(path string, id uint) uint {
		key, err := ipc.Ftok(path, id)
		if err != nil {
			t.Fatalf("err=%q path=%q id=%d", err, path, id)
		}
		return key
	}

	cases := []struct {
		key  uint
		perm int
	}{
		{keyFunc(".", 123), 0660},
		{keyFunc("/dev/null", 123), 0660},
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
