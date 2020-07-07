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

func init() {
	methodMap = make(map[string]reflect.Method)
	reflectApis()
}

const (
	Max_Msg_Count = 5120
)

var methodMap map[string]reflect.Method

func reflectApis() {
	typ := reflect.TypeOf(&HallBase{})
	num := typ.NumMethod()
	for i := 0; i < num; i++ {
		//fmt.Printf("method[%d]%s\n", i, typ.Method(i).Name)
		methodMap[typ.Method(i).Name] = typ.Method(i)
	}
}

type UserMessage struct {
	S   *Session
	U   UserImpl
	Msg *pb.Pb
}

type HallImpl interface {
	Fini()
	GameID() int32
	SendMessage(um *UserMessage)
	HandleOtherMessage(um *UserMessage)
	Login(impl HallImpl, um *UserMessage)
	CreateRoom(impl HallImpl, um *UserMessage)
	CheckValid(account string, passwd string) int64
}

type HallBase struct {
	Impl   HallImpl
	GameId int32

	IdFrom int64

	// 消息通道, 外部接口调用
	msg     chan *UserMessage
	msgLock sync.Mutex
	abort   chan struct{}

	// 玩家管理器
	umgr           *UserMgr
	createUserFunc func(userid int64, connid int64) UserImpl

	// 房间管理器
	rmgr           *RoomMgr
	createRoomFunc func(count int32, passwd string) RoomImpl
}

func (h *HallBase) Init(gameid int32, impl HallImpl, cfu func(userid int64, connid int64) UserImpl, cfr func(count int32, passwd string) RoomImpl) {
	h.GameId = gameid
	h.IdFrom = 1
	h.Impl = impl
	h.msg = make(chan *UserMessage, Max_Msg_Count)
	h.abort = make(chan struct{})
	h.umgr = CreateUserMgr()
	h.rmgr = CreateRoomMgr()
	h.createUserFunc = cfu
	h.createRoomFunc = cfr
	h.Run()
}

func (h *HallBase) Fini() {
	h.msgLock.Lock()
	defer h.msgLock.Unlock()
	close(h.msg)
	close(h.abort)
}

func (h *HallBase) GameID() int32 {
	return h.GameId
}

func (h *HallBase) Run() {
	go func() {
		utils.TLog.Info(fmt.Sprintf("Hall[%v] Start", h.GameId))
		for {
			select {
			case pbmsg := <-h.msg:
				func() {
					defer utils.TryCatch()
					if pbmsg != nil {
						h.HandleMsg(pbmsg)
					}
				}()
			case <-h.abort:
				goto hallEnd
			}
		}
	hallEnd:
		utils.TLog.Info(fmt.Sprintf("Hall[%v] Stop", h.GameId))
	}()
}

func (h *HallBase) SendMessage(um *UserMessage) {
	h.msgLock.Lock()
	defer h.msgLock.Unlock()

	select {
	case h.msg <- um:
		return
	default:
		utils.TLog.Error("HallBase::SendMessage")
	}
}

func (h *HallBase) HandleMsg(um *UserMessage) {
	utils.TLog.Debug(fmt.Sprintf("HandleMsg msg[%v]", um.Msg.String()))
	if cb, ok := methodMap[um.Msg.GetCmd()]; ok {
		if um.S != nil {
			u := h.umgr.Find(um.S.ConnId())
			um.U = u
		}

		var params []reflect.Value

		// 因为反射接口都是 HallBase 的
		hb := (*HallBase)(unsafe.Pointer(&h.Impl))
		//params = append(params, reflect.ValueOf(h.Impl))
		params = append(params, reflect.ValueOf(hb))
		params = append(params, reflect.ValueOf(h.Impl))
		params = append(params, reflect.ValueOf(um))
		cb.Func.Call(params)
	} else {
		h.Impl.HandleOtherMessage(um)
	}
}

func (h *HallBase) HandleOtherMessage(um *UserMessage) {
	utils.TLog.Debug("HallBase::HandleOtherMessage")
}

func (h *HallBase) CheckValid(account string, passwd string) int64 {
	// TODO check valid

	// test accounts
	testAccounts := make(map[string]int64)
	testAccounts["test001"] = 10001
	testAccounts["test002"] = 10002
	testAccounts["test003"] = 10003
	testAccounts["test004"] = 10004
	testAccounts["test005"] = 10005
	testAccounts["test006"] = 10006

	if uid, ok := testAccounts[account]; ok {
		return uid
	}
	return -1
}

//////////////////////////////////////////////////////////////////////////////////////////

