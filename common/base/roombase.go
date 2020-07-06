package base

import (
	"fmt"
	"reflect"
	"rgames/common/utils"
	pb "rgames/protobuf"
	"sync"
	"time"
	"unsafe"
)

const (
	TICKER_INTERVAL_TIME = 500 * time.Millisecond // tick间隔时间
	SPLITE_MSG_COUNT     = 100
)

func init() {
	roomMethodMap = make(map[string]reflect.Method)
	reflectRoomApis()
}

var roomMethodMap map[string]reflect.Method

func reflectRoomApis() {
	typ := reflect.TypeOf(&RoomBase{})
	num := typ.NumMethod()
	for i := 0; i < num; i++ {
		//fmt.Printf("method[%d]%s\n", i, typ.Method(i).Name)
		roomMethodMap[typ.Method(i).Name] = typ.Method(i)
	}
}

type RoomImpl interface {
	RoomId() int64
	SetInfo(roomid int64, creater int64, impl HallImpl)
	Close(reason int32)
	SendMessage(um *UserMessage, needclose bool) bool
	SetState(st int32)
	GetState() int32
	Round() int32
	CurRound() int32
	SetRound(round int32)
	HandleOtherMessage(um *UserMessage) bool
	DumpAllUsers() []*pb.UserInfo
	GameStart()
	GameEnd()

	OnUserEnter(user UserImpl, seat int32)
	OnUserLeave(user UserImpl, seat int32)
	OnBeforeUserLeave(user UserImpl)
	OnFirstUserEnter(user UserImpl)
	OnLastUserLeave(user UserImpl)
	OnUserSitDown(user UserImpl, seatno int32)
	OnBeforeUserStandUp(user UserImpl, seatno int32)
	OnUserStandUp(user UserImpl, seatno int32)
	OnUserReady(user UserImpl, seatno int32)
	OnUserCancelReady(user UserImpl, seatno int32)
	CheckTimeOut(now int64)
	OnTimeOut(now int64)
	OnGameStart()
	OnGameEnd()
	OnRoundEnd()
	Clean()

	CanEnter(user UserImpl, passwd string) int32
	CanLeave(user UserImpl, reason int32) int32
	CanSitDown(user UserImpl, seatno int32) bool
	CanStandUp(user UserImpl, seatno int32) bool
	CanReady(user UserImpl) bool
	CanCancelReady(user UserImpl) bool
	CanChat(user UserImpl) bool
	CanStartGame() bool
}

type RoomBase struct {
	Impl       RoomImpl
	Id         int64
	Creater    int64
	Passwd     string
	Count      int32
	Users      []UserImpl
	CurCount   int32
	Watchers   map[int64]UserImpl
	closed     bool
	HallPtr    HallImpl
	state      int32
	stateLock  sync.Mutex
	createTime int64
	round      int32
	curRound   int32

	// 消息通道, 外部接口调用 正常情况应该只被该游戏的hall协程调用
	msg     chan *UserMessage
	msgLock sync.Mutex
	ticker  *time.Ticker
}

func (r *RoomBase) Init(impl RoomImpl, count int32, passwd string) {
	r.Impl = impl
	r.Passwd = passwd
	r.Count = count
	r.Users = make([]UserImpl, count)
	r.CurCount = 0
	r.Watchers = make(map[int64]UserImpl)
	r.msg = make(chan *UserMessage, Max_Msg_Count)
	r.ticker = time.NewTicker(TICKER_INTERVAL_TIME)
	r.closed = false
	r.SetState(int32(pb.RoomState_Idle))
	r.createTime = time.Now().Unix()
	r.Run()
}

func (r *RoomBase) Fini() {
	r.msgLock.Lock()
	defer r.msgLock.Unlock()

	close(r.msg)
	r.ticker.Stop()

	pmsg := &pb.Pb{
		Cmd:    "Inner_RemoveRoom",
		Roomid: r.Id,
	}
	um := &UserMessage{
		Msg: pmsg,
	}
	r.HallPtr.SendMessage(um)
}

func (r *RoomBase) SetState(st int32) {
	r.stateLock.Lock()
	defer r.stateLock.Unlock()
	r.state = st
}

func (r *RoomBase) GetState() int32 {
	r.stateLock.Lock()
	defer r.stateLock.Unlock()
	return r.state
}

func (r *RoomBase) Round() int32 {
	return r.round
}

