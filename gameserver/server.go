package main

import (
	//"fmt"
	//"reflect"
	"net/http"
	_ "net/http/pprof"
	"rgames/common/antnet"
	"rgames/common/base"
	"rgames/gameserver/ddz"
	"rgames/gameserver/game1"
	"rgames/gameserver/hzmj"
	//"unsafe"

	//"rgames/github.com/golang/protobuf/proto"
	apis "rgames/common/apis"
	"rgames/common/utils"
	pb "rgames/protobuf"
)

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

// 注册需要开放的游戏
func installGames() {
	base.GetHallMgr().Register(int32(pb.GameId_Hzmj), hzmj.Create)
	base.GetHallMgr().Register(int32(pb.GameId_Ddz), ddz.Create)
	base.GetHallMgr().Register(int32(pb.GameId_Game1), game1.Create)
}

func unInstallGames() {
	base.GetHallMgr().UnRegisterAll()
}

func main() {
	// TODO 参数应该从配置中取
	utils.InitLog(true, 30, "log/gs.log", "debug")
	utils.TLog.Info("Server Start ...")
	installGames()
	utils.InitRoomIdMgr(100000, 200000)

	go func(pprofAddr string) {
		defer utils.TryCatch()
		http.ListenAndServe(pprofAddr, nil)
	}("0.0.0.0:9900")

	msgHandler := &MyMsgHandler{}
	antnet.StartServer("ws://:5001", msgHandler)
	antnet.WaitForSystemExit(nil)

	unInstallGames()
	utils.TLog.Info("Server Stop")
}
