package core

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"aoi_mmo_game/mmopb"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/zinx_app_demo/mmo_game/pb"
	"github.com/golang/protobuf/proto"
)

// Player 玩家对象
type Player struct {
	PlayerId int32              // 玩家id
	Conn     ziface.IConnection // 当前玩家连接
	X        float32            // 平面x坐标
	Y        float32            // 高度
	Z        float32            // 平面y坐标
	V        float32            // 旋转0-360度
}

// playerIdGen playerId生成器
var playerIdGen int32 = 1
var playerIdLock sync.Mutex

func NewPlayer(conn ziface.IConnection) *Player {
	playerIdLock.Lock()
	playerId := playerIdGen
	playerIdGen++
	playerIdLock.Unlock()

	return &Player{
		PlayerId: playerId,
		Conn:     conn,
		X:        float32(160 + rand.Intn(10)),
		Y:        0,
		Z:        float32(134 + rand.Intn(17)),
		V:        0,
	}
}

func (p *Player) SendMessage(msgId uint32, data proto.Message) {
	fmt.Printf("before Marshal data = %+v\n", data)
	msg, err := proto.Marshal(data)
	if err != nil {
		fmt.Println("marshal msg err: ", err)
		return
	}
	fmt.Printf("after Marshal data = %+v\n", data)

	if p.Conn == nil {
		fmt.Println("connection in player is nil")
		return
	}

	fmt.Println("send message id=", msgId, " len=", len(msg))
	if err := p.Conn.SendMsg(msgId, msg); err != nil {
		fmt.Println("player send message err: ", err)
		return
	}
}

// SyncPlayerId 同步playerId给客户端
func (p *Player) SyncPlayerId() {
	msg := &mmopb.SyncPlayerId{
		PlayerId: p.PlayerId,
	}
	p.SendMessage(mmopb.SCMsgIdSyncPlayerId, msg)
}

