package ipc_test

import (
	"log"
	"math/rand"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/siadat/ipc"
	"github.com/stretchr/testify/assert"
)

func randomGen() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func TestMsgctl(t *testing.T) {
	cases := []struct {
		path string
		id   uint32
		perm int
	}{
		{".", randomGen().Uint32(), 0600},
		{"/dev/null", randomGen().Uint32(), 0600},
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

func TestMsgctlExtend(t *testing.T) {
	tryCreateQueue := func() uint {
		limit := 1024
		for i := 0; i < limit; i++ {
			key, err := ipc.Ftok(".", uint(randomGen().Uint32()))
			if err != nil {
				continue
			}
			qid, err := ipc.Msgget(key, 0600|ipc.IPC_CREAT)
			if err != nil {
				continue
			}
			return qid
		}
		t.Fatalf("Fails to create random queue")
		return 0
	}
	deleteQueue := func(qid uint) {
		assert.NoError(t, ipc.Msgctl(qid, ipc.IPC_RMID))
	}
	type args struct {
		qid  uint
		cmd  int
		mbuf interface{}
	}
	lastMsqidDS := ipc.MsqidDS{}
	lastMsginfo := ipc.Msginfo{}
	noSetup := func(args) {}
	noTest := func(args, int) {}
	tests := []struct {
		name    string
		args    args
		setup   func(args)
		test    func(args, int)
		wantErr bool
	}{
		{
			"IPC_RMID: works",
			args{
				tryCreateQueue(),
				ipc.IPC_RMID,
				nil,
			},
			noSetup,
			func(x args, ret int) {
				if _, err := ipc.Msgget(x.qid, 0600); err != syscall.ENOENT {
					t.Fatalf("Queue still exist over %q with error %s", x.qid, err.Error())
				}
			},
			false,
		},
		{
			"IPC_STAT: works",
			args{
				tryCreateQueue(),
				ipc.IPC_STAT,
				&lastMsqidDS,
			},
			func(x args) {
				buf := &ipc.Msgbuf{}
				buf.Mtype = 1
				buf.Mtext = []byte("payload")
				for i := 0; i < 42; i++ {
					assert.NoError(t,
						ipc.Msgsnd(x.qid, buf, ipc.IPC_NOWAIT))
				}
			},
			func(x args, ret int) {
				defer deleteQueue(x.qid)
				now := time.Now().UnixNano()
				assert.Equal(t, 42*len("payload"), int(lastMsqidDS.MsgCbytes))
				assert.Equal(t, 42, int(lastMsqidDS.MsgQnum))
				assert.Less(t, 0, int(lastMsqidDS.MsgQbytes))
				assert.Less(t, int64(lastMsqidDS.MsgStime.Nanosecond()), now)
				assert.Equal(t, int64(0), int64(lastMsqidDS.MsgRtime.Nanosecond()))
				assert.Less(t, int64(lastMsqidDS.MsgCtime.Nanosecond()), now)
				assert.Equal(t, os.Getpid(), int(lastMsqidDS.MsglSpid))
				assert.Equal(t, 0, int(lastMsqidDS.MsglRpid))
				assert.Equal(t, os.Geteuid(), int(lastMsqidDS.MsgPerm.Uid))
				assert.Equal(t, os.Getegid(), int(lastMsqidDS.MsgPerm.Gid))
			},
			false,
		},
		{
			"IPC_SET: change maximum number of bytes allowed in queue",
			args{
				tryCreateQueue(),
				ipc.IPC_SET,
				&lastMsqidDS,
			},
			func(x args) {
				lastMsqidDS.MsgQbytes >>= 1
			},
			func(x args, ret int) {
				defer deleteQueue(x.qid)
				var localBuf ipc.MsqidDS
				_, err := ipc.MsgctlExtend(x.qid, ipc.IPC_STAT, &localBuf)
				assert.NoError(t, err)
				// Queue has a new size in term of allowed bytes
				assert.Equal(t, lastMsqidDS.MsgQbytes, localBuf.MsgQbytes)
			},
			false,
		},
		{
			"IPC_SET && MSG_STAT_ANY: make a queue not accessible for a user and read it anyway",
			args{
				tryCreateQueue(),
				ipc.IPC_SET,
				&lastMsqidDS,
			},
			func(x args) {
				lastMsqidDS.MsgPerm.Mode ^= lastMsqidDS.MsgPerm.Mode
			},
			func(x args, ret int) {
				defer deleteQueue(x.qid)
				var localBuf ipc.MsqidDS
				// Queue is not more accessible for current user
				_, err := ipc.MsgctlExtend(x.qid, ipc.IPC_STAT, &localBuf)
				assert.Equal(t, syscall.EACCES, err)
				// Using MSG_STAT_ANY is possible to access without check the permissions:
				// as long as the `cat /proc/sysvipc/msg` does for any users
				_, err = ipc.MsgctlExtend(x.qid, ipc.MSG_STAT_ANY, &localBuf)
				assert.NoError(t, err)
			},
			false,
		},
		{
			"IPC_SET: error, wrong type",
			args{
				0,
				ipc.IPC_SET,
				&lastMsginfo,
			},
			noSetup,
			noTest,
			true,
		},
		{
			"IPC_INFO: works",
			args{
				0,
				ipc.IPC_INFO,
				&lastMsginfo,
			},
			func(x args) {},
			func(x args, ret int) {
				assert.Less(t, 0, int(lastMsginfo.Msgpool))
				assert.Less(t, 0, int(lastMsginfo.Msgmap))
				assert.Less(t, 0, int(lastMsginfo.Msgmax))
				assert.Less(t, 0, int(lastMsginfo.Msgmnb))
				assert.Less(t, 0, int(lastMsginfo.Msgmni))
				assert.Less(t, 0, int(lastMsginfo.Msgssz))
				assert.Less(t, 0, int(lastMsginfo.Msgtql))
				assert.Less(t, 0, int(lastMsginfo.Msgseg))
			},
			false,
		},
		{
			"IPC_INFO: error, wrong type",
			args{
				0,
				ipc.IPC_INFO,
				&lastMsqidDS,
			},
			noSetup,
			noTest,
			true,
		},
		{
			"MSG_INFO: works",
			args{
				0,
				ipc.MSG_INFO,
				&lastMsginfo,
			},
			noSetup,
			func(x args, ret int) {
				assert.Less(t, 0, int(lastMsginfo.Msgpool))
				assert.LessOrEqual(t, 0, int(lastMsginfo.Msgmap))
				assert.Less(t, 0, int(lastMsginfo.Msgmax))
				assert.Less(t, 0, int(lastMsginfo.Msgmnb))
				assert.Less(t, 0, int(lastMsginfo.Msgmni))
				assert.Less(t, 0, int(lastMsginfo.Msgssz))
				assert.LessOrEqual(t, 0, int(lastMsginfo.Msgtql))
				assert.Less(t, 0, int(lastMsginfo.Msgseg))
			},
			false,
		},
		{
			"MSG_INFO: error, wrong type",
			args{
				0,
				ipc.MSG_INFO,
				&lastMsqidDS,
			},
			noSetup,
			noTest,
			true,
		},
		{
			"MSG_INFO && MSG_STAT: loop over the existing messages to query",
			args{
				0,
				ipc.MSG_INFO,
				&lastMsginfo,
			},
			noSetup,
			func(x args, ret int) {
				// see https://man7.org/tlpi/code/online/dist/svmsg/svmsg_ls.c.html
				for ret > 0 {
					ret--
					var localBuf ipc.MsqidDS
					r, err := ipc.MsgctlExtend(uint(ret), ipc.MSG_STAT, &localBuf)
					t.Logf("id = %d, ret = %d, key = %d", ret, r, localBuf.MsgPerm.Key)
					if err != nil {
						if err == syscall.EACCES || err == syscall.EINVAL {
							continue
						}
						t.Fatalf("Unexcepted error %s", err.Error())
					}
				}
			},
			false,
		},
		{
			"default case: error",
			args{
				0,
				42,
				nil,
			},
			noSetup,
			noTest,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ret int
			var err error
			tt.setup(tt.args)
			{
				ret, err = ipc.MsgctlExtend(tt.args.qid, tt.args.cmd, tt.args.mbuf)
				if err != nil != tt.wantErr {
					if err == syscall.EIDRM {
						t.Log("test [", tt.name, "] Required CAP_IPC_OWNER capability")
						return
					} else {
						t.Errorf("MsgctlExtend() error = %v, wantErr %v", err, tt.wantErr)
					}
				}
			}
			tt.test(tt.args, ret)
		})
	}
}

func ExampleMsgctlExtend_IPC_INFO_IPC_STAT_IPC_SET_IPC_RMID() {
	// fill Msginfo data structure
	var bufInfo ipc.Msginfo
	_, err := ipc.MsgctlExtend(0, ipc.IPC_INFO, &bufInfo)
	if err != nil {
		log.Fatal(err)
	}

	// create an ftok key
	key, err := ipc.Ftok("/dev/null", 20<<2)
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
		_, err := ipc.MsgctlExtend(qid, ipc.IPC_RMID, nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// send random number of message
	send := func() error {
		msg := &ipc.Msgbuf{Mtype: 12, Mtext: []byte("bonjour")}
		return ipc.Msgsnd(qid, msg, 0)
	}

	N := rand.Intn(16) + 1
	if len("bonjour") > int(bufInfo.Msgmax) ||
		N*len("bonjour") > int(bufInfo.Msgmnb) {
		log.Fatalf("Can't say 'bonjour' %d times", N)
	}
	for i := 0; i < N; i++ {
		if err := send(); err != nil {
			log.Fatal(err)
		}
	}

	// read MsqidDS data structure
	var bufStat ipc.MsqidDS
	_, err = ipc.MsgctlExtend(qid, ipc.IPC_STAT, &bufStat)
	if err != nil {
		log.Fatal(err)
	}
	if N != int(bufStat.MsgQnum) {
		log.Fatal("Mismatch number of messages currently on the message queue")
	}
	log.Printf("Sent 'bonjour' msg %d times", N)

	bufStat.MsgPerm.Mode = 0400
	// set queue in read only mode
	_, err = ipc.MsgctlExtend(qid, ipc.IPC_SET, &bufStat)
	if err != nil {
		log.Fatal(err)
	}

	// can't send data
	if err = send(); err != syscall.EACCES {
		log.Fatalf("Should be inaccessible %s", err.Error())
	}
}

func ExampleMsgctlExtend_MSG_INFO_MSG_STAT() {

	// fill Msginfo
	var bufInfo ipc.Msginfo
	ret, err := ipc.MsgctlExtend(0, ipc.MSG_INFO, &bufInfo)
	if err != nil {
		log.Fatal(err)
	}

	// loop over messages
	for ret > 0 {
		ret--
		var localBuf ipc.MsqidDS
		r, err := ipc.MsgctlExtend(uint(ret), ipc.MSG_STAT, &localBuf)
		log.Printf("id = %d, ret = %d, key = %d", ret, r, localBuf.MsgPerm.Key)
		if err != nil {
			if err == syscall.EACCES || err == syscall.EINVAL {
				continue
			}
			log.Fatalf("Unexcepted error %s", err.Error())
		}
	}
}
