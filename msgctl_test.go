package ipc_test

import (
	"math/rand"
	"syscall"
	"testing"
	"time"

	"github.com/siadat/ipc"
)

func TestMsgctl(t *testing.T) {
	randomGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	cases := []struct {
		path string
		id   uint32
		perm int
	}{
		{".", randomGen.Uint32(), 0600},
		{"/dev/null", randomGen.Uint32(), 0600},
	}

	for _, tt := range cases {
		key, err := ipc.Ftok(tt.path, uint(tt.id))
		if err != nil {
			t.Fatalf("err=%q path=%q id=%d", err, tt.path, tt.id)
		}

		qid, err := ipc.Msgget(key, tt.perm|ipc.IPC_CREAT)
		if err == syscall.EEXIST {
			t.Errorf("queue(key=0x%x) exists", key)
		}
		if err != nil {
			t.Fatal(err)
		}

		err = ipc.Msgctl(uint(qid), ipc.IPC_RMID)
		if err != nil {
			t.Fatal(err)
		}

		_, err = ipc.Msgget(key, tt.perm)
		if want, got := syscall.ENOENT, err; want != got {
			t.Fatalf("want %q, got %v", want, got)
		}
	}
}