// BroadCastStartPosition 广播玩家的出生地信息
func (p *Player) BroadCastStartPosition() {
	msg := &mmopb.BroadCast{
		PlayerId: p.PlayerId,
		Type:     mmopb.BroadCastType_Player_Pos,
		Data: &mmopb.BroadCast_Pos{
			Pos: &mmopb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	// 告知自己的位置
	p.SendMessage(mmopb.SCMsgIdBroadCast, msg)
}

// BroadCastTalk 广播玩家聊天
func (p *Player) BroadCastTalk(content string) {
	msg := &mmopb.BroadCast{
		PlayerId: p.PlayerId,
		Type:     mmopb.BroadCastType_World_Chat,
		Data: &mmopb.BroadCast_Content{
			Content: content,
		},
	}

	players := WorldMgrObj.GetAllPlayers()

	for _, player := range players {
		player.SendMessage(mmopb.SCMsgIdBroadCast, msg)
	}
}

func (p *Player) TalkToTargetPlayer(targetPlayerId int32, content string) {
	targetPlayer := WorldMgrObj.GetPlayerById(targetPlayerId)
	if targetPlayer != nil {
		talkToMsg := &mmopb.BroadCast{
			PlayerId: p.PlayerId,
			Type:     mmopb.BroadCastType_World_Chat,
			Data: &mmopb.BroadCast_Content{
				Content: content,
			},
		}
		targetPlayer.SendMessage(mmopb.SCMsgIdBroadCast, talkToMsg)
	}
}

// SyncSurrounding 给当前九宫格范围内玩家广播自己的位置
func (p *Player) SyncSurrounding() {
	// 找出附近的玩家id
	pids := WorldMgrObj.AoiMgr.GetPlayerIdsByPos(p.X, p.Z)
	// 找出附近的玩家对象
	players := make([]*Player, len(pids))
	for _, pid := range pids {
		player := WorldMgrObj.GetPlayerById(int32(pid))
		if player != nil {
			players = append(players, player)
		}
	}

	msg := &mmopb.BroadCast{
		PlayerId: p.PlayerId,
		Type:     mmopb.BroadCastType_Player_Pos,
		Data: &mmopb.BroadCast_Pos{
			Pos: &mmopb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	// 发送位置消息，并对自己同步周围玩家信息
	playersData := make([]*mmopb.Player, len(players))
	for _, player := range players {
		if player == nil {
			continue
		}

		// 不用自己给自己同步位置
		if p.PlayerId != player.PlayerId {
			player.SendMessage(mmopb.SCMsgIdBroadCast, msg)

			mmoplayer := &mmopb.Player{
				PlayerId: player.PlayerId,
				Pos: &mmopb.Position{
					X: player.X,
					Y: player.Y,
					Z: player.Z,
					V: player.V,
				},
			}
			playersData = append(playersData, mmoplayer)
		}
	}

	syncMsg := &mmopb.SyncPlayers{
		Players: playersData,
	}
	p.SendMessage(mmopb.SCMsgIdSyncPlayers, syncMsg)
}

// UpdatePos 玩家更新位置
func (p *Player) UpdatePos(x float32, y float32, z float32, v float32) {
	// 计算新旧格子变化
	oldGid := WorldMgrObj.AoiMgr.GetGidByPos(p.X, p.Z)
	newGid := WorldMgrObj.AoiMgr.GetGidByPos(x, z)

	// 更新玩家坐标
	p.X = x
	p.Y = y
	p.Z = z
	p.V = v

	if oldGid != newGid {
		// 切换格子
		// 把playerId从旧的grid中移除
		WorldMgrObj.AoiMgr.RemovePlayerIdFromGrid(int(p.PlayerId), oldGid)
		// 添加到新的格子去
		WorldMgrObj.AoiMgr.AddPlayerIdToGrid(int(p.PlayerId), oldGid)
		// 视野切换
		_ = p.OnExchangeAoiGrid(oldGid, newGid)
	}

	// 同步自己的位置给周围玩家
	msg := &mmopb.BroadCast{
		PlayerId: p.PlayerId,
		Type:     mmopb.BroadCastType_After_Move,
		Data: &mmopb.BroadCast_Pos{
			Pos: &mmopb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	// 找到附近的玩家
	players := p.GetSurroundingPlayers()
	for _, player := range players {
		if player != nil && player.PlayerId > 0 {
			player.SendMessage(mmopb.SCMsgIdBroadCast, msg)
		}
	}
}

// GetSurroundingPlayers 找到九宫格内的所有玩家
func (p *Player) GetSurroundingPlayers() []*Player {
	// 获得当前aoi区域的所有pid
	pids := WorldMgrObj.AoiMgr.GetPlayerIdsByPos(p.X, p.Z)

	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerById(int32(pid)))
	}

	return players
}

// LostConnection 玩家下线
func (p *Player) LostConnection() {
	// 1 获取周围AOI九宫格内的玩家
	players := p.GetSurroundingPlayers()

	// 2 封装MsgID:201消息
	msg := &pb.SyncPid{
		Pid: p.PlayerId,
	}

	// 3 向周围玩家发送消息
	for _, player := range players {
		// 不用通知自己吧
		if player != nil && player.PlayerId != p.PlayerId {
			player.SendMessage(mmopb.SCMsgIdPlayerLeave, msg)
		}
	}

	// 4 世界管理器将当前玩家从AOI中摘除
	WorldMgrObj.AoiMgr.RemoveFromGridByPos(int(p.PlayerId), p.X, p.Z)
	WorldMgrObj.RemovePlayerById(p.PlayerId)
}

// OnExchangeAoiGrid 跨格子视野切换
func (p *Player) OnExchangeAoiGrid(oldGid int, newGid int) error {
	// 获取旧的九宫格成员
	oldGrids := WorldMgrObj.AoiMgr.GetSurroundGridsByGid(oldGid)
	// 为旧的九宫格成员建立哈希表,用来快速查找
	oldGridsMap := make(map[int]struct{}, len(oldGrids))
	for _, grid := range oldGrids {
		oldGridsMap[grid.GID] = struct{}{}
	}

	// 获取新的九宫格成员
	newGrids := WorldMgrObj.AoiMgr.GetSurroundGridsByGid(newGid)
	newGridsMap := make(map[int]struct{}, len(newGrids))
	for _, grid := range newGrids {
		newGridsMap[grid.GID] = struct{}{}
	}

	// ========== 处理视野消失 ==========
	offlineMsg := &mmopb.SyncPlayerId{
		PlayerId: p.PlayerId,
	}

	// 找到在旧的九宫格中出现,但是在新的九宫格中没有出现的格子
	leavingGrids := make([]*Grid, 0)
	for _, grid := range oldGrids {
		if _, ok := newGridsMap[grid.GID]; !ok {
			leavingGrids = append(leavingGrids, grid)
		}
	}

	// 获取需要消失的格子中的全部玩家
	for _, grid := range leavingGrids {
		players := WorldMgrObj.GetPlayersByGid(grid.GID)
		for _, player := range players {
			if player != nil {
				// 让自己在其他玩家的客户端中消失
				player.SendMessage(mmopb.SCMsgIdPlayerLeave, offlineMsg)

				// 将其他玩家信息 在自己的客户端中消失
				anotherOfflineMsg := &mmopb.SyncPlayerId{
					PlayerId: player.PlayerId,
				}
				p.SendMessage(mmopb.SCMsgIdPlayerLeave, anotherOfflineMsg)
				time.Sleep(200 * time.Millisecond)
			}
		}
	}

	// ========== 处理视野出现 ==========
	enteringGrids := make([]*Grid, 0)
	for _, grid := range newGrids {
		if _, ok := oldGridsMap[grid.GID]; !ok {
			enteringGrids = append(enteringGrids, grid)
		}
	}

	onlineMsg := &mmopb.BroadCast{
		PlayerId: p.PlayerId,
		Type:     mmopb.BroadCastType_Player_Pos,
		Data: &mmopb.BroadCast_Pos{
			Pos: &mmopb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	// 获取需要显示格子的全部玩家
	for _, grid := range enteringGrids {
		players := WorldMgrObj.GetPlayersByGid(grid.GID)
		for _, player := range players {
			if player != nil {
				// 让自己出现在别人视野中
				player.SendMessage(mmopb.SCMsgIdBroadCast, onlineMsg)

				// 让其他人出现在自己的视野中
				anotherOnlineMsg := &mmopb.BroadCast{
					PlayerId: player.PlayerId,
					Type:     mmopb.BroadCastType_Player_Pos,
					Data: &mmopb.BroadCast_Pos{
						Pos: &mmopb.Position{
							X: player.X,
							Y: player.Y,
							Z: player.Z,
							V: player.V,
						},
					},
				}

				p.SendMessage(mmopb.SCMsgIdBroadCast, anotherOnlineMsg)
				time.Sleep(200 * time.Millisecond)
			}
		}
	}
	return nil
}
