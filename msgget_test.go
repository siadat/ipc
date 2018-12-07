package ipc

import (
	"fmt"
	"syscall"
	"testing"
)

func TestMsgget(t *testing.T) {
	keyFunc := func(path string, id uint64) uint64 {
		key, err := Ftok(path, id)
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
		_, err := Msgget(tt.key, tt.perm)
		if want, got := syscall.ENOENT, err; want != got {
			t.Fatalf("msgget for non-existing queue should fail without IPC_CREAT, want %q, got %v", want, got)
		}
		qid, err := Msgget(tt.key, IPC_CREAT|IPC_EXCL|tt.perm)
		if err == syscall.EEXIST {
			t.Errorf("queue with key 0x%x exists", tt.key)
		}
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err = Msgctl(qid, IPC_RMID)
			if err != nil {
				t.Fatal(err)
			}
		}()
		qidAgain, err := Msgget(tt.key, tt.perm)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := fmt.Sprintf("0x%x", qid), fmt.Sprintf("0x%x", qidAgain); got != want {
			t.Errorf("want %s, got %s", want, got)
		}
	}
}
