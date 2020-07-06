package hzmj

import "rgames/common/base"

func init() {

}

type User struct {
	base.UserBase
}

func CreateUser(userid int64, connid int64) base.UserImpl {
	u := &User{}
	u.Init(u, userid, connid)
	return u
}

func (u *User) OnActionTimeOut(action int32) {

}
