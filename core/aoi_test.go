package core

import (
	"fmt"
	"testing"
)

func TestNewAOIManager(t *testing.T) {
	mgr := NewAOIManager(100, 300, 4, 200, 450, 5)
	fmt.Println(mgr)
}

func TestAOIManager_GetSurroundGridsByGid(t *testing.T) {
	mgr := NewAOIManager(0, 250, 5, 0, 250, 5)
	for gid, _ := range mgr.grids {
		// 获得当前格子周围的九宫格
		grids := mgr.GetSurroundGridsByGid(gid)
		// 得到九宫格内的所有ids
		fmt.Println("gid: ", gid, " grids len: ", len(grids))

		gids := make([]int, 0, len(grids))
		for _, grid := range grids {
			gids = append(gids, grid.GID)
		}
		fmt.Printf("grid id: %d, surrounding grid ids are %v\n", gid, gids)
	}

}
