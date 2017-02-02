package main

import (
	"sync"
)

type RankingManager struct {
	launchedMutex *sync.RWMutex
	launched      map[int64]int
}

func NewRankingManager() *RankingManager {
	return new(RankingManager)
}

// New 成功時に接続先ポート番号を返す
func (rm *RankingManager) Launch(cid int64) (int, error) {
	rm.launchedMutex.RLock()
	p, ok := rm.launched[cid]
	rm.launchedMutex.RUnlock()

	if ok {
		return p, nil
	} else {

	}

	return 0, nil
}
