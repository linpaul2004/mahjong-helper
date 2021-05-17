package model

import "fmt"

type PlayerInfo struct {
	HandTiles34 []int  // 手牌，不含副露
	Melds       []Meld // 副露
	DoraTiles   []int  // 寶牌指示牌産生的寶牌，可以重複
	NumRedFives []int  // 按照 mps 的順序，各個赤5的個數（手牌和副露中的）

	IsTsumo       bool // 是否自摸
	WinTile       int  // 自摸/榮和的牌
	RoundWindTile int  // 場風
	SelfWindTile  int  // 自風
	IsParent      bool // 是否為親家
	IsDaburii     bool // 是否雙立直
	IsRiichi      bool // 是否立直

	DiscardTiles []int // 自家捨牌，用於判斷和率，是否振聽等  *注意創建 PlayerInfo 的時候把負數調整成正的！
	LeftTiles34  []int // 剩餘牌

	LeftDrawTilesCount int // 剩餘可以摸的牌數

	//LeftRedFives []int // 剩餘赤5個數，用於估算打點
	//AvgUraDora float64 // 平均裏寶牌個數，用於計算立直時的打點

	NukiDoraNum int // 拔北寶牌數
}

func NewSimplePlayerInfo(tiles34 []int, melds []Meld) *PlayerInfo {
	leftTiles34 := InitLeftTiles34WithTiles34(tiles34)
	for _, meld := range melds {
		for _, tile := range meld.Tiles {
			leftTiles34[tile]--
			if leftTiles34[tile] < 0 {
				panic(fmt.Sprint("副露數據不合法", melds))
			}
		}
	}
	return &PlayerInfo{
		HandTiles34:   tiles34,
		Melds:         melds,
		NumRedFives:   make([]int, 3),
		RoundWindTile: 27,
		SelfWindTile:  27,
		LeftTiles34:   leftTiles34,
	}
}

// 根據手牌、副露、赤5，結合哪些是寶牌，計算出擁有的寶牌個數
func (pi *PlayerInfo) CountDora() (count int) {
	for _, doraTile := range pi.DoraTiles {
		count += pi.HandTiles34[doraTile]
		for _, m := range pi.Melds {
			for _, tile := range m.Tiles {
				if tile == doraTile {
					count++
				}
			}
		}
	}
	// 手牌和副露中的赤5
	for _, num := range pi.NumRedFives {
		count += num
	}
	// 拔北寶牌
	if pi.NukiDoraNum > 0 {
		count += pi.NukiDoraNum
		// 特殊：西為指示牌
		for _, doraTile := range pi.DoraTiles {
			if doraTile == 30 {
				count += pi.NukiDoraNum
			}
		}
	}
	return
}

// 立直時，根據牌山計算和了時的裏寶牌個數
// TODO: 考慮 WinTile
//func (pi *PlayerInfo) CountUraDora() (count float64) {
//	if !pi.IsRiichi || pi.IsNaki() {
//		return 0
//	}
//	uraDoraTileLeft := make([]int, len(pi.LeftTiles34))
//	for tile, left := range pi.LeftTiles34 {
//		uraDoraTileLeft[DoraTile(tile)] = left
//	}
//	sum := 0
//	weight := 0
//	for tile, c := range pi.HandTiles34 {
//		w := uraDoraTileLeft[tile]
//		sum += w * c
//		weight += w
//	}
//	for _, meld := range pi.Melds {
//		for tile, c := range meld.Tiles {
//			w := uraDoraTileLeft[tile]
//			sum += w * c
//			weight += w
//		}
//	}
//	// 簡化計算，直接乘上寶牌指示牌的個數
//	return float64(len(pi.DoraTiles)*sum) / float64(weight)
//}

// 是否已鳴牌（暗槓不算）
// 可以用來判斷該玩家能否立直，計算門清加符、役種番數等
func (pi *PlayerInfo) IsNaki() bool {
	for _, meld := range pi.Melds {
		if meld.MeldType != MeldTypeAnkan {
			return true
		}
	}
	return false
}

// 是否振聽
// 僅限聽牌時調用
// TODO: Waits 移進來
func (pi *PlayerInfo) IsFuriten(waits map[int]int) bool {
	for _, discardTile := range pi.DiscardTiles {
		if _, ok := waits[discardTile]; ok {
			return true
		}
	}
	return false
}

/************* 以下接口暫為內部調用 ************/

func (pi *PlayerInfo) FillLeftTiles34() {
	pi.LeftTiles34 = InitLeftTiles34WithTiles34(pi.HandTiles34)
}

// 手上的這種牌只有赤5
func (pi *PlayerInfo) IsOnlyRedFive(tile int) bool {
	return tile < 27 && tile%9 == 4 && pi.HandTiles34[tile] > 0 && pi.HandTiles34[tile] == pi.NumRedFives[tile/9]
}

func (pi *PlayerInfo) DiscardTile(tile int, isRedFive bool) {
	// 從手牌中捨去一張牌到牌河
	pi.HandTiles34[tile]--
	if isRedFive {
		pi.NumRedFives[tile/9]--
	}
	pi.DiscardTiles = append(pi.DiscardTiles, tile)
}

func (pi *PlayerInfo) UndoDiscardTile(tile int, isRedFive bool) {
	// 複原從手牌中捨去一張牌到牌河的動作，即把這張牌從牌河移回手牌
	pi.DiscardTiles = pi.DiscardTiles[:len(pi.DiscardTiles)-1]
	pi.HandTiles34[tile]++
	if isRedFive {
		pi.NumRedFives[tile/9]++
	}
}

//func (pi *PlayerInfo) DrawTile(tile int) {
//	// 從牌山中摸牌
//}
//
//func (pi *PlayerInfo) UndoDrawTile(tile int) {
//	// 複原從牌山中摸牌的動作，即把這張牌放回牌山
//}

func (pi *PlayerInfo) AddMeld(meld Meld) {
	// 用手牌中的牌去鳴牌
	// 原有的寶牌數量並未發生變化
	for _, tile := range meld.SelfTiles {
		pi.HandTiles34[tile]--
	}
	pi.Melds = append(pi.Melds, meld)
	if meld.RedFiveFromOthers {
		tile := meld.Tiles[0]
		pi.NumRedFives[tile/9]++
	}
}

func (pi *PlayerInfo) UndoAddMeld() {
	// 複原鳴牌動作
	latestMeld := pi.Melds[len(pi.Melds)-1]
	for _, tile := range latestMeld.SelfTiles {
		pi.HandTiles34[tile]++
	}
	pi.Melds = pi.Melds[:len(pi.Melds)-1]
	if latestMeld.RedFiveFromOthers {
		tile := latestMeld.Tiles[0]
		pi.NumRedFives[tile/9]--
	}
}