func (r *RoomBase) CurRound() int32 {
	return r.curRound
}

func (r *RoomBase) SetRound(round int32) {
	r.round = round
}

func (r *RoomBase) Close(reason int32) {
	utils.TLog.Info(fmt.Sprintf("room[%v] close[%v]", r.Id, reason))
	msg := &pb.Pb{
		Cmd: "Inner_CloseRoom",
	}
	um := &UserMessage{
		Msg: msg,
	}
	r.SendMessage(um, true)
	r.SetState(int32(pb.RoomState_Closing))
}

func (r *RoomBase) Run() {
	go func() {
		utils.TLog.Info(fmt.Sprintf("Go RoomBase[%v] Start", r.Id))
		for {
			select {
			case pbmsg := <-r.msg:
				isend := false
				func() {
					defer utils.TryCatch()
					if pbmsg != nil {
						isend = r.HandleMsg(pbmsg)
					}
				}()
				if isend {
					goto roomEnd
				}
			case <-r.ticker.C:
				func() {
					defer utils.TryCatch()
					r.OnTimer(time.Now().Unix())
				}()
			}
		}
	roomEnd:
		utils.TLog.Info(fmt.Sprintf("Go RoomBase[%v] Stop", r.Id))
		r.Fini()
	}()
}

func (r *RoomBase) SendMessage(um *UserMessage, needclose bool) bool {
	r.msgLock.Lock()
	defer r.msgLock.Unlock()

	// 已经关闭的情况下返回失败
	if r.closed {
		return false
	}
	r.closed = needclose

	select {
	case r.msg <- um:
		return true
	default:
		utils.TLog.Error("RoomBase::SendMessage")
	}
	return true
}

func (r *RoomBase) HandleMsg(um *UserMessage) bool {
	utils.TLog.Debug(fmt.Sprintf("HandleMsg msg[%v]", um.Msg.String()))
	if cb, ok := roomMethodMap[um.Msg.GetCmd()]; ok {
		var params []reflect.Value

		hb := (*RoomBase)(unsafe.Pointer(&r.Impl))
		params = append(params, reflect.ValueOf(hb))
		params = append(params, reflect.ValueOf(um))
		ret := cb.Func.Call(params)
		return ret[0].Bool()
	} else {
		return r.Impl.HandleOtherMessage(um)
	}
	return false
}

func (r *RoomBase) HandleOtherMessage(um *UserMessage) bool {
	return false
}

func (r *RoomBase) RoomId() int64 {
	return r.Id
}

func (r *RoomBase) SetInfo(roomid int64, creater int64, impl HallImpl) {
	r.Id = roomid
	r.Creater = creater
	r.HallPtr = impl
	r.curRound = 0
}

func (r *RoomBase) BroadcastMsg(pbmsg *pb.Pb) {
	for _, u := range r.Users {
		if u != nil {
			u.SendMsg(pbmsg)
		}
	}
	for _, u := range r.Watchers {
		if u != nil {
			u.SendMsg(pbmsg)
		}
	}
}

func (r *RoomBase) DumpAllUsers() []*pb.UserInfo {
	uinfos := make([]*pb.UserInfo, 0)
	for _, u := range r.Users {
		if u != nil {
			uinfos = append(uinfos, u.DumpUserInfo())
		}
	}
	for _, u := range r.Watchers {
		if u != nil {
			uinfos = append(uinfos, u.DumpUserInfo())
		}
	}
	return uinfos
}

func (r *RoomBase) GameStart() {
	r.curRound++
	r.SetState(int32(pb.RoomState_Gaming))
	for _, u := range r.Users {
		if u != nil && u.GetGameState() == int8(pb.UserGameState_Ready) {
			u.SetGameState(int8(pb.UserGameState_Playing))
		}
	}
	utils.TLog.Info(fmt.Sprintf("GameStart room[%v] curRound[%v]", r.Id, r.curRound))
	r.Impl.OnGameStart()
}

func (r *RoomBase) GameEnd() {
	r.SetState(int32(pb.RoomState_Wait))
	for _, u := range r.Users {
		if u != nil && u.GetGameState() == int8(pb.UserGameState_Playing) {
			u.SetGameState(int8(pb.UserGameState_UnReady))
		}
	}

	r.Impl.OnGameEnd()
	if r.curRound >= r.round {
		r.Close(int32(pb.RoomCloseReason_Normal))
	}
}

