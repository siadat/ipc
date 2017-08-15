package ipc_test

import (
	"fmt"
	"log"
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

func ExampleMsgsnd() {
	// create an ftok key
	key, err := ipc.Ftok("/dev/null", 42)
	if err != nil {
		panic(err)
	}

	// create a new message queue
	qid, err := ipc.Msgget(key, ipc.IPC_CREAT|ipc.IPC_EXCL|0600)
	if err == syscall.EEXIST {
		log.Fatalf("queue(key=0x%x) exists", key)
	}
	if err != nil {
		log.Fatal(err)
	}

	// remove queue in the end
	defer func() {
		err := ipc.Msgctl(qid, ipc.IPC_RMID)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// send a message
	go func() {
		msg := &ipc.Msgbuf{Mtype: 12, Mtext: []byte("bonjour")}
		err = ipc.Msgsnd(qid, msg, 0)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// receive the message
	msg := &ipc.Msgbuf{Mtype: 12}
	err = ipc.Msgrcv(qid, msg, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("received message: %q", msg.Mtext)

	// Output:
	// received message: "bonjour"
}
