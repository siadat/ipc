package ipc

func init() {
	var buf Msginfo
	_, err := MsgctlExtend(0, IPC_INFO, &buf)
	if nil != err {
		return
	}
	msgmax = int(buf.Msgmax)
}
