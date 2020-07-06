package antnet

import (
	"net"
	"strings"
	"sync"
	"time"
)

var DefMsgQueTimeout int = 180

type MsgType int

type NetType int

const (
	NetTypeTcp NetType = iota //TCP类型
	NetTypeUdp                //UDP类型
	NetTypeWs                 //websocket
)

type ConnType int

const (
	ConnTypeListen ConnType = iota //监听
	ConnTypeConn                   //连接产生的
	ConnTypeAccept                 //Accept产生的
)

type IMsgQue interface {
	Id() uint32
	GetConnType() ConnType
	GetNetType() NetType

	LocalAddr() string
	RemoteAddr() string
	SetRealRemoteAddr(addr string)

	Stop()
	IsStop() bool
	Available() bool
	IsProxy() bool

	Send(m *Message) (re bool)
	SendString(str string) (re bool)
	SendStringLn(str string) (re bool)
	SendByteStr(str []byte) (re bool)
	SendByteStrLn(str []byte) (re bool)
	SetTimeout(t int)
	SetCmdReadRaw()
	GetTimeout() int
	Reconnect(t int) //重连间隔  最小1s，此函数仅能连接关闭是调用

	GetHandler() IMsgHandler

	SetUser(user interface{})
	GetUser() interface{}

	SetGroupId(group string)
	DelGroupId(group string)
	ClearGroupId(group string)
	IsInGroup(group string) bool
}

type msgQue struct {
	id uint32 //唯一标示

	cwrite  chan *Message //写入通道
	stop    int32         //停止标记
	connTyp ConnType      //通道类型

	handler  IMsgHandler //处理者
	timeout  int         //传输超时
	lastTick int64

	init           bool
	available      bool
	multiplex      bool
	group          map[string]int
	user           interface{}
	callbackLock   sync.Mutex
	gmsgId         uint16
	realRemoteAddr string //当使用代理是，需要特殊设置客户端真实IP
}

func (r *msgQue) SetUser(user interface{}) {
	r.user = user
}

func (r *msgQue) getGMsg(add bool) *gMsg {
	if add {
		r.gmsgId++
	}
	gm := gmsgArray[r.gmsgId]
	return gm
}
func (r *msgQue) SetCmdReadRaw() {

}
func (r *msgQue) Available() bool {
	return r.available
}

func (r *msgQue) GetUser() interface{} {
	return r.user
}

func (r *msgQue) GetHandler() IMsgHandler {
	return r.handler
}

func (r *msgQue) GetConnType() ConnType {
	return r.connTyp
}

func (r *msgQue) Id() uint32 {
	return r.id
}

func (r *msgQue) SetTimeout(t int) {
	if t >= 0 {
		r.timeout = t
	}
}

func (r *msgQue) isTimeout(tick *time.Timer) bool {
	left := int(Timestamp - r.lastTick)
	if left < r.timeout || r.timeout == 0 {
		if r.timeout == 0 {
			tick.Reset(time.Second * time.Duration(DefMsgQueTimeout))
		} else {
			tick.Reset(time.Second * time.Duration(r.timeout-left))
		}
		return false
	}
	LogInfo("msgque close because timeout id:%v wait:%v timeout:%v", r.id, left, r.timeout)
	return true
}

func (r *msgQue) GetTimeout() int {
	return r.timeout
}

func (r *msgQue) Reconnect(t int) {

}

func (r *msgQue) IsProxy() bool {
	return r.realRemoteAddr != ""
}

func (r *msgQue) SetRealRemoteAddr(addr string) {
	r.realRemoteAddr = addr
}

func (r *msgQue) SetGroupId(group string) {
	r.callbackLock.Lock()
	if r.group == nil {
		r.group = make(map[string]int)
	}
	r.group[group] = 0
	r.callbackLock.Unlock()
}

func (r *msgQue) DelGroupId(group string) {
	r.callbackLock.Lock()
	if r.group != nil {
		delete(r.group, group)
	}
	r.callbackLock.Unlock()
}

func (r *msgQue) ClearGroupId(group string) {
	r.callbackLock.Lock()
	r.group = nil
	r.callbackLock.Unlock()
}

func (r *msgQue) IsInGroup(group string) bool {
	re := false
	r.callbackLock.Lock()
	if r.group != nil {
		_, re = r.group[group]
	}
	r.callbackLock.Unlock()
	return re
}

