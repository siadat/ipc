package main

import (
	"fmt"
	"ipc"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	nargs := len(os.Args)
	ftokpath := "/dev/null"
	ftokid := uint64(109)
	mtype := uint64(42)
	perm := 0600

	if nargs < 2 {
		usage()
		return
	}

	key, err := ipc.Ftok(ftokpath, ftokid)
	if err != nil {
		log.Fatal(err)
	}

	var qid uint64

	switch os.Args[1] {
	case "s", "r":
		qid, err = ipc.Msgget(key, ipc.IPC_CREAT|perm)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("ftok path %s\n", ftokpath)
	fmt.Printf("ftok perm 0%o\n", perm)
	fmt.Printf("ftok id   %d\n", ftokid)
	fmt.Printf("key       0x%x\n", key)
	fmt.Printf("qid       %d\n", qid)
	fmt.Printf("mtype     %d\n", mtype)

	switch os.Args[1] {
	case "s":
		if nargs < 3 {
			usage()
			return
		}

		m := &ipc.Msgbuf{Mtype: mtype, Mtext: []byte(os.Args[2])}
		log.Printf("Sending a message: %s\n", string(m.Mtext[:]))
		err := ipc.Msgsnd(qid, m, 0)
		if err != nil {
			log.Fatal("msgsnd2", err)
		}
	case "r":
		fmt.Printf("Waiting for a message...\n")

		m := &ipc.Msgbuf{Mtype: mtype}
		err := ipc.Msgrcv(qid, m, 0)
		if err != nil {
			log.Fatal("msgrcv2", err)
		}
		log.Printf("Received a message: %s\n", string(m.Mtext))
	case "rm":
		if nargs < 3 {
			usage()
			return
		}

		if strings.HasPrefix(os.Args[2], "0x") {
			os.Args[2] = os.Args[2][2:]
		}

		key, err := strconv.ParseUint(os.Args[2], 16, 64)
		if err != nil {
			log.Fatal(err)
		}

		qid, err := ipc.Msgget(key, perm)
		if err != nil {
			log.Fatal(err)
		}

		err = ipc.Msgctl(qid, ipc.IPC_RMID)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func usage() {
	fmt.Println("Usage: ipc COMMAND [ARGS...]")
	fmt.Println("       ipc r")
	fmt.Println("       ipc s <message>")
	fmt.Println("       ipc rm <hex-key>")
}
