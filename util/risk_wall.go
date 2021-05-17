package util

import "sort"

const (
	WallSafeTypeDoubleNoChance = iota // 只輸單騎對碰
	WallSafeTypeNoChance              // 單騎對碰邊張坎張
	WallSafeTypeDoubleOneChance
	WallSafeTypeMixedOneChance  // 對於456來說，一半 double，一半不是 double
	WallSafeTypeOneChance
)

type WallSafeTile struct {
	Tile34   int
	SafeType int
}

type WallSafeTileList []WallSafeTile

func (l WallSafeTileList) String() string {
	tiles := []int{}
	for _, t := range l {
		tiles = append(tiles, t.Tile34)
	}
	return TilesToStr(tiles)
}

func (l WallSafeTileList) sort() {
	normalIndex := func(tile34 int) int {
		idx := tile34 % 9
		if idx >= 5 {
			// 5678 -> 3210
			idx = 8 - idx
		}
		return idx
	}

	sort.Slice(l, func(i, j int) bool {
		li, lj := l[i], l[j]

		liIndex := normalIndex(li.Tile34)
		ljIndex := normalIndex(lj.Tile34)
		// 先判斷345
		if liIndex > 2 && ljIndex <= 2 || liIndex <= 2 && ljIndex > 2 {
			return liIndex < ljIndex
		}

		if li.SafeType != lj.SafeType {
			return li.SafeType < lj.SafeType
		}

		return liIndex < ljIndex
	})
}

func (l WallSafeTileList) FilterWithHands(handsTiles34 []int) WallSafeTileList {
	newSafeTiles34 := WallSafeTileList{}
	for _, safeTile := range l {
		if handsTiles34[safeTile.Tile34] > 0 {
			newSafeTiles34 = append(newSafeTiles34, safeTile)
		}
	}
	newSafeTiles34.sort()
	return newSafeTiles34
}

