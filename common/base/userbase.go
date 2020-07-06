package base

import (
	pb "rgames/protobuf"
	"sync"
)

func init() {

}

const (
	Mask_online         = 0x1
	Mask_offline        = 0xFFFFFFFE
	Mask_Trusteeship    = 0x2
	Mask_UnTrusteeship  = 0xFFFFFFFD
	Mask_GameState      = 0xC
	Mask_ClearGameState = 0xFFFFFFF3
)

type UserImpl interface {
	ConnId() int64
	UserId() int64
	RoomId() int64
	SetRoomId(roomid int64)
	Seat() int32
	SetSeat(seat int32)
	SetSession(session *Session)
	SetOnline(online bool)
	SetTrusteeship(trusteeship bool)
	SetGameState(state int8)
	IsOnline() bool
	IsTrusteeship() bool
	GetGameState() int8
	SendMsg(pbmsg *pb.Pb)
	DumpUserInfo() *pb.UserInfo
	OnEnterHall()
	OnLeaveHall()
	OnEnterRoom(roomid int64)
	OnLeaveRoom(roomid int64)
	OnSitDown(roomid int64, seatno int32)
	OnStandUp(roomid int64, seatno int32)
	OnBreakEnter(roomid int64, seatno int32)
	CheckActionTimeOut(now int64)
	OnActionTimeOut(action int32)
}

type UserBase struct {
	Impl     UserImpl
	connId   int64
	session  *Session
	userId   int64
	Nickname string
	Header   string
	Items    []int64
	// state bit[0]=不在线/在线 bit[1]=不托管/托管 bit[2-3]=不参与游戏/未准备/已准备/正在游戏中
	state   uint32
	actions map[int32]int64
	seat    int32

	roomId int64
	idLock sync.Mutex
}

func (u *UserBase) Init(impl UserImpl, userid int64, connid int64) {
	u.userId = userid
	u.connId = connid
	u.seat = -1
	u.Impl = impl
	u.roomId = 0
	u.actions = make(map[int32]int64)
	u.session = nil
}

func (u *UserBase) ConnId() int64 {
	return u.connId
}

func (u *UserBase) UserId() int64 {
	return u.userId
}

func (u *UserBase) RoomId() int64 {
	u.idLock.Lock()
	defer u.idLock.Unlock()
	return u.roomId
}

func (u *UserBase) SetRoomId(roomid int64) {
	u.idLock.Lock()
	defer u.idLock.Unlock()
	u.roomId = roomid
}

func (u *UserBase) Seat() int32 {
	return u.seat
}

func (u *UserBase) SetSeat(seat int32) {
	u.seat = seat
}

func (u *UserBase) SetSession(session *Session) {
	u.session = session
}

func (u *UserBase) SetOnline(online bool) {
	if online {
		u.state = u.state | Mask_online
	} else {
		u.state = u.state & Mask_offline
	}
}

func (u *UserBase) SetTrusteeship(trusteeship bool) {
	if trusteeship {
		u.state = u.state | Mask_Trusteeship
	} else {
		u.state = u.state & Mask_UnTrusteeship
	}
}

func (u *UserBase) SetGameState(state int8) {
	state = state << 2
	u.state = u.state & Mask_ClearGameState
	u.state = u.state | uint32(state)
}

func (u *UserBase) IsOnline() bool {
	return (u.state & Mask_online) > 0
}

func (u *UserBase) IsTrusteeship() bool {
	return (u.state & Mask_Trusteeship) > 0
}

func (u *UserBase) GetGameState() int8 {
	return int8((u.state & Mask_GameState) >> 2)
}

func (u *UserBase) SendMsg(pbmsg *pb.Pb) {
	if pbmsg != nil && u.session != nil {
		u.session.Send(pbmsg)
	}
}

func (u *UserBase) DumpUserInfo() *pb.UserInfo {
	uinfo := &pb.UserInfo{
		Userid:   u.UserId(),
		Nickname: u.Nickname,
		Header:   u.Header,
		Seat:     u.Seat(),
		State:    u.state,
		Jdata:    "",
	}
	return uinfo
}

func (u *UserBase) OnEnterHall() {
}

func (u *UserBase) OnLeaveHall() {
}

func (u *UserBase) OnEnterRoom(roomid int64) {
}

func (u *UserBase) OnLeaveRoom(roomid int64) {
}

func (u *UserBase) OnSitDown(roomid int64, seatno int32) {
}

func (u *UserBase) OnStandUp(roomid int64, seatno int32) {
}

func (u *UserBase) OnBreakEnter(roomid int64, seatno int32) {
}

func (u *UserBase) CheckActionTimeOut(now int64) {
	for action, expire := range u.actions {
		if now >= expire {
			// 先删后调用回调，可能会出现在回调用重新加入此action的下次expire
			delete(u.actions, action)
			u.Impl.OnActionTimeOut(action)
		}
	}
}

func (u *UserBase) OnActionTimeOut(action int32) {
}
