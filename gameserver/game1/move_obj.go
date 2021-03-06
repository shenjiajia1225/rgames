package game1

import (
	"fmt"
	"rgames/common/utils"
	pb "rgames/protobuf"
)

func init() {
}

type MoveObj struct {
	Pos        *pb.Vec3
	Dir        *pb.Vec3
	MvType     pb.MoveType
	Speed      int32
	LastUpTime int64
	Targets    []*pb.Vec3
}

type MoveImpl interface {
	Start(id int64, speed int32, mtype pb.MoveType, vecs []*pb.Vec3)
	Stop(id int64)
	TurnUp(id int64, speed int32, dir *pb.Vec3)
	Update(now int64)
}

func (mo *MoveObj) Start(id int64, speed int32, mtype pb.MoveType, vecs []*pb.Vec3) {
	if vecs == nil || len(vecs) <= 0 || mtype == pb.MoveType_Stop {
		return
	}
	mo.MvType = mtype
	mo.Speed = speed
	mo.LastUpTime = 0
	if mtype == pb.MoveType_Direction {
		mo.Dir = vecs[0]
		mo.Targets = nil
	} else if mtype == pb.MoveType_Direction {
		mo.Targets = vecs
	}
	MoveMgrInstance().Add(id, mo)
	utils.TLog.Debug(fmt.Sprintf("move start [%v] pos[%v,%v,%v] dir[%v,%v,%v] speed[%v]",
		id, mo.Pos.X, mo.Pos.Y, mo.Pos.Z, mo.Dir.X, mo.Dir.Y, mo.Dir.Z, speed))
}

func (mo *MoveObj) Stop(id int64) {
	mo.Speed = 0
	MoveMgrInstance().Remove(id)
	utils.TLog.Debug(fmt.Sprintf("move stop [%v] pos[%v,%v,%v] dir[%v,%v,%v]",
		id, mo.Pos.X, mo.Pos.Y, mo.Pos.Z, mo.Dir.X, mo.Dir.Y, mo.Dir.Z))
}

func (mo *MoveObj) TurnUp(id int64, speed int32, dir *pb.Vec3) {
	mo.Speed = speed
	mo.Dir = dir
	utils.TLog.Debug(fmt.Sprintf("move turnup [%v] pos[%v,%v,%v] dir[%v,%v,%v] speed[%v]",
		id, mo.Pos.X, mo.Pos.Y, mo.Pos.Z, mo.Dir.X, mo.Dir.Y, mo.Dir.Z, speed))
}

func (mo *MoveObj) Update(now int64) {
	dis := (now - mo.LastUpTime) * int64(mo.Speed) * 1000
	mo.Pos.X += dis * mo.Dir.X
	mo.Pos.Y += dis * mo.Dir.Y
	mo.Pos.Z += dis * mo.Dir.Z
	utils.TLog.Debug(fmt.Sprintf("move update pos[%v,%v,%v] dir[%v,%v,%v] dis[%v]",
		mo.Pos.X, mo.Pos.Y, mo.Pos.Z, mo.Dir.X, mo.Dir.Y, mo.Dir.Z, dis))
}
