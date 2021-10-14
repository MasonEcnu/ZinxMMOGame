package core

import (
	"fmt"
	"sync"
)

// Grid 地图上的一个格子
type Grid struct {
	GID            int              // 格子ID
	MinX           int              // 格子左边界坐标
	MaxX           int              // 格子右边界坐标
	MinY           int              // 格子上边界坐标
	MaxY           int              // 格子下边界坐标
	playerIds      map[int]struct{} // 当前格子内的玩家或物体成员id
	playerIdIdLock sync.RWMutex     // playerIdId的map保护锁
}

// NewGrid 创建一个格子
func NewGrid(gid, minx, maxx, miny, maxy int) *Grid {
	return &Grid{
		GID:       gid,
		MinX:      minx,
		MaxX:      maxx,
		MinY:      miny,
		MaxY:      maxy,
		playerIds: make(map[int]struct{}),
	}
}

// Add 向格子中添加一个玩家
func (g *Grid) Add(playerId int) {
	g.playerIdIdLock.Lock()
	defer g.playerIdIdLock.Unlock()

	g.playerIds[playerId] = struct{}{}
}

// Remove 从格子中移除一个玩家
func (g *Grid) Remove(playerId int) {
	g.playerIdIdLock.Lock()
	defer g.playerIdIdLock.Unlock()

	delete(g.playerIds, playerId)
}

// GetPlayerIds 获取格子中的所有玩家
func (g *Grid) GetPlayerIds() []int {
	g.playerIdIdLock.RLock()
	defer g.playerIdIdLock.RUnlock()
	playerIds := make([]int, len(g.playerIds))
	for id, _ := range g.playerIds {
		playerIds = append(playerIds, id)
	}
	return playerIds
}

// String 格子结构消息
func (g *Grid) String() string {
	return fmt.Sprintf("Grid id: %d, minX:%d, maxX:%d, minY:%d, maxY:%d, playerIds:%v",
		g.GID, g.MinX, g.MaxX, g.MinY, g.MaxY, g.playerIds)
}
