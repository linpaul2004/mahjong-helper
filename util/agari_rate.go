package util

import (
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"sort"
)

// 計算各張待牌的和率
// 剩餘為 0 則和率為 0
func CalculateAgariRateOfEachTile(waits Waits, playerInfo *model.PlayerInfo) map[int]float64 {
	if playerInfo == nil {
		playerInfo = &model.PlayerInfo{}
	}

	tileAgariRate := map[int]float64{}

	// 振聽的話和率簡化成和枚數相關
	if playerInfo.IsFuriten(waits) {
		for tile, left := range waits {
			rate := 0.0
			for i := 0; i < left; i++ {
				rate = rate + furitenBaseAgariRate - rate*furitenBaseAgariRate/100
			}
			tileAgariRate[tile] = rate
		}
		return tileAgariRate
	}

	// 特殊處理字牌單騎的情況
	if len(waits) == 1 {
		for tile, left := range waits {
			if tile >= 27 {
				rate := honorTileDankiAgariTable[left]
				if InInts(tile, playerInfo.DoraTiles) {
					// 調整聽寶牌時的和率
					// 忽略 dora 複合的影響
					rate *= honorDoraAgariMulti
				}
				tileAgariRate[tile] = rate
				return tileAgariRate
			}
		}
	}

	// 根據自家捨牌，確定各個牌的類型（無筋、半筋、筋、兩筋），從而得出不同的和率
	tileType27 := calcTileType27(playerInfo.DiscardTiles)
	for tile, left := range waits {
		var rate float64
		if tile < 27 { // 數牌
			rate = agariMap[tileType27[tile]][left]
		} else { // 字牌，非單騎
			rate = honorTileNonDankiAgariTable[left]
		}
		if InInts(tile, playerInfo.DoraTiles) {
			// 調整聽寶牌時的和率
			// 忽略 dora 複合的影響
			if tile >= 27 {
				rate *= honorDoraAgariMulti
			} else {
				rate *= numberDoraAgariMulti
			}
		}
		tileAgariRate[tile] = rate
	}

	return tileAgariRate
}

// 計算平均和率
func CalculateAvgAgariRate(waits Waits, playerInfo *model.PlayerInfo) float64 {
	if playerInfo == nil {
		playerInfo = &model.PlayerInfo{}
	}

	// 振聽的話和率簡化成和枚數相關
	if playerInfo.IsFuriten(waits) {
		rate := 0.0
		for i := 0; i < waits.AllCount(); i++ {
			rate = rate + furitenBaseAgariRate - rate*furitenBaseAgariRate/100
		}
		return rate
	}

	tileAgariRate := CalculateAgariRateOfEachTile(waits, playerInfo)
	agariRate := 0.0
	for _, rate := range tileAgariRate {
		agariRate = agariRate + rate - agariRate*rate/100
	}

	// 調整兩面和牌率
	// 需要 waits 恰好是筋牌關係，不能有非筋牌
	waitTiles := []int{}
	for tile, left := range waits {
		if left > 0 {
			if tile >= 27 {
				return agariRate
			}
			waitTiles = append(waitTiles, tile)
		}
	}
	if len(waitTiles) > 1 {
		suitType := waitTiles[0] / 9
		for _, tile := range waitTiles[1:] {
			if tile/9 != suitType {
				return agariRate
			}
		}
		sort.Ints(waitTiles)
		if len(waitTiles) == 2 && waitTiles[0]+3 == waitTiles[1] ||
			len(waitTiles) == 3 && waitTiles[0]+3 == waitTiles[1] && waitTiles[1]+3 == waitTiles[2] {
			agariRate *= ryanmenAgariMulti
		}
	}

	return agariRate
}
