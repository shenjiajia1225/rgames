package apis

import (
	"fmt"
	"reflect"
	"rgames/common/base"
	"rgames/common/utils"
	pb "rgames/protobuf"
)

func init() {
	apiMsg = ApiMessage{}
	methodMap = make(map[string]reflect.Value)
	reflectApis()
}

type ApiMessage struct {
}

var methodMap map[string]reflect.Value
var apiMsg ApiMessage

func reflectApis() {
	val := reflect.ValueOf(apiMsg)
	typ := val.Type()

	kd := val.Kind()
	if kd != reflect.Struct {
		panic("apis expect struct")
	}

	num := val.NumMethod()
	for i := 0; i < num; i++ {
		//fmt.Printf("method[%d]%s\n", i, typ.Method(i).Name)
		methodMap[typ.Method(i).Name] = val.Method(i)
	}
}

func CallApiMessage(s *base.Session, msg *pb.Pb) bool {
	defer func() {
		if r := recover(); r != nil {
			utils.TLog.Error("CallApiMessage panic")
		}
	}()

	utils.TLog.Debug(fmt.Sprintf("CallApiMessage cmd[%v] gameid[%v] roomid[%v]", msg.GetCmd(), msg.GetGameid(), msg.GetRoomid()))

	h := base.GetHallMgr().Get(msg.GetGameid())
	if h == nil {
		return true
	}

	um := &base.UserMessage{
		S:   s,
		Msg: msg,
	}
	h.SendMessage(um)
	return true

	/*
		if cb, ok := methodMap[msg.GetCmd()]; ok {
			var params []reflect.Value
			params = append(params, reflect.ValueOf(s))
			params = append(params, reflect.ValueOf(msg))
			ret := cb.Call(params)
			return ret[0].Bool()
		} else {
			return HandleOtherMessage(s, msg)
		}
	*/
}

func HandleOtherMessage(s *base.Session, msg *pb.Pb) bool {
	return true
}

/////////////////////////////// ALL Apis ///////////////////////////////////////////////////

func (a ApiMessage) Test123(s *base.Session, msg *pb.Pb) bool {
	fmt.Printf("test123\n")
	return true
}

func (a ApiMessage) KeepAlive(s *base.Session, msg *pb.Pb) bool {
	return true
}

func (a ApiMessage) Login(s *base.Session, msg *pb.Pb) bool {
	h := base.GetHallMgr().Get(msg.Gameid)
	if h == nil {
		panic("Login No hall")
	}

	um := &base.UserMessage{
		S:   s,
		Msg: msg,
	}
	h.SendMessage(um)
	return true
}
