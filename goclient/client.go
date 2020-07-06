package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"rgames/common/antnet"
	pb "rgames/protobuf"
	"strconv"
	"strings"
)

func main() {
	addr := "ws://127.0.0.1:5001"
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			pbmsg := &pb.Pb{}
			antnet.PBUnPack(message[8:], pbmsg)
			log.Printf("===> recv: %v", pbmsg.String())
		}
	}()

	for {
		inputReader := bufio.NewReader(os.Stdin)
		fmt.Printf(">>> ")
		input, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Println("errors reading, exiting program.")
			break
		}
		input = input[:len(input)-1]
		cmds := strings.Split(input, " ")
		fmt.Printf("cmd=%v cmds[%v[\n", input, cmds)

		var gameid int32 = 101

		switch cmds[0] {
		case "login":
			{
				req := &pb.CSLoginReq{
					Account: cmds[1],
					Gameid:  gameid,
					Passwd:  "welcome",
				}
				msg := &pb.Pb{LoginReq: req}
				msg.Gameid = req.Gameid
				msg.Cmd = "Login"
				msg.Roomid = 0

				bs, err := antnet.PBPack(msg)
				if err != nil {
					fmt.Println("errors PBPack")
					continue
				}
				m := antnet.NewDataMsg(bs, 0)
				c.WriteMessage(websocket.BinaryMessage, m.Bytes())
				fmt.Printf("send login\n")
			}
		case "create":
			{
				req := &pb.CSCreateRoomReq{
					Count:  2,
					Round:  3,
					Passwd: "hello",
					Jdata:  "hahaha",
				}
				msg := &pb.Pb{CreateRoomReq: req}
				msg.Gameid = gameid
				msg.Cmd = "CreateRoom"
				bs, err := antnet.PBPack(msg)
				if err != nil {
					fmt.Println("errors PBPack")
					continue
				}
				m := antnet.NewDataMsg(bs, 0)
				c.WriteMessage(websocket.BinaryMessage, m.Bytes())
				fmt.Printf("send create room\n")
			}
		case "enter":
			{
				roomid, _ := strconv.ParseInt(cmds[1], 10, 64)
				req := &pb.CSEnterRoomReq{
					Roomid: roomid,
					Passwd: "hello",
				}
				msg := &pb.Pb{EnterRoomReq: req}
				msg.Gameid = gameid
				msg.Cmd = "EnterRoom"
				bs, err := antnet.PBPack(msg)
				if err != nil {
					fmt.Println("errors PBPack")
					continue
				}
				m := antnet.NewDataMsg(bs, 0)
				c.WriteMessage(websocket.BinaryMessage, m.Bytes())
				fmt.Printf("send enter room\n")
			}
		case "leave":
			{
				req := &pb.CSLeaveRoomReq{
					Reason: int32(pb.LeaveReason_UserOp),
				}
				msg := &pb.Pb{LeaveRoomReq: req}
				msg.Gameid = gameid
				msg.Cmd = "LeaveRoom"
				bs, err := antnet.PBPack(msg)
				if err != nil {
					fmt.Println("errors PBPack")
					continue
				}
				m := antnet.NewDataMsg(bs, 0)
				c.WriteMessage(websocket.BinaryMessage, m.Bytes())
				fmt.Printf("send leave room\n")
			}
		case "ready":
			{
				req := &pb.CSReadyReq{}
				msg := &pb.Pb{ReadyReq: req}
				msg.Gameid = gameid
				msg.Cmd = "Ready"
				bs, err := antnet.PBPack(msg)
				if err != nil {
					fmt.Println("errors PBPack")
					continue
				}
				m := antnet.NewDataMsg(bs, 0)
				c.WriteMessage(websocket.BinaryMessage, m.Bytes())
				fmt.Printf("send ready\n")
			}
		case "unready":
			{
				req := &pb.CSCancelReadyReq{}
				msg := &pb.Pb{CancelReadyReq: req}
				msg.Gameid = gameid
				msg.Cmd = "CancelReady"
				bs, err := antnet.PBPack(msg)
				if err != nil {
					fmt.Println("errors PBPack")
					continue
				}
				m := antnet.NewDataMsg(bs, 0)
				c.WriteMessage(websocket.BinaryMessage, m.Bytes())
				fmt.Printf("send cancel ready\n")
			}
		case "other":
			{
				req := &pb.CSLoginReq{
					Account: "abc",
					Gameid:  gameid,
					Passwd:  "welcome",
				}
				msg := &pb.Pb{LoginReq: req}
				msg.Gameid = req.Gameid
				msg.Cmd = "HAHAH"
				msg.Roomid = 0

				bs, err := antnet.PBPack(msg)
				if err != nil {
					fmt.Println("errors PBPack")
					continue
				}
				m := antnet.NewDataMsg(bs, 0)
				c.WriteMessage(websocket.BinaryMessage, m.Bytes())
				fmt.Printf("send other\n")
			}
		case "quit":
			fmt.Printf("Disconnect\n")
			c.Close()
			goto clientEnd
		default:
			fmt.Println("what ???")
		}
	}
clientEnd:
	fmt.Printf("client end\n")
}