func (r *msgQue) Send(m *Message) (re bool) {
	if m == nil {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			re = false
		}
	}()
	select {
	case r.cwrite <- m:
	default:
		LogWarn("msgque write channel full msgque:%v", r.id)
		r.cwrite <- m
	}

	return true
}

func (r *msgQue) SendString(str string) (re bool) {
	return r.Send(&Message{Data: []byte(str)})
}

func (r *msgQue) SendStringLn(str string) (re bool) {
	return r.SendString(str + "\n")
}

func (r *msgQue) SendByteStr(str []byte) (re bool) {
	return r.SendString(string(str))
}

func (r *msgQue) SendByteStrLn(str []byte) (re bool) {
	return r.SendString(string(str) + "\n")
}

func (r *msgQue) baseStop() {
	if r.cwrite != nil {
		close(r.cwrite)
	}

	msgqueMapSync.Lock()
	delete(msgqueMap, r.id)
	msgqueMapSync.Unlock()
	LogInfo("msgque close id:%d", r.id)
}

func (r *msgQue) processMsg(msgque IMsgQue, msg *Message) bool {
	if r.handler != nil {
		r.handler.OnProcessMsg(msgque, msg)
	}
	return true
}

type IMsgHandler interface {
	OnNewMsgQue(msgque IMsgQue) bool                //新的消息队列
	OnDelMsgQue(msgque IMsgQue)                     //消息队列关闭
	OnProcessMsg(msgque IMsgQue, msg *Message) bool //默认的消息处理函数
	OnConnectComplete(msgque IMsgQue, ok bool) bool //连接成功
}

type DefMsgHandler struct {
}

func (r *DefMsgHandler) OnNewMsgQue(msgque IMsgQue) bool                { return true }
func (r *DefMsgHandler) OnDelMsgQue(msgque IMsgQue)                     {}
func (r *DefMsgHandler) OnProcessMsg(msgque IMsgQue, msg *Message) bool { return true }
func (r *DefMsgHandler) OnConnectComplete(msgque IMsgQue, ok bool) bool { return true }

func StartServer(addr string, handler IMsgHandler) error {
	addrs := strings.Split(addr, "://")
	if addrs[0] == "tcp" || addrs[0] == "all" {
		listen, err := net.Listen("tcp", addrs[1])
		if err == nil {
			msgque := newTcpListen(listen, handler, addr)
			Go(func() {
				LogDebug("process listen for tcp msgque:%d", msgque.id)
				msgque.listen()
				LogDebug("process listen end for tcp msgque:%d", msgque.id)
			})
		} else {
			LogError("listen on %s failed, errstr:%s", addr, err)
			return err
		}
	}
	if addrs[0] == "udp" || addrs[0] == "all" {
		naddr, err := net.ResolveUDPAddr("udp", addrs[1])
		if err != nil {
			LogError("listen on %s failed, errstr:%s", addr, err)
			return err
		}
		conn, err := net.ListenUDP("udp", naddr)
		if err == nil {
			msgque := newUdpListen(conn, handler, addr)
			Go(func() {
				LogDebug("process listen for udp msgque:%d", msgque.id)
				msgque.listen()
				LogDebug("process listen end for udp msgque:%d", msgque.id)
			})
		} else {
			LogError("listen on %s failed, errstr:%s", addr, err)
			return err
		}
	}
	if addrs[0] == "ws" || addrs[0] == "wss" {
		naddr := strings.SplitN(addrs[1], "/", 2)
		url := "/"
		if len(naddr) > 1 {
			url = "/" + naddr[1]
		}
		if addrs[0] == "wss" {
			Config.EnableWss = true
		}
		msgque := newWsListen(naddr[0], url, handler)
		Go(func() {
			LogDebug("process listen for ws msgque:%d", msgque.id)
			msgque.listen()
			LogDebug("process listen end for ws msgque:%d", msgque.id)
		})
	}
	return nil
}

func StartConnect(netType string, addr string, handler IMsgHandler, user interface{}) IMsgQue {
	var msgque IMsgQue
	if netType == "ws" || netType == "wss" {
		msgque = newWsConn(addr, nil, handler, user)
	} else {
		msgque = newTcpConn(netType, addr, nil, handler, user)
	}
	if handler.OnNewMsgQue(msgque) {
		msgque.Reconnect(0)
		return msgque
	} else {
		msgque.Stop()
	}
	return nil
}
