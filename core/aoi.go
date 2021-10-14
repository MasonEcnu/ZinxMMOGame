package core

import (
	"fmt"
)

// AOIManager AOI管理模块
type AOIManager struct {
	MinX  int           // 区域左边界坐标
	MaxX  int           // 区域右边界坐标
	CntsX int           // x方向格子数量
	MinY  int           // 区域上边界坐标
	MaxY  int           // 区域下边界坐标
	CntsY int           // y方向格子数量
	grids map[int]*Grid // 当前区域中的格子
}

func NewAOIManager(minx, maxx, cntsx, miny, maxy, cntsy int) *AOIManager {
	aoiMgr := &AOIManager{
		MinX:  minx,
		MaxX:  maxx,
		CntsX: cntsx,
		MinY:  miny,
		MaxY:  maxy,
		CntsY: cntsy,
		grids: make(map[int]*Grid),
	}

	// 给AOI区域初始化所有的格子
	for y := 0; y < cntsy; y++ {
		for x := 0; x < cntsx; x++ {
			// 利用格子坐标计算格子id
			// id = idy * nx + idx
			gid := y*cntsx + x
			aoiMgr.grids[gid] = NewGrid(gid,
				aoiMgr.MinX+x*aoiMgr.gridWidth(),
				aoiMgr.MinX+(x+1)*aoiMgr.gridWidth(),
				aoiMgr.MinY+y*aoiMgr.gridLength(),
				aoiMgr.MinY+(y+1)*aoiMgr.gridLength())
		}
	}
	return aoiMgr
}

// GetSurroundGridsByGid 根据格子id得到周边的九宫格内格子的信息
func (mgr *AOIManager) GetSurroundGridsByGid(gid int) (grids []*Grid) {
	// 判断id是否存在
	if _, ok := mgr.grids[gid]; !ok {
		return
	}

	// 将当前grid添加到九宫格中
	grids = append(grids, mgr.grids[gid])

	// 根据gid得到当前格子所在的x轴编号
	idx := gid % mgr.CntsX
	// 判断当前idx左边是否还有格子
	if idx > 0 {
		grids = append(grids, mgr.grids[gid-1])
	}

	// 判断当前idx右边是否还有格子
	if idx < mgr.CntsX-1 {
		grids = append(grids, mgr.grids[gid+1])
	}

	// 将x轴当前的格子都取出，进行遍历，分别判断每个格子上下是否有格子

	// 取出当前x轴都id集合
	gridsX := make([]int, 0, len(grids))
	for _, g := range grids {
		gridsX = append(gridsX, g.GID)
	}

	// 遍历x轴格子
	for _, xgid := range gridsX {
		// 计算该格子处于第几列
		idy := xgid / mgr.CntsY
		// 判断当前idy上边是否有格子
		if idy > 0 {
			grids = append(grids, mgr.grids[xgid-mgr.CntsX])
		}
		// 判读当前idy下边是否有格子
		if idy < mgr.CntsY-1 {
			grids = append(grids, mgr.grids[xgid+mgr.CntsX])
		}
	}
	return
}

// GetGidByPos 通过坐标获取对应格子id
func (mgr *AOIManager) GetGidByPos(x, y float32) int {
	gx := (int(x) - mgr.MinX) / mgr.gridWidth()
	gy := (int(y) - mgr.MinY) / mgr.gridLength()
	return gy*mgr.CntsX + gx
}

// GetPlayerIdsByPos 通过坐标获取周围九宫格内的全部playerIds
func (mgr *AOIManager) GetPlayerIdsByPos(x, y float32) (playerIds []int) {
	// 根据坐标获取格子id
	gid := mgr.GetGidByPos(x, y)
	// 根据格子id获取周围九宫格信息
	grids := mgr.GetSurroundGridsByGid(gid)
	// 打包
	for _, g := range grids {
		playerIds = append(playerIds, g.GetPlayerIds()...)
		fmt.Printf("===> grid ID : %d, playerIds : %v  ====\n", g.GID, g.GetPlayerIds())
	}
	return
}

// GetPlayerIdsByGid 通过gid获取指定格子内的全部playerIds
func (mgr *AOIManager) GetPlayerIdsByGid(gid int) (playerIds []int) {
	if grid, ok := mgr.grids[gid]; ok {
		playerIds = append(playerIds, grid.GetPlayerIds()...)
	}
	return
}

// RemovePlayerIdFromGrid 移除一个格子中的PlayerID
func (mgr *AOIManager) RemovePlayerIdFromGrid(playerId, gid int) {
	if grid, ok := mgr.grids[gid]; ok {
		grid.Remove(playerId)
	}
}

// AddPlayerIdToGrid 添加一个PlayerID到一个格子中
func (mgr *AOIManager) AddPlayerIdToGrid(playerId, gid int) {
	if grid, ok := mgr.grids[gid]; ok {
		grid.Add(playerId)
	}
}

// AddPlayerIdToGridByPos 通过横纵坐标添加一个Player到一个格子中
func (mgr *AOIManager) AddPlayerIdToGridByPos(playerId int, x, y float32) {
	gid := mgr.GetGidByPos(x, y)
	if grid, ok := mgr.grids[gid]; ok {
		grid.Add(playerId)
	}
}

// RemoveFromGridByPos 通过横纵坐标把一个Player从对应的格子中删除
func (mgr *AOIManager) RemoveFromGridByPos(playerId int, x, y float32) {
	gid := mgr.GetGidByPos(x, y)
	if grid, ok := mgr.grids[gid]; ok {
		grid.Remove(playerId)
	}
}

// gridWidth 每个格子在x轴方向的宽度
func (mgr *AOIManager) gridWidth() int {
	return (mgr.MaxX - mgr.MinX) / mgr.CntsX
}

// gridLength 每个格子在y轴方向的长度
func (mgr *AOIManager) gridLength() int {
	return (mgr.MaxY - mgr.MinY) / mgr.CntsY
}

// String AOIManager 结构消息
func (mgr *AOIManager) String() string {
	s := fmt.Sprintf("AOIManagr:\nminX:%d, maxX:%d, cntsX:%d, minY:%d, maxY:%d, cntsY:%d\n Grids in AOI Manager:\n",
		mgr.MinX, mgr.MaxX, mgr.CntsX, mgr.MinY, mgr.MaxY, mgr.CntsY)
	for _, grid := range mgr.grids {
		s += fmt.Sprintln(grid)
	}
	return s
}