////////////////////////////////////////////////////////////////////////////////////////

func (r *RoomBase) Inner_CloseRoom(um *UserMessage) bool {
	r.cleanRoomUsers()

	if r.curRound > 0 {
		r.Impl.OnRoundEnd()
	}

	// 这条消息是最后一条消息
	return true
}

func (r *RoomBase) Inner_LeaveRoom(um *UserMessage) bool {
	ret := r.Impl.CanLeave(um.U, int32(pb.LeaveReason_ConnBreak))
	user := um.U
	utils.TLog.Info(fmt.Sprintf("user[%v] leave room[%v] reason[%v] ret[%v]",
		user.UserId(), r.RoomId(), um.Msg.GetLeaveRoomReq().GetReason(), ret))

	if um.Msg.GetLeaveRoomReq().GetReason() != int32(pb.LeaveReason_ConnBreak) {
		rsp := &pb.SCLeaveRoomRsp{
			Result: ret,
		}
		user.SendMsg(&pb.Pb{
			Cmd:          "LeaveRoomRsp",
			Tserver:      time.Now().Unix(),
			LeaveRoomRsp: rsp,
		})
	}

	if ret == int32(pb.ErrNo_Success) {
		r.leaveRoom(user)
	}

	if um.Msg.GetLeaveRoomReq().GetReason() == int32(pb.LeaveReason_ConnBreak) {
		// 断线需要通知hall协程清除数据
		um.Msg.Cmd = "Inner_LeaveRoom_Done"
		um.U = user
		um.Msg.Roomid = r.Id
		if ret == int32(pb.ErrNo_Success) {
			um.Msg.Roomid = 0
		}
		r.HallPtr.SendMessage(um)
	}

	if ret == int32(pb.ErrNo_Success) {
		// 房间空了要回收
		leftcount := r.totalUserCount()
		if leftcount <= 0 {
			r.Close(int32(pb.RoomCloseReason_Empty))
		}
	}

	return false
}

func (r *RoomBase) Inner_EnterRoom(um *UserMessage) bool {
	user := um.U
	ret := r.Impl.CanEnter(um.U, um.Msg.GetEnterRoomReq().GetPasswd())
	utils.TLog.Info(fmt.Sprintf("user[%v] enter room[%v] passwd[%v] ret[%v]",
		user.UserId(), r.RoomId(), um.Msg.GetEnterRoomReq().GetPasswd(), ret))
	rsp := &pb.SCEnterRoomRsp{
		Result: ret,
		Roomid: r.RoomId(),
	}
	user.SendMsg(&pb.Pb{
		Cmd:          "EnterRoomRsp",
		Tserver:      time.Now().Unix(),
		EnterRoomRsp: rsp,
	})

	if ret == int32(pb.ErrNo_Success) {
		r.enterRoom(user)
	}

	return false
}

func (r *RoomBase) Inner_BreakEnterRoom(um *UserMessage) bool {
	user := um.U
	user.OnBreakEnter(r.RoomId(), user.Seat())
	return false
}

func (r *RoomBase) Inner_Ready(um *UserMessage) bool {
	user := um.U
	seat := user.Seat()

	rsp := &pb.SCReadyRsp{
		Result: int32(pb.ErrNo_Success),
	}

	if seat >= 0 {
		u := r.Users[seat]
		if u == nil || u.UserId() != user.UserId() {
			panic("Inner_Ready:seatUser not valid")
		}

		if r.state == int32(pb.RoomState_Wait) && user.GetGameState() == int8(pb.UserGameState_UnReady) {
			r.ready(user)
		} else {
			rsp.Result = int32(pb.ErrNo_InvalidState)
		}
	} else {
		rsp.Result = int32(pb.ErrNo_NotInSeat)
	}

	user.SendMsg(&pb.Pb{
		Cmd:      "ReadyRsp",
		Tserver:  time.Now().Unix(),
		ReadyRsp: rsp,
	})
	if rsp.Result != int32(pb.ErrNo_Success) {
		return false
	}
	ntf := &pb.SCReadyNtf{
		Uinfo: user.DumpUserInfo(),
	}
	r.BroadcastMsg(&pb.Pb{
		Cmd:      "ReadyNtf",
		Tserver:  time.Now().Unix(),
		ReadyNtf: ntf,
	})
	r.checkGameStart()

	return false
}

