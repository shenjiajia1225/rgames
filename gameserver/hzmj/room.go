package hzmj

import (
	"fmt"
	"rgames/common/base"
	"rgames/common/utils"
	pb "rgames/protobuf"
	"time"
)

func init() {

}

type Room struct {
	base.RoomBase
	startTime int64
}

func CreateRoom(count int32, passwd string) base.RoomImpl {
	r := &Room{}
	r.Init(r, count, passwd)
	return r
}

func (r *Room) OnTimeOut(now int64) {
	// for testing. 5s one round
	gtime := now - r.startTime
	if r.GetState() == int32(pb.RoomState_Gaming) && gtime >= 5 {
		r.GameEnd()
	}
}

func (r *Room) OnGameStart() {
	r.startTime = time.Now().Unix()
	utils.TLog.Info(fmt.Sprintf("hzmj OnGameStart room[%v] curRound[%v] start[%v]", r.Id, r.CurCount, r.startTime))

	ntf := &pb.SCGameStartNtf{}
	r.BroadcastMsg(&pb.Pb{
		Cmd:          "GameStartNtf",
		Tserver:      time.Now().Unix(),
		GameStartNtf: ntf,
	})
}

func (r *Room) OnGameEnd() {
	endTime := time.Now().Unix()
	utils.TLog.Info(fmt.Sprintf("hzmj OnGameEnd room[%v] curRound[%v] end[%v]", r.Id, r.CurCount, endTime))

	ntf := &pb.SCGameEndNtf{}
	r.BroadcastMsg(&pb.Pb{
		Cmd:        "GameEndNtf",
		Tserver:    time.Now().Unix(),
		GameEndNtf: ntf,
	})
}

func (r *Room) OnRoundEnd() {
	endTime := time.Now().Unix()
	utils.TLog.Info(fmt.Sprintf("hzmj OnRoundEnd room[%v] curRound[%v] end[%v]", r.Id, r.CurCount, endTime))

	ntf := &pb.SCRoundEndNtf{}
	r.BroadcastMsg(&pb.Pb{
		Cmd:         "RoundEndNtf",
		Tserver:     time.Now().Unix(),
		RoundEndNtf: ntf,
	})
}

func (r *Room) HandleOtherMessage(um *base.UserMessage) bool {
	return false
}

func (r *Room) CanEnter(user base.UserImpl, passwd string) int32 {
	ret := r.RoomBase.CanEnter(user, passwd)
	return ret
}

func (r *Room) CanLeave(user base.UserImpl, reason int32) int32 {
	ret := r.RoomBase.CanLeave(user, reason)
	return ret
}
