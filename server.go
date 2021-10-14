package main

import (
	"fmt"

	"aoi_mmo_game/api"
	"aoi_mmo_game/core"
	"aoi_mmo_game/mmopb"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

func main() {
	// 创建服务
	s := znet.NewServer()

	// 设置客户端连接到来时的处理函数
	s.SetOnConnStart(onConnectionAdd)
	// 断线
	s.SetOnConnStop(onConnectionLost)

	// 注册聊天路由
	s.AddRouter(mmopb.CSMsgIdTalk, &api.WorldChatRouter{})
	// 移动路由
	s.AddRouter(mmopb.CSMsgIdMove, &api.PlayerMoveRouter{})

	// 开启服务
	s.Serve()
}

// onConnectionLost 客户端断开连接
func onConnectionLost(conn ziface.IConnection) {
	// 获得断线的玩家id
	playerId, err := conn.GetProperty("playerId")
	if err != nil || playerId.(int32) <= 0 {
		fmt.Println("conn property playerId not exist")
		conn.Stop()
		return
	}

	// 根据pid获取对应的玩家对象
	player := core.WorldMgrObj.GetPlayerById(playerId.(int32))

	// 触发玩家下线业务
	if player != nil {
		player.LostConnection()

		fmt.Println("======> player id = ", player.PlayerId, " left <======")
	}
}

// onConnectionAdd 当客户端建立连接时当hook函数
func onConnectionAdd(conn ziface.IConnection) {
	player := core.NewPlayer(conn)
	// 同步当前playerId给客户端，走msgId:1消息
	player.SyncPlayerId()
	// 同步当前玩家的初始化坐标信息给客户端，走msgId:200消息
	player.BroadCastStartPosition()

	// 添加到世界管理器
	core.WorldMgrObj.AddPlayer(player)

	// 绑定连接和playerId
	conn.SetProperty("playerId", player.PlayerId)

	// 同步周围玩家信息
	player.SyncSurrounding()

	fmt.Println("======> player id = ", player.PlayerId, " arrived <======")
}
