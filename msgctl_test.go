package ipc_test

import (
	"math/rand"
	"syscall"
	"testing"
	"time"

	"github.com/siadat/ipc"
	"github.com/siadat/ipc/cgo_ftok"
)

func TestMsgctl(t *testing.T) {
	randomGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	cases := []struct {
		path string
		id   uint64
		perm int
	}{
		{"msgget.go", randomGen.Uint64(), 0600},
		{"ftok.go", randomGen.Uint64(), 0600},
	}

	for _, tt := range cases {
		key, err := cgo_ftok.Ftok(tt.path, tt.id)
		if err != nil {
			t.Fatal(err)
		}

		qid, err := ipc.Msgget(key, tt.perm|ipc.IPC_CREAT)
		if err == syscall.EEXIST {
			t.Errorf("queue(key=0x%x) exists", key)
		}
		if err != nil {
			t.Fatal(err)
		}

		err = ipc.Msgctl(uint64(qid), ipc.IPC_RMID)
		if err != nil {
			t.Fatal(err)
		}

		_, err = ipc.Msgget(key, tt.perm)
		if want, got := syscall.ENOENT, err; want != got {
			t.Fatalf("want %q, got %v", want, got)
		}
	}
}
