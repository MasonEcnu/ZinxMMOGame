package api

import (
	"fmt"

	"aoi_mmo_game/core"
	"aoi_mmo_game/mmopb"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
	"github.com/golang/protobuf/proto"
)

// WorldChatRouter 世界聊天路由
type WorldChatRouter struct {
	znet.BaseRouter
}

func (*WorldChatRouter) Handle(request ziface.IRequest) {
	msg := &mmopb.Talk{}
	err := proto.Unmarshal(request.GetData(), msg)
	if err != nil {
		fmt.Println("Talk unmarshal error ", err)
		return
	}
	// 聊天消息是谁发的
	playerId, err := request.GetConnection().GetProperty("playerId")
	if err != nil {
		fmt.Println("GetProperty playerId error ", err)
		request.GetConnection().Stop()
		return
	}

	// 找到发聊天的player
	player := core.WorldMgrObj.GetPlayerById(playerId.(int32))

	if msg.TargetPlayerId > 0 {
		// 个人聊天
		player.TalkToTargetPlayer(msg.TargetPlayerId, msg.Content)
	} else {
		// 广播全服聊天
		player.BroadCastTalk(msg.Content)
	}
}
