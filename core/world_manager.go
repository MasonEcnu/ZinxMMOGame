package core

import (
	"sync"
)

// WorldManager 游戏世界管理器
type WorldManager struct {
	AoiMgr     *AOIManager       // 世界地图aoi管理器
	Players    map[int32]*Player // 在线玩家集合
	playerLock sync.RWMutex      // 保护Players的读写锁
}

// WorldMgrObj 提供一个对外的句柄
var WorldMgrObj *WorldManager

func init() {
	WorldMgrObj = &WorldManager{
		AoiMgr:  NewAOIManager(AOI_MIN_X, AOI_MAX_X, AOI_CNTS_X, AOI_MIN_Y, AOI_MAX_Y, AOI_CNTS_Y),
		Players: make(map[int32]*Player, 50),
	}
}

// AddPlayer 玩家上线，添加到世界管理器到玩家列表中
func (wm *WorldManager) AddPlayer(player *Player) {
	// 添加到世界管理器
	wm.playerLock.Lock()
	wm.Players[player.PlayerId] = player
	wm.playerLock.Unlock()

	// 添加到aoi网格中
	wm.AoiMgr.AddPlayerIdToGridByPos(int(player.PlayerId), player.X, player.Z)
}

// RemovePlayerById 玩家下线，从世界管理器中移除
func (wm *WorldManager) RemovePlayerById(playerId int32) {
	// 添加到世界管理器
	wm.playerLock.Lock()
	delete(wm.Players, playerId)
	wm.playerLock.Unlock()
}

// GetPlayerById 通过id获取玩家信息
func (wm *WorldManager) GetPlayerById(playerId int32) *Player {
	// 添加到世界管理器
	wm.playerLock.RLock()
	defer wm.playerLock.RUnlock()
	return wm.Players[playerId]
}

// GetAllPlayers 获取全部在线玩家
func (wm *WorldManager) GetAllPlayers() []*Player {
	// 添加到世界管理器
	wm.playerLock.RLock()
	defer wm.playerLock.RUnlock()

	players := make([]*Player, 0, len(wm.Players))

	for _, p := range wm.Players {
		players = append(players, p)
	}
	return players
}

// GetPlayersByGid 获取指定gid中的所有player信息
func (wm *WorldManager) GetPlayersByGid(gid int) (players []*Player) {
	if grid, ok := wm.AoiMgr.grids[gid]; ok {
		players = make([]*Player, len(grid.GetPlayerIds()))
		wm.playerLock.RLock()
		for _, playerId := range grid.GetPlayerIds() {
			if player, ok := wm.Players[int32(playerId)]; ok {
				players = append(players, player)
			}
		}
		wm.playerLock.RUnlock()
	}
	return
}
