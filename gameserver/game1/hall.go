package game1

import (
	"fmt"
	"math/rand"
	"reflect"
	"rgames/common/base"
	"rgames/common/utils"
	pb "rgames/protobuf"
	"time"
)

func init() {
	methodMap = make(map[string]reflect.Method)
	reflectApis()
}

var methodMap map[string]reflect.Method

func reflectApis() {
	typ := reflect.TypeOf(&Hall{})
	num := typ.NumMethod()
	for i := 0; i < num; i++ {
		//fmt.Printf("method[%d]%s\n", i, typ.Method(i).Name)
		methodMap[typ.Method(i).Name] = typ.Method(i)
	}
}

const (
	MUL = 1000
)

type Hall struct {
	base.HallBase
	num int32

	Length int64 // x
	Width  int64 // y
	Height int64 // z
}

func Create(gameid int32) base.HallImpl {
	h := &Hall{
		Length: 100 * MUL,
		Width:  100 * MUL,
		Height: 0,
	}
	h.Init(gameid, h, CreateUser, CreateRoom)
	return h
}

func (h *Hall) HandleOtherMessage(um *base.UserMessage) {
	utils.TLog.Debug("Game1 HallBase::HandleOtherMessage")
	fmt.Printf("num=%v\n", h.num)

	utils.TLog.Debug(fmt.Sprintf("HandleMsg msg[%v]", um.Msg.String()))
	if cb, ok := methodMap[um.Msg.GetCmd()]; ok {
		var params []reflect.Value
		params = append(params, reflect.ValueOf(h))
		params = append(params, reflect.ValueOf(um))
		cb.Func.Call(params)
	} else {
		utils.TLog.Error(fmt.Sprintf("unknown cmd msg[%v]", um.Msg.String()))
	}
}

/////////////////////////////////////////////////////////////////////////////////////

func (h *Hall) EnterScene(um *base.UserMessage) {
	u := um.U
	if u == nil {
		utils.TLog.Error(fmt.Sprintf("EnterScene user not found"))
		return
	}

	mo := &MoveObj{
		Pos:        &pb.Vec3{rand.Int63n(h.Length), rand.Int63n(h.Width), 0},
		Dir:        &pb.Vec3{0, 0, 0},
		Speed:      0,
		LastUpTime: 0,
	}
	u1 := (u).(*User)
	u1.MImpl = mo
	utils.TLog.Info(fmt.Sprintf("Game1 EnterScene id:%v mo:%v", u.UserId(), mo))

	ntf := &pb.SCEnterNtf{
		Id:  u.UserId(),
		Pos: mo.Pos,
		Dir: mo.Dir,
	}

	h.BroadcastMsg(&pb.Pb{
		Cmd:      "EnterSceneNtf",
		Tserver:  time.Now().Unix(),
		EnterNtf: ntf,
	})
}

func (h *Hall) MoveStart(um *base.UserMessage) {
}

func (h *Hall) oveStop(um *base.UserMessage) {
}
