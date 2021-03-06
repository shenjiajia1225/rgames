// Code generated by protoc-gen-go. DO NOT EDIT.
// source: errno.proto

package protocol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type ErrNo int32

const (
	ErrNo_Success       ErrNo = 0
	ErrNo_BreakEnter    ErrNo = 1
	ErrNo_Unknow        ErrNo = -1
	ErrNo_AleardyInRoom ErrNo = -11
	ErrNo_NotInRoom     ErrNo = -12
	ErrNo_RoomInGame    ErrNo = -13
	ErrNo_NotInSeat     ErrNo = -14
	ErrNo_InvalidState  ErrNo = -15
)

var ErrNo_name = map[int32]string{
	0:   "Success",
	1:   "BreakEnter",
	-1:  "Unknow",
	-11: "AleardyInRoom",
	-12: "NotInRoom",
	-13: "RoomInGame",
	-14: "NotInSeat",
	-15: "InvalidState",
}
var ErrNo_value = map[string]int32{
	"Success":       0,
	"BreakEnter":    1,
	"Unknow":        -1,
	"AleardyInRoom": -11,
	"NotInRoom":     -12,
	"RoomInGame":    -13,
	"NotInSeat":     -14,
	"InvalidState":  -15,
}

func (x ErrNo) String() string {
	return proto.EnumName(ErrNo_name, int32(x))
}
func (ErrNo) EnumDescriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func init() {
	proto.RegisterEnum("protocol.ErrNo", ErrNo_name, ErrNo_value)
}

func init() { proto.RegisterFile("errno.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 177 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4e, 0x2d, 0x2a, 0xca,
	0xcb, 0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x00, 0x53, 0xc9, 0xf9, 0x39, 0x5a, 0x3b,
	0x19, 0xb9, 0x58, 0x5d, 0x8b, 0x8a, 0xfc, 0xf2, 0x85, 0xb8, 0xb9, 0xd8, 0x83, 0x4b, 0x93, 0x93,
	0x53, 0x8b, 0x8b, 0x05, 0x18, 0x84, 0xf8, 0xb8, 0xb8, 0x9c, 0x8a, 0x52, 0x13, 0xb3, 0x5d, 0xf3,
	0x4a, 0x52, 0x8b, 0x04, 0x18, 0x85, 0x84, 0xb9, 0xd8, 0x42, 0xf3, 0xb2, 0xf3, 0xf2, 0xcb, 0x05,
	0xfe, 0xc3, 0x00, 0xa3, 0x90, 0x14, 0x17, 0xaf, 0x63, 0x4e, 0x6a, 0x62, 0x51, 0x4a, 0xa5, 0x67,
	0x5e, 0x50, 0x7e, 0x7e, 0xae, 0xc0, 0x57, 0x84, 0x9c, 0x18, 0x17, 0xa7, 0x5f, 0x7e, 0x09, 0x54,
	0xfc, 0x0b, 0x42, 0x5c, 0x9c, 0x8b, 0x0b, 0x24, 0xe4, 0x99, 0xe7, 0x9e, 0x98, 0x9b, 0x2a, 0xf0,
	0x19, 0x53, 0x43, 0x70, 0x6a, 0x62, 0x89, 0xc0, 0x27, 0x84, 0xb8, 0x24, 0x17, 0x8f, 0x67, 0x5e,
	0x59, 0x62, 0x4e, 0x66, 0x4a, 0x70, 0x49, 0x62, 0x49, 0xaa, 0xc0, 0x47, 0xb8, 0x54, 0x12, 0x1b,
	0xd8, 0x17, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x27, 0xad, 0x30, 0x72, 0xdb, 0x00, 0x00,
	0x00,
}
