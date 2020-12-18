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
	keyFunc := func(path string, id uint) uint {
		key, err := ipc.Ftok(path, id)
		if err != nil {
			t.Fatal(err)
		}
		return key
	}

	cases := []struct {
		key  uint
		perm int
	}{
		{keyFunc("/dev/null", uint('m')), 0600},
	}

	for _, tt := range cases {
		qid, err := ipc.Msgget(tt.key, ipc.IPC_CREAT|ipc.IPC_EXCL|tt.perm)
		if err == syscall.EEXIST {
			t.Errorf("queue(key=0x%x) exists", tt.key)
		}
		if err != nil {
			t.Fatal(err)
		}
		defer func(qid uint) {
			err := ipc.Msgctl(qid, ipc.IPC_RMID)
			if err != nil {
				t.Fatal(err)
			}
		}(qid)
		mtext := "hello"
		done := make(chan struct{})
		errChan := make(chan error, 1)
		go func() {
			qbuf := &ipc.Msgbuf{Mtype: 12}
			err := ipc.Msgrcv(qid, qbuf, 0)
			if err != nil {
				errChan <- err
			}
			if want, got := mtext, string(qbuf.Mtext); want != got {
				errChan <- fmt.Errorf("want %#v, got %#v", want, got)
			}
			fmt.Printf("Received: %s\n", string(qbuf.Mtext))
			done <- struct{}{}
		}()

		m := &ipc.Msgbuf{Mtype: 12, Mtext: []byte(mtext)}
		err = ipc.Msgsnd(qid, m, 0)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-done:
		case err := <-errChan:
			t.Fatal(err)
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

func TestPrepareMsg(t *testing.T) {
	cases := []struct {
		label   string
		mtext   []byte
		wantErr bool
	}{
		{"message too long", make([]byte, ipc.Msgmax()+1), true},
		{"ok", []byte("test text"), false},
	}

	for _, tt := range cases {
		msg := ipc.Msgbuf{
			Mtype: 1234,
			Mtext: tt.mtext,
		}
		ipcMsg, err := ipc.PrepareMsg(&msg)

		if tt.wantErr {
			if ipcMsg != nil {
				t.Errorf("case %s: Expected ipcMsg to be <nil>", tt.label)
			}
			if err == nil {
				t.Errorf("case %s: Expected an error", tt.label)
			}
		} else {
			if ipcMsg == nil {
				t.Errorf("case %s: Expected an ipc message", tt.label)
			}
			if err != nil {
				t.Errorf("case %s: Expected no error, got %v", tt.label, err)
			}

		}
	}

}