func (r *RoomBase) Inner_CancelReady(um *UserMessage) bool {
	user := um.U
	seat := user.Seat()

	rsp := &pb.SCCancelReadyRsp{
		Result: int32(pb.ErrNo_Success),
	}

	if seat >= 0 {
		u := r.Users[seat]
		if u == nil || u.UserId() != user.UserId() {
			panic("Inner_Ready:seatUser not valid")
		}

		if r.state == int32(pb.RoomState_Wait) && user.GetGameState() == int8(pb.UserGameState_Ready) {
			r.cancelReady(user)
		} else {
			rsp.Result = int32(pb.ErrNo_InvalidState)
		}
	} else {
		rsp.Result = int32(pb.ErrNo_NotInSeat)
	}

	user.SendMsg(&pb.Pb{
		Cmd:            "CancelReadyRsp",
		Tserver:        time.Now().Unix(),
		CancelReadyRsp: rsp,
	})
	if rsp.Result != int32(pb.ErrNo_Success) {
		return false
	}
	ntf := &pb.SCCancelReadyNtf{
		Uinfo: user.DumpUserInfo(),
	}
	r.BroadcastMsg(&pb.Pb{
		Cmd:            "ReadyNtf",
		Tserver:        time.Now().Unix(),
		CancelReadyNtf: ntf,
	})

	return false
}

//////////////////////////////////////////////////////////////////////////////////////

func (r *RoomBase) OnUserEnter(user UserImpl, seat int32) {

}

func (r *RoomBase) OnUserLeave(user UserImpl, seat int32) {

}

func (r *RoomBase) OnBeforeUserLeave(user UserImpl) {

}

func (r *RoomBase) OnFirstUserEnter(user UserImpl) {

}

func (r *RoomBase) OnLastUserLeave(user UserImpl) {

}

func (r *RoomBase) OnUserSitDown(user UserImpl, seatno int32) {

}

func (r *RoomBase) OnBeforeUserStandUp(user UserImpl, seatno int32) {

}

func (r *RoomBase) OnUserStandUp(user UserImpl, seatno int32) {

}

func (r *RoomBase) OnUserReady(user UserImpl, seatno int32) {

}

func (r *RoomBase) OnUserCancelReady(user UserImpl, seatno int32) {

}

func (r *RoomBase) OnTimer(now int64) {
	r.Impl.CheckTimeOut(now)
}

func (r *RoomBase) CheckTimeOut(now int64) {
	r.checkRoomExpire(now)
	r.checkActionTimeOut(now)
	r.Impl.OnTimeOut(now)
}

func (r *RoomBase) OnTimeOut(now int64) {

}

func (r *RoomBase) OnGameStart() {
}

func (r *RoomBase) OnRoundEnd() {

}

func (r *RoomBase) Clean() {

}

func (r *RoomBase) CanEnter(user UserImpl, passwd string) int32 {
	if r.userInRoom(user) {
		return int32(pb.ErrNo_AleardyInRoom)
	}
	if r.curRound > 0 {
		return int32(pb.ErrNo_RoomInGame)
	}
	return int32(pb.ErrNo_Success)
}

func (r *RoomBase) CanLeave(user UserImpl, reason int32) int32 {
	if !r.userInRoom(user) {
		return int32(pb.ErrNo_NotInRoom)
	}
	if r.curRound > 0 {
		return int32(pb.ErrNo_RoomInGame)
	}
	return int32(pb.ErrNo_Success)
}

func (r *RoomBase) CanSitDown(user UserImpl, seatno int32) bool {
	return true
}

func (r *RoomBase) CanStandUp(user UserImpl, seatno int32) bool {
	return true
}

func (r *RoomBase) CanReady(user UserImpl) bool {
	return true
}

func (r *RoomBase) CanCancelReady(user UserImpl) bool {
	return true
}

func (r *RoomBase) CanChat(user UserImpl) bool {
	return true
}

func (r *RoomBase) CanStartGame() bool {
	if r.CurCount < r.Count {
		return false
	}
	readycount := 0
	for _, u := range r.Users {
		if u != nil && u.GetGameState() == int8(pb.UserGameState_Ready) {
			readycount++
		}
	}
	return int32(readycount) == r.CurCount
}

/////////////////////////////////////////////////////////////////////////////////////////////

