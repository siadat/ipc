package ipc

import (
	"fmt"
	"log"
	"syscall"
	"testing"
	"time"
)

func TestMsgsnd(t *testing.T) {
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
		{keyFunc("/dev/null", uint64('m')), 0600},
	}

	for _, tt := range cases {
		qid, err := Msgget(tt.key, IPC_CREAT|IPC_EXCL|tt.perm)
		if err == syscall.EEXIST {
			t.Errorf("queue(key=0x%x) exists", tt.key)
		}
		if err != nil {
			t.Fatal(err)
		}
		defer func(qid uint64) {
			err := Msgctl(qid, IPC_RMID)
			if err != nil {
				t.Fatal(err)
			}
		}(qid)

		mtext := "hello"
		done := make(chan struct{})
		go func() {
			qbuf := &Msgbuf{Mtype: 12}
			err := Msgrcv(qid, qbuf, 0)
			if err != nil {
				t.Fatal(err)
			}
			if want, got := mtext, string(qbuf.Mtext); want != got {
				t.Fatalf("want %#v, got %#v", want, got)
			}
			done <- struct{}{}
		}()

		m := &Msgbuf{Mtype: 12, Mtext: []byte(mtext)}
		err = Msgsnd(qid, m, 0)
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
	key, err := Ftok("/dev/null", 42)
	if err != nil {
		panic(err)
	}

	// create a new message queue
	qid, err := Msgget(key, IPC_CREAT|IPC_EXCL|0600)
	if err == syscall.EEXIST {
		log.Fatalf("queue(key=0x%x) exists", key)
	}
	if err != nil {
		log.Fatal(err)
	}

	// remove queue in the end
	defer func() {
		err := Msgctl(qid, IPC_RMID)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// send a message
	go func() {
		msg := &Msgbuf{Mtype: 12, Mtext: []byte("bonjour")}
		err = Msgsnd(qid, msg, 0)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// receive the message
	msg := &Msgbuf{Mtype: 12}
	err = Msgrcv(qid, msg, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("received message: %q", msg.Mtext)

	// Output:
	// received message: "bonjour"
}