// 根據剩餘牌 leftTiles34 中的某些牌是否為 0（壁），來判斷哪些牌較為安全（Double No Chance：只輸單騎、雙碰）
func CalcDNCSafeTiles(leftTiles34 []int) (dncSafeTiles WallSafeTileList) {
	nc := func(idx int) bool {
		return leftTiles34[idx] == 0
	}
	or := func(idx ...int) bool {
		for _, i := range idx {
			if nc(i) {
				return true
			}
		}
		return false
	}
	and := func(idx ...int) bool {
		for _, i := range idx {
			if !nc(i) {
				return false
			}
		}
		return true
	}

	const safeType = WallSafeTypeDoubleNoChance
	for i := 0; i < 3; i++ {
		// 2/3斷的1
		if or(9*i+1, 9*i+2) {
			dncSafeTiles = append(dncSafeTiles, WallSafeTile{9 * i, safeType})
		}
		// 3/14斷的2
		if nc(9*i+2) || and(9*i, 9*i+3) {
			dncSafeTiles = append(dncSafeTiles, WallSafeTile{9*i + 1, safeType})
		}
		// 14/24/25斷的3（4567同理）
		for j := 2; j <= 6; j++ {
			idx := 9*i + j
			if and(idx-2, idx+1) || and(idx-1, idx+1) || and(idx-1, idx+2) {
				dncSafeTiles = append(dncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
		// 7/69斷的8
		if nc(9*i+6) || and(9*i+5, 9*i+8) {
			dncSafeTiles = append(dncSafeTiles, WallSafeTile{9*i + 7, safeType})
		}
		// 7/8斷的9
		if or(9*i+6, 9*i+7) {
			dncSafeTiles = append(dncSafeTiles, WallSafeTile{9*i + 8, safeType})
		}
	}
	dncSafeTiles.sort()
	return
}

// 根據剩餘牌 leftTiles34 中的某些牌是否為 0（壁），來判斷哪些牌較為安全（Double No Chance：只輸單騎、雙碰）
// 這裏加上現物，相比 CalcDNCSafeTiles 可以得到更加精確的結果
// 註：雖然說在 4 為現物的情況下，1 也可以認為是只輸單騎、雙碰的，但這不在壁的討論範圍內，故不考慮這種情況
func CalcDNCSafeTilesWithDiscards(leftTiles34 []int, safeTiles34 []bool) (dncSafeTiles WallSafeTileList) {
	nc := func(idx int) bool {
		return leftTiles34[idx] == 0
	}

	const safeType = WallSafeTypeDoubleNoChance

	dncSafeTiles = CalcDNCSafeTiles(leftTiles34)

	// 在相鄰一側牌為壁的情況下，檢查另一側是否有現物筋牌
	// 例如 3，相鄰的 2 為壁且 6 為現物，則其為 DNC。其他的 2~8 同理（456 要判斷左側或右側，滿足一種即為 DNC）
	for i := 0; i < 3; i++ {
		for j := 1; j < 3; j++ {
			idx := 9*i + j
			if nc(idx-1) && safeTiles34[idx+3] {
				dncSafeTiles = append(dncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			if nc(idx-1) && safeTiles34[idx+3] || nc(idx+1) && safeTiles34[idx-3] {
				dncSafeTiles = append(dncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
		for j := 6; j < 8; j++ {
			idx := 9*i + j
			if nc(idx+1) && safeTiles34[idx-3] {
				dncSafeTiles = append(dncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
	}

	dncSafeTiles.sort()
	return
}

// 根據剩餘牌 leftTiles34 中的某些牌是否為 0（壁），來判斷哪些牌較為安全（No Chance：不輸兩面）
func CalcNCSafeTiles(leftTiles34 []int) (ncSafeTiles WallSafeTileList) {
	nc := func(idx int) bool {
		return leftTiles34[idx] == 0
	}
	or := func(idx ...int) bool {
		for _, i := range idx {
			if nc(i) {
				return true
			}
		}
		return false
	}

	const safeType = WallSafeTypeNoChance
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			idx := 9*i + j
			if or(idx+1, idx+2) {
				ncSafeTiles = append(ncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			if or(idx-2, idx-1) && or(idx+1, idx+2) {
				ncSafeTiles = append(ncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
		for j := 6; j < 9; j++ {
			idx := 9*i + j
			if or(idx-2, idx-1) {
				ncSafeTiles = append(ncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
	}
	ncSafeTiles.sort()
	return
}

// 根據剩餘牌 leftTiles34 中的某些牌是否為 1（薄壁），來判斷哪些牌較為安全（One Chance：早巡大概率不輸兩面）
func CalcOCSafeTiles(leftTiles34 []int) (ocSafeTiles WallSafeTileList) {
	oc := func(idx int) bool {
		return leftTiles34[idx] == 1
	}
	or := func(idx ...int) bool {
		for _, i := range idx {
			if oc(i) {
				return true
			}
		}
		return false
	}
	and := func(idx ...int) bool {
		for _, i := range idx {
			if !oc(i) {
				return false
			}
		}
		return true
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			idx := 9*i + j
			if and(idx+1, idx+2) {
				ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeDoubleOneChance})
			} else if or(idx+1, idx+2) {
				ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeOneChance})
			}
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			if or(idx-2, idx-1) && or(idx+1, idx+2) {
				if and(idx-2, idx-1, idx+1, idx+2) {
					// 兩邊都是 double
					ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeDoubleOneChance})
				} else if and(idx-2, idx-1) || and(idx+1, idx+2) {
					// 一半 double，一半不是 double
					ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeMixedOneChance})
				} else {
					ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeOneChance})
				}
			}
		}
		for j := 6; j < 9; j++ {
			idx := 9*i + j
			if and(idx-2, idx-1) {
				ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeDoubleOneChance})
			} else if or(idx-2, idx-1) {
				ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeOneChance})
			}
		}
	}
	ocSafeTiles.sort()
	return
}

func CalcWallTiles(leftTiles34 []int) (safeTiles WallSafeTileList) {
	safeTiles = append(safeTiles, CalcNCSafeTiles(leftTiles34)...)
	safeTiles = append(safeTiles, CalcOCSafeTiles(leftTiles34)...)
	safeTiles.sort()
	return
}
