package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"

	"aoi_mmo_game/mmopb"

	"github.com/aceld/zinx/znet"
	"github.com/golang/protobuf/proto"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:12121")
	if err != nil {
		log.Fatalln("client start err, exit!", err)
	}

	// 启一个协程用于接收消息
	go receiveMessage(conn)

	// 启一个协程用于发送消息
	go sendMessage(conn)

	// 挂起client
	func() {
		select {}
	}()
}

func sendMessage(conn net.Conn) {
	for {
		showMenu()
		var choose uint

		_, err := fmt.Scan(&choose)
		if err != nil {
			fmt.Println("输入错误，请重新输入！")
			continue
		} else {
			_, ok := orderMap[choose]
			if !ok {
				fmt.Println("输入命令不存在，请重新输入！")
				continue
			}
		}

		switch choose {
		case 0:
			_ = conn.Close()
			fmt.Println("程序退出！")
			os.Exit(0)
		case 1:
			handleMove(conn)
		case 2:
			handleSingleTalk(conn)
		case 3:
			handleServerTalk(conn)
		}
	}
}

func handleServerTalk(conn net.Conn) {
	fmt.Println("请输入聊天内容")
	var content string
	scanf, err := fmt.Scanf("%s", &content)
	if err != nil || scanf != 1 || len(content) == 0 {
		log.Println("handleServerTalk--输入错误或参数个数不足!", err)
		return
	}

	request := &mmopb.Talk{
		Content: content,
	}

	// 发封包message消息
	dp := znet.NewDataPack()
	data, err := proto.Marshal(request)
	if err != nil {
		log.Println("handleServerTalk--proto Marshal错误!", err)
		return
	}

	msg, _ := dp.Pack(znet.NewMsgPackage(mmopb.CSMsgIdTalk, data))
	_, err = conn.Write(msg)
	if err != nil {
		log.Println("handleServerTalk--conn写入数据错误!", err)
		return
	}
}

func handleSingleTalk(conn net.Conn) {
	fmt.Println("请输入目标玩家id、聊天内容（参数用空格分割）")
	var playerId int32
	var content string
	scanf, err := fmt.Scanf("%d %s", &playerId, &content)
	if err != nil || scanf != 2 || len(content) == 0 || playerId <= 0 {
		log.Println("handleServerTalk--输入错误或参数个数不足!", err)
		return
	}

	request := &mmopb.Talk{
		TargetPlayerId: playerId,
		Content:        content,
	}

	// 发封包message消息
	dp := znet.NewDataPack()
	data, err := proto.Marshal(request)
	if err != nil {
		log.Println("handleServerTalk--proto Marshal错误!", err)
		return
	}

	msg, _ := dp.Pack(znet.NewMsgPackage(mmopb.CSMsgIdTalk, data))
	_, err = conn.Write(msg)
	if err != nil {
		log.Println("handleServerTalk--conn写入数据错误!", err)
		return
	}
}

func handleMove(conn net.Conn) {
	fmt.Println("请输入移动的位置x、y、z、v（参数用空格分割）")
	var x int32
	var y int32
	var z int32
	var v int32
	scanf, err := fmt.Scanf("%d %d %d %d", &x, &y, &z, &v)
	if err != nil || scanf != 4 {
		log.Println("handleMove--输入错误或参数个数不足!", err)
		return
	}

	request := &mmopb.Position{
		X: float32(x),
		Y: float32(y),
		Z: float32(z),
		V: float32(v),
	}

	// 发封包message消息
	dp := znet.NewDataPack()
	data, err := proto.Marshal(request)
	if err != nil {
		log.Println("handleMove--proto Marshal错误!", err)
		return
	}

	msg, _ := dp.Pack(znet.NewMsgPackage(mmopb.CSMsgIdMove, data))
	_, err = conn.Write(msg)
	if err != nil {
		log.Println("handleServerTalk--conn写入数据错误!", err)
		return
	}
}

var orderMap = map[uint]string{
	0: "退出",
	1: "移动",
	2: "个人聊天",
	3: "全服聊天",
}

var orderSlice = []uint{0, 1, 2, 3}

func showMenu() {
	fmt.Println("客户端功能菜单：")
	for _, order := range orderSlice {
		desc := orderMap[order]
		fmt.Printf("%d ====> %s\n", order, desc)
	}
	fmt.Println("请输入功能选项：")
}

func receiveMessage(conn net.Conn) {
	for {
		dp := znet.NewDataPack()
		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err := io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			fmt.Println("read head error")
			break
		}
		//将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack err:", err)
			return
		}

		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err:", err)
				return
			}

			pbmsg, ok := mmopb.SCId2Message[msg.Id]
			if !ok {
				fmt.Println("msg id is not exist: ", msg.Id)
				return
			}
			err = proto.Unmarshal(msg.Data, pbmsg)
			if err != nil {
				fmt.Println("resolving proto message failed: ", msg.Id)
				return
			}

			fmt.Printf("==> Receive Msg: ID=%d, message=%s\n", msg.Id, convertOctonaryUtf8(pbmsg.String()))
		}
	}
}

// convertOctonaryUtf8 将八进制utf8编码的中文转正常显示
func convertOctonaryUtf8(in string) string {
	s := []byte(in)
	reg := regexp.MustCompile(`\\[0-7]{3}`)

	out := reg.ReplaceAllFunc(s,
		func(b []byte) []byte {
			i, _ := strconv.ParseInt(string(b[1:]), 8, 0)
			return []byte{byte(i)}
		})
	return string(out)
}