func (r *RoomBase) checkRoomExpire(now int64) {
	alivetime := now - r.createTime
	if (r.curRound == 0 && alivetime >= 180) || alivetime >= 3600 {
		r.Close(int32(pb.RoomCloseReason_Expire))
	}
}

func (r *RoomBase) checkActionTimeOut(now int64) {
	for _, impl := range r.Users {
		if impl == nil {
			continue
		}
		impl.CheckActionTimeOut(now)
	}
}

func (r *RoomBase) leaveRoom(user UserImpl) {
	userid := user.UserId()
	seat := user.Seat()
	if seat < 0 {
		delete(r.Watchers, userid)
	} else {
		r.Users[seat] = nil
		r.CurCount--
		if r.CurCount < 0 {
			r.CurCount = 0
		}
	}
	user.SetSeat(-1)
	user.SetRoomId(0)
	utils.TLog.Info(fmt.Sprintf("user[%v] leaveRoom[%v] oldseat[%v] curcount[%v]",
		userid, r.RoomId(), seat, r.CurCount))

	ntf := &pb.SCLeaveRoomNtf{
		Uinfo: user.DumpUserInfo(),
	}
	r.BroadcastMsg(&pb.Pb{
		Cmd:          "LeaveRoomNtf",
		Tserver:      time.Now().Unix(),
		LeaveRoomNtf: ntf,
	})
	r.Impl.OnUserLeave(user, seat)
}

func (r *RoomBase) findSeat() int32 {
	for s, u := range r.Users {
		if u == nil {
			return int32(s)
		}
	}
	return -1
}

func (r *RoomBase) enterRoom(user UserImpl) {
	userid := user.UserId()
	seat := r.findSeat()
	user.SetSeat(seat)
	if seat >= 0 {
		r.Users[seat] = user
		r.CurCount++
		user.SetGameState(int8(pb.UserGameState_UnReady))
	} else {
		r.Watchers[user.UserId()] = user
		user.SetGameState(int8(pb.UserGameState_Watching))
	}
	r.SetState(int32(pb.RoomState_Wait))
	user.SetRoomId(r.RoomId())
	utils.TLog.Info(fmt.Sprintf("user[%v] enterRoom[%v] seat[%v] curcount[%v]",
		userid, r.RoomId(), seat, r.CurCount))

	ntf := &pb.SCEnterRoomNtf{
		Uinfo: user.DumpUserInfo(),
	}
	r.BroadcastMsg(&pb.Pb{
		Cmd:          "EnterRoomNtf",
		Tserver:      time.Now().Unix(),
		EnterRoomNtf: ntf,
	})

	// 分包发送
	alluinfos := r.DumpAllUsers()
	from := 0
	end := SPLITE_MSG_COUNT
	sz := len(alluinfos)
	for {
		if end > sz {
			end = sz
		}
		ntf := &pb.SCRoomUserInfos{
			Uinfo: alluinfos[from:end],
		}
		user.SendMsg(&pb.Pb{
			Cmd:           "RoomUserInfos",
			Tserver:       time.Now().Unix(),
			RoomUserInfos: ntf,
		})
		from = end
		end += SPLITE_MSG_COUNT
		if from >= sz {
			break
		}
	}

	r.Impl.OnUserEnter(user, seat)
}

func (r *RoomBase) userInRoom(user UserImpl) bool {
	_, ok := r.Watchers[user.UserId()]
	if ok {
		return true
	}
	for _, u := range r.Users {
		if u != nil && u.UserId() == user.UserId() {
			return true
		}
	}

	return false
}

func (r *RoomBase) ready(user UserImpl) {
	user.SetGameState(int8(pb.UserGameState_Ready))
}

func (r *RoomBase) cancelReady(user UserImpl) {
	user.SetGameState(int8(pb.UserGameState_UnReady))
}

func (r *RoomBase) checkGameStart() {
	if r.Impl.CanStartGame() {
		r.Impl.GameStart()
	}
}

func (r *RoomBase) cleanRoomUsers() {
	for _, u := range r.Users {
		if u != nil {
			u.SetRoomId(0)
			u.SetSeat(-1)
		}
	}

	for _, u := range r.Watchers {
		if u != nil {
			u.SetRoomId(0)
			u.SetSeat(-1)
		}
	}
}

func (r *RoomBase) totalUserCount() int32 {
	return int32(len(r.Watchers)) + r.CurCount
}
