package main

import (
	"fmt"
	"reflect"
	"rgames/common/antnet"
	"rgames/common/base"
	"rgames/gameserver/ddz"
	"rgames/gameserver/hzmj"
	//"unsafe"

	//"rgames/github.com/golang/protobuf/proto"
	apis "rgames/common/apis"
	"rgames/common/utils"
	pb "rgames/protobuf"
)

var gamerMap = map[string]antnet.IMsgQue{}

func C2SHandlerFunc(msgque antnet.IMsgQue, msg *antnet.Message) bool {
	ppb := &pb.Pb{}
	antnet.PBUnPack(msg.Data, ppb)

	antnet.LogInfo("%v", ppb.LoginReq)
	antnet.LogInfo("%v", ppb.LoginRsp)

	if ppb.LoginReq != nil {

	}
	return true
}

type MyMsgHandler struct {
}

func (r *MyMsgHandler) OnNewMsgQue(msgque antnet.IMsgQue) bool { return true }
func (r *MyMsgHandler) OnDelMsgQue(m antnet.IMsgQue) {
	gameid := base.GetHallMgr().GetConn2Game(int64(m.Id()))
	pbmsg := &pb.Pb{
		Cmd:    "Disconnect",
		Gameid: gameid,
	}
	s := &base.Session{
		Msgque: m,
	}
	apis.CallApiMessage(s, pbmsg)
}
func (r *MyMsgHandler) OnProcessMsg(m antnet.IMsgQue, msg *antnet.Message) bool {
	pbmsg := &pb.Pb{}
	antnet.PBUnPack(msg.Data, pbmsg)
	s := &base.Session{
		Msgque: m,
	}
	return apis.CallApiMessage(s, pbmsg)
}
func (r *MyMsgHandler) OnConnectComplete(msgque antnet.IMsgQue, ok bool) bool { return true }

type TestSSS struct {
	Count int
}

func (t *TestSSS) Test123(n int) {
	m := t.Count + n
	fmt.Printf("test123 m=%v\n", m)
}

var methodMap map[string]reflect.Method

func reflectApis() {
	methodMap = make(map[string]reflect.Method)
	val := reflect.ValueOf(&TestSSS{})
	typ := val.Type()

	kd := val.Kind()
	if kd != reflect.Struct {
		//panic("apis expect struct")
	}
	num := val.NumMethod()
	for i := 0; i < num; i++ {
		fmt.Printf("method[%d]%s\n", i, typ.Method(i).Name)
		methodMap[typ.Method(i).Name] = typ.Method(i)
	}

	alluinfos := make([]int, 0)
	for i := 0; i < 36; i++ {
		alluinfos = append(alluinfos, i)
	}
	SPLITE_MSG_COUNT := 10
	from := 0
	end := SPLITE_MSG_COUNT
	sz := len(alluinfos)
	for {
		if end > sz {
			end = sz
		}
		Uinfo := alluinfos[from:end]
		fmt.Printf("%v\n", Uinfo)
		from = end
		end += SPLITE_MSG_COUNT
		if from >= sz {
			break
		}
	}
}

// 注册需要开放的游戏
func registerGames() {
	base.GetHallMgr().Register(int32(pb.GameId_Hzmj), hzmj.Create)
	base.GetHallMgr().Register(int32(pb.GameId_Ddz), ddz.Create)
}

func main() {
	reflectApis()

	t1 := &TestSSS{
		Count: 10,
	} //*/
	var params1 []reflect.Value
	params1 = append(params1, reflect.ValueOf(t1))
	params1 = append(params1, reflect.ValueOf(2))
	cb := methodMap["Test123"]

	//cb.SetPointer()
	//cb.SetPointer(unsafe.Pointer(t1))
	//cb.Set(reflect.ValueOf(t1))
	cb.Func.Call(params1)
	//cb.Call(params1)
	t2 := &TestSSS{
		Count: 100,
	} //*/
	var params2 []reflect.Value
	params2 = append(params2, reflect.ValueOf(t2))
	params2 = append(params2, reflect.ValueOf(2))
	cb.Func.Call(params2)
	//return

	// TODO 参数应该从配置中取
	utils.InitLog(true, 30, "log/gs.log", "debug")
	utils.TLog.Info("Server Start ...")
	registerGames()
	utils.InitRoomIdMgr(100000, 200000)

	msgHandler := &MyMsgHandler{}
	antnet.StartServer("ws://:5001", msgHandler)
	antnet.WaitForSystemExit(nil)
}