func (h *HallBase) Inner_RemoveRoom(impl HallImpl, um *UserMessage) {
	roomid := um.Msg.GetRoomid()
	utils.TLog.Info(fmt.Sprintf("Inner_RemoveRoom[%v]", roomid))
	utils.GetRoomIdMgr().Release(int(roomid))
	h.rmgr.Del(roomid)
}

func (h *HallBase) Inner_LeaveRoom_Done(impl HallImpl, um *UserMessage) {
	connid := um.U.ConnId()
	userid := um.U.UserId()
	utils.TLog.Debug(fmt.Sprintf("Inner_LeaveRoom_Done connid[%v] userid[%v] roomid[%v]", connid, userid, um.Msg.Roomid))
	// 要区分能否离开房间
	if um.Msg.Roomid == 0 {
		h.umgr.Del(userid)
	} else {
		h.umgr.DelIdx(connid)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////

func (h *HallBase) KeepAlive(impl HallImpl, um *UserMessage) {
}

func (h *HallBase) Login(impl HallImpl, um *UserMessage) {
	connid := um.S.ConnId()
	utils.TLog.Info(fmt.Sprintf("account[%v] conn[%v] login", um.Msg.GetLoginReq().GetAccount(), connid))
	u := um.U
	if u != nil {
		utils.TLog.Error(fmt.Sprintf("account[%v] aleardy login", um.Msg.GetLoginReq().GetAccount()))
		return
	}

	userid := h.Impl.CheckValid(um.Msg.GetLoginReq().GetAccount(), um.Msg.GetLoginReq().GetPasswd())
	if userid < 0 {
		utils.TLog.Error(fmt.Sprintf("account[%v] error", um.Msg.GetLoginReq().GetAccount()))
		return
	}

	rsp := &pb.SCLoginRsp{
		Result: int32(pb.ErrNo_Success),
		Gameid: h.GameId,
		Jdata:  "welcome",
	}

	user := h.umgr.FindByUserid(userid)
	if user != nil {
		// break enter
		rsp.Result = int32(pb.ErrNo_BreakEnter)
		h.umgr.AddIdx(connid, userid)
	} else {
		// normal login
		user = h.createUserFunc(userid, connid)
		if user == nil {
			utils.TLog.Error(fmt.Sprintf("create user account[%v][%v] aleardy login", um.Msg.GetLoginReq().GetAccount(), userid))
			return
		}
		h.umgr.Add(user)
	}

	user.SetSession(um.S)
	GetHallMgr().AddConn2Game(connid, h.GameId)

	um.S.Send(&pb.Pb{
		Cmd:      "LoginRsp",
		Tserver:  time.Now().Unix(),
		LoginRsp: rsp,
	})

	if rsp.Result == int32(pb.ErrNo_BreakEnter) {
		h.tryBreakEnter(user, um)
	}
}

func (h *HallBase) Disconnect(impl HallImpl, um *UserMessage) {
	u := um.U
	if u == nil {
		utils.TLog.Error(fmt.Sprintf("Disconnect user not found"))
		return
	}
	connid := u.ConnId()
	GetHallMgr().DelConn2Game(connid)

	roomid := u.RoomId()
	sw := 0
	if roomid > 0 {
		// 尝试通知room玩家离开房间
		r := h.rmgr.Find(roomid)
		if r != nil {
			um.Msg.Cmd = "Inner_LeaveRoom"
			leaveReq := &pb.CSLeaveRoomReq{
				Reason: int32(pb.LeaveReason_ConnBreak),
			}
			um.Msg.LeaveRoomReq = leaveReq
			um.U = u
			if r.SendMessage(um, false) {
				return
			} else {
				sw = 2
			}
		} else {
			sw = 1
		}
	}

	h.umgr.Del(u.UserId())
	utils.TLog.Error(fmt.Sprintf("Disconnect user [%v] sw[%v]", u.UserId(), sw))
}

func (h *HallBase) CreateRoom(impl HallImpl, um *UserMessage) {
	u := um.U
	if u == nil {
		utils.TLog.Error(fmt.Sprintf("CreateRoom user not found"))
		return
	}

	roomid := u.RoomId()
	if roomid > 0 {
		utils.TLog.Error(fmt.Sprintf("CreateRoom user[%v] alerady in room[%v]", u.UserId(), roomid))
		return
	}

	roomid = int64(utils.GetRoomIdMgr().Generate())
	if roomid == 0 {
		utils.TLog.Error(fmt.Sprintf("CreateRoom user[%v] no empty room", u.UserId()))
		return
	}

	r := h.createRoomFunc(um.Msg.GetCreateRoomReq().GetCount(), um.Msg.GetCreateRoomReq().GetPasswd())
	if r == nil {
		utils.TLog.Error(fmt.Sprintf("CreateRoom user[%v] faild", u.UserId()))
		return
	}
	r.SetInfo(roomid, u.UserId(), impl)
	round := um.Msg.GetCreateRoomReq().GetRound()
	if round <= 0 {
		round = 1
	}
	r.SetRound(round)
	utils.TLog.Info(fmt.Sprintf("CreateRoom user[%v] roomid[%v] round[%v]", u.UserId(), roomid, round))
	h.rmgr.Add(r)

	rsp := &pb.SCCreateRoomRsp{
		Result: int32(pb.ErrNo_Success),
		Roomid: roomid,
		Jdata:  "",
	}
	u.SendMsg(&pb.Pb{
		Cmd:           "CreateRoomRsp",
		Tserver:       time.Now().Unix(),
		CreateRoomRsp: rsp,
	})
}

func (h *HallBase) EnterRoom(impl HallImpl, um *UserMessage) {
	u := um.U
	if u == nil {
		utils.TLog.Error(fmt.Sprintf("EnterRoom user not found"))
		return
	}

	roomid := u.RoomId()
	if roomid > 0 {
		utils.TLog.Error(fmt.Sprintf("EnterRoom user[%v] aleardy in room[%v]", u.UserId(), roomid))
		return
	}

	roomid = um.Msg.GetEnterRoomReq().GetRoomid()
	r := h.rmgr.Find(roomid)
	if r == nil {
		utils.TLog.Error(fmt.Sprintf("EnterRoom user[%v] not find room[%v]", u.UserId(), roomid))
		return
	}

	// 先设置roomid，在转交给room协程, 防止相同请求
	u.SetRoomId(roomid)
	um.Msg.Cmd = "Inner_EnterRoom"
	if !r.SendMessage(um, false) {
		utils.TLog.Error(fmt.Sprintf("EnterRoom user[%v] room[%v] sendmsg failed", u.UserId(), roomid))
		u.SetRoomId(0)
	}
}

func (h *HallBase) LeaveRoom(impl HallImpl, um *UserMessage) {
	h.commonCheckRoomMessage(impl, um, "Inner_LeaveRoom")
}

func (h *HallBase) Ready(impl HallImpl, um *UserMessage) {
	h.commonCheckRoomMessage(impl, um, "Inner_Ready")
}

func (h *HallBase) CancelReady(impl HallImpl, um *UserMessage) {
	h.commonCheckRoomMessage(impl, um, "Inner_CancelReady")
}

////////////////////////////////////////////////////////////////////////////////////////////

func (h *HallBase) tryBreakEnter(impl UserImpl, um *UserMessage) {
	roomid := impl.RoomId()
	if roomid <= 0 {
		utils.TLog.Warn(fmt.Sprintf("tryBreakEnter user[%v] not in room", impl.UserId()))
		return
	}

	r := h.rmgr.Find(roomid)
	um.Msg.Cmd = "Inner_BreakEnterRoom"
	um.U = impl
	if r == nil || !r.SendMessage(um, false) {
		impl.SetRoomId(0)
		utils.TLog.Warn(fmt.Sprintf("tryBreakEnter user[%v] room[%v] state wrong", impl.UserId(), roomid))
		return
	}
}

func (h *HallBase) commonCheckRoomMessage(impl HallImpl, um *UserMessage, cmd string) {
	u := um.U
	if u == nil {
		utils.TLog.Error(fmt.Sprintf("commonCheckRoomMessage user not found cmd[%v]", cmd))
		return
	}

	roomid := u.RoomId()
	if roomid <= 0 {
		utils.TLog.Error(fmt.Sprintf("commonCheckRoomMessage user[%v] not in room[%v] cmd[%v]", u.UserId(), roomid, cmd))
		return
	}

	r := h.rmgr.Find(roomid)
	if r == nil {
		utils.TLog.Error(fmt.Sprintf("commonCheckRoomMessage user[%v] not find room[%v] cmd[%v]", u.UserId(), roomid, cmd))
		return
	}

	um.Msg.Cmd = cmd
	if !r.SendMessage(um, false) {
		utils.TLog.Error(fmt.Sprintf("commonCheckRoomMessage user[%v] room[%v] cmd[%v] sendmsg failed", u.UserId(), roomid, cmd))
	}
}
