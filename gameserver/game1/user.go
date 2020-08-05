package game1

import "rgames/common/base"

func init() {

}

type User struct {
	base.UserBase
	MImpl MoveImpl
}

func CreateUser(userid int64, connid int64) base.UserImpl {
	u := &User{
		MImpl: nil,
	}
	u.Init(u, userid, connid)
	return u
}

func (u *User) OnActionTimeOut(action int32) {

}
