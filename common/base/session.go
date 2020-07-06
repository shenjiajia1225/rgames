package base

import (
	"rgames/common/antnet"
	pb "rgames/protobuf"
)

type Session struct {
	Msgque antnet.IMsgQue
}

func (s *Session) Send(pbmsg *pb.Pb) bool {
	bs, err := antnet.PBPack(pbmsg)
	if err != nil {
		return false
	}
	m := antnet.NewDataMsg(bs, 0)
	return s.Msgque.Send(m)
}

func (s *Session) ConnId() int64 {
	return int64(s.Msgque.Id())
}
