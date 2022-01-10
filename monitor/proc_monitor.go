package monitor

import (
	"bytes"
	"encoding/binary"
	"log"
	"lxraa/global"
	"os"
	"sync/atomic"
	"syscall"
)

/**
使用cn_proc实现进程监控
内核向user space发通知需要开启CONFIG_PROC_EVENTS=y，可以通过zcat /proc/config.gz查看是否支持
参考资料：https://www.codenong.com/6075013/
代码：https://github.com/kinvolk/nswatch
**/

const (
	CN_IDX_PROC = 0x1
	CN_VAL_PROC = 0X1

	PROC_CN_GET_FEATURES = 0
	PROC_CN_MCAST_LISTEN = 1
	PROC_CN_MCAST_IGNORE = 2

	PROC_EVENT_NONE     = 0x00000000
	PROC_EVENT_FORK     = 0x00000001
	PROC_EVENT_EXEC     = 0x00000002
	PROC_EVENT_UID      = 0x00000004
	PROC_EVENT_GID      = 0x00000040
	PROC_EVENT_SID      = 0x00000080
	PROC_EVENT_PTRACE   = 0x00000100
	PROC_EVENT_COMM     = 0x00000200
	PROC_EVENT_NS       = 0x00000400
	PROC_EVENT_COREDUMP = 0x40000000
	PROC_EVENT_EXIT     = 0x80000000
)

type cbId struct {
	Idx uint32
	Val uint32
}

type cnMsg struct {
	Id    cbId
	Seq   uint32
	Ack   uint32
	Len   uint16
	Flags uint16
}

type netlinkProcMessage struct {
	Header syscall.NlMsghdr
	Data   cnMsg
}

type Netlink struct {
	addr *syscall.SockaddrNetlink
	sock int32
	seq  uint32
}

func (nl *Netlink) Getfd() int {
	return int(atomic.LoadInt32(&nl.sock))
}
func (nl *Netlink) Getaddr() *syscall.SockaddrNetlink {
	return nl.addr
}

func (nl *Netlink) Connect() error {
	sock, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_DGRAM, syscall.NETLINK_CONNECTOR)
	if err != nil {
		return err
	}
	nl.sock = int32(sock)
	nl.addr = &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		// Pid:
		// Groups: uint32(CN_IDX_PROC),
	}
	err2 := syscall.Bind(nl.Getfd(), nl.Getaddr())
	if err2 != nil {
		syscall.Close(nl.Getfd())
		return err
	}
	return nil
}

func (nl *Netlink) Receive() {
	bufP := global.GetByteArray()
	for {
		nr, _, err := syscall.Recvfrom(nl.Getfd(), *bufP, 0)
		if err != nil {
			continue
		}
		if nr < syscall.NLMSG_HDRLEN {
			continue
		}
		// msgChannel <- (*bufP)[:nr]
		msgs, err2 := syscall.ParseNetlinkMessage((*bufP)[:nr])
		if nil != err2 {
			log.Println("解析netlink msg失败")
			continue
		}
		for _, msg := range msgs {
			log.Println(msg.Data)
		}
	}
}

/**
订阅内核消息
**/
func (nl *Netlink) send(op uint32) error {
	nl.seq++
	pr := &netlinkProcMessage{}
	plen := binary.Size(pr.Data) + binary.Size(op)
	pr.Header.Len = syscall.NLMSG_HDRLEN + uint32(plen)
	pr.Header.Type = uint16(syscall.NLMSG_DONE)
	pr.Header.Flags = 0
	pr.Header.Seq = nl.seq
	pr.Header.Pid = uint32(os.Getpid())
	pr.Data.Id.Idx = CN_IDX_PROC
	pr.Data.Len = uint16(binary.Size(op))
	buf := bytes.NewBuffer(make([]byte, 0, pr.Header.Len))
	binary.Write(buf, binary.LittleEndian, pr)
	binary.Write(buf, binary.LittleEndian, op)
	err := syscall.Sendto(nl.Getfd(), buf.Bytes(), 0, nl.addr)
	return err

}

//channel chan interface{}
func (nl *Netlink) StartProcM() {
	if nl == nil {
		log.Fatal("初始化netlink失败")
	}
	err := nl.Connect()
	if err != nil {
		log.Fatal("连接netlink失败")
	}
	nl.send(PROC_CN_MCAST_LISTEN)
}

func EndProcM() {

}
