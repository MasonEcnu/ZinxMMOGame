package api

import (
	"fmt"

	"aoi_mmo_game/core"
	"aoi_mmo_game/mmopb"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
	"github.com/golang/protobuf/proto"
)

// PlayerMoveRouter 玩家移动路由
type PlayerMoveRouter struct {
	znet.BaseRouter
}

func (*PlayerMoveRouter) Handle(request ziface.IRequest) {
	msg := &mmopb.Position{}
	err := proto.Unmarshal(request.GetData(), msg)
	if err != nil {
		fmt.Println("Position unmarshal error ", err)
		return
	}
	// 谁要移动
	playerId, err := request.GetConnection().GetProperty("playerId")
	if err != nil {
		fmt.Println("GetProperty playerId error ", err)
		request.GetConnection().Stop()
		return
	}

	// 找到要移动的player
	player := core.WorldMgrObj.GetPlayerById(playerId.(int32))

	fmt.Printf("user pid = %d , move(%f,%f,%f,%f)", playerId, msg.X, msg.Y, msg.Z, msg.V)

	if player != nil {
		// 更新位置
		player.UpdatePos(msg.X, msg.Y, msg.Z, msg.V)
	}
}
