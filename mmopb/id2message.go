package mmopb

import (
	"github.com/golang/protobuf/proto"
)

// 客户端消息
const (
	CSMsgIdTalk uint32 = 1
	CSMsgIdMove uint32 = 2
)

// 服务器消息
const (
	SCMsgIdSyncPlayerId uint32 = 1
	SCMsgIdBroadCast    uint32 = 2
	SCMsgIdSyncPlayers  uint32 = 3
	SCMsgIdMove         uint32 = 4
	SCMsgIdPlayerLeave  uint32 = 5
)

// SCId2Message server to client id message map
var SCId2Message map[uint32]proto.Message

// CSId2Message client to server id message map
var CSId2Message map[uint32]proto.Message

func init() {
	// 客户端消息
	CSId2Message = map[uint32]proto.Message{
		CSMsgIdTalk: &BroadCast{},
		CSMsgIdMove: &BroadCast{},
	}

	// 服务器消息
	SCId2Message = map[uint32]proto.Message{
		SCMsgIdSyncPlayerId: &SyncPlayerId{},
		SCMsgIdBroadCast:    &BroadCast{},
		SCMsgIdSyncPlayers:  &SyncPlayers{},
		SCMsgIdMove:         &Position{},
		SCMsgIdPlayerLeave:  &SyncPlayerId{},
	}
}
