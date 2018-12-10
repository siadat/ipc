package ipc_test

import (
	"bytes"
	"fmt"
	"log"
	"syscall"
	"testing"
	"time"

	"github.com/siadat/ipc"
)

func TestMsgrcv(t *testing.T) {

	mykey, err := ipc.Ftok("/dev/null", 42)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate key: %s\n", err))
	} else {
		fmt.Printf("Generate key %d\n", mykey)
	}

	qid, err := ipc.Msgget(mykey, ipc.IPC_CREAT|0600)
	if err != nil {
		panic(fmt.Sprintf("Failed to create ipc key %d: %s\n", mykey, err))
	} else {
		fmt.Printf("Create ipc queue id %d\n", qid)
	}

	defer func() {
		err = ipc.Msgctl(qid, ipc.IPC_RMID)
		if err != nil {
			t.Fatal(err)
		}
	}()

	input := []byte{0x18, 0x2d, 0x44, 0x00, 0xfb, 0x21, 0x09, 0x40}

	msg := ipc.Msgbuf{Mtype: 12, Mtext: input}
	err = ipc.Msgsnd(qid, &msg, 0)
	if err != nil {
		panic(fmt.Sprintf("Failed to send message to ipc id %d: %s\n", qid, err))
	} else {
		fmt.Printf("Message %v send to ipc id %d\n", input, qid)
	}

	qbuf := &ipc.Msgbuf{Mtype: 12}

	err = ipc.Msgrcv(qid, qbuf, 0)

	if err != nil {
		panic(fmt.Sprintf("Failed to receive message to ipc id %d: %s\n", qid, err))
	} else {
		fmt.Printf("Message %v receive to ipc id %d\n", qbuf.Mtext, qid)
	}

	if !bytes.Equal(input, qbuf.Mtext) {
		t.Errorf("Input = %v, want %v", qbuf.Mtext, input)
	}

}

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
