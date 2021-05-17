package util

import (
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"fmt"
)

type shantenSearchNode13 struct {
	shanten  int
	waits    Waits
	children map[int]*shantenSearchNode14 // 向聽前進的摸牌-node14
}

func (n *shantenSearchNode13) printWithPrefix(prefix string) string {
	output := ""
	for drawTile, node14 := range n.children {
		output += prefix + fmt.Sprintln("摸", Mahjong[drawTile]) + node14.printWithPrefix(prefix+"  ")
	}
	return output
}

func (n *shantenSearchNode13) String() string {
	return n.printWithPrefix("")
}

type shantenSearchNode14 struct {
	shanten  int
	children map[int]*shantenSearchNode13 // 向聽不變的捨牌-node13
}

func (n *shantenSearchNode14) printWithPrefix(prefix string) string {
	if n == nil || n.shanten == shantenStateAgari {
		return prefix + "end\n"
	}
	output := ""
	for discardTile, node13 := range n.children {
		output += prefix + fmt.Sprintln("捨", Mahjong[discardTile]) + node13.printWithPrefix(prefix+"  ")
	}
	return output
}

func (n *shantenSearchNode14) String() string {
	return n.printWithPrefix("")
}

func _search13(currentShanten int, playerInfo *model.PlayerInfo, stopAtShanten int) *shantenSearchNode13 {
	waits := Waits{}
	children := map[int]*shantenSearchNode14{}
	tiles34 := playerInfo.HandTiles34
	leftTiles34 := playerInfo.LeftTiles34

	isTenpai := currentShanten == shantenStateTenpai

	// 剪枝：檢測浮牌，在不考慮國士無雙的情況下，這種牌是不可能讓向聽數前進的（但有改良的可能，不過本函數不考慮這個）
	//if !isTenpai {
	//	needCheck34 := make([]bool, 34)
	//	idx := -1
	//	for i := 0; i < 3; i++ {
	//		for j := 0; j < 9; j++ {
	//			idx++
	//			if tiles34[idx] == 0 {
	//				continue
	//			}
	//			if j == 0 {
	//				needCheck34[idx] = true
	//				needCheck34[idx+1] = true
	//				needCheck34[idx+2] = true
	//			} else if j == 1 {
	//				needCheck34[idx-1] = true
	//				needCheck34[idx] = true
	//				needCheck34[idx+1] = true
	//				needCheck34[idx+2] = true
	//			} else if j < 7 {
	//				needCheck34[idx-2] = true
	//				needCheck34[idx-1] = true
	//				needCheck34[idx] = true
	//				needCheck34[idx+1] = true
	//				needCheck34[idx+2] = true
	//			} else if j == 7 {
	//				needCheck34[idx-2] = true
	//				needCheck34[idx-1] = true
	//				needCheck34[idx] = true
	//				needCheck34[idx+1] = true
	//			} else {
	//				needCheck34[idx-2] = true
	//				needCheck34[idx-1] = true
	//				needCheck34[idx] = true
	//			}
	//		}
	//	}
	//	for i := 27; i < 34; i++ {
	//		if tiles34[i] > 0 {
	//			needCheck34[i] = true
	//		}
	//	}
	//}

	for i := 0; i < 34; i++ {
		//if !needCheck34[i] {
		//	continue
		//}
		if tiles34[i] == 4 {
			continue
		}
		tiles34[i]++
		if isTenpai {
			// 優化：聽牌時改用更為快速的 IsAgari
			if IsAgari(tiles34) {
				waits[i] = leftTiles34[i]
				children[i] = nil
			}
		} else {
			if CalculateShanten(tiles34) < currentShanten {
				// 向聽前進了，則換的這張牌為進張，進張數即剩餘枚數
				// 有可能為 0，但考慮到判斷振聽時需要進張種類，所以記錄
				waits[i] = leftTiles34[i]
				if leftTiles34[i] > 0 && currentShanten-1 >= stopAtShanten {
					leftTiles34[i]--
					children[i] = _search14(currentShanten-1, playerInfo, stopAtShanten)
					leftTiles34[i]++
				} else {
					children[i] = nil
				}
			}
		}
		tiles34[i]--
	}

	return &shantenSearchNode13{
		shanten:  currentShanten,
		waits:    waits,
		children: children,
	}
}

// 技巧：傳入的 targetShanten 若為當前手牌的向聽+1，則為向聽倒退
func _search14(targetShanten int, playerInfo *model.PlayerInfo, stopAtShanten int) *shantenSearchNode14 {
	// 不需要判斷 targetShanten 是否為 shantenStateAgari：因為_search13 中用的是 IsAgari，所以 targetShanten 是 >=0 的
	children := map[int]*shantenSearchNode13{}
	tiles34 := playerInfo.HandTiles34
	for i := 0; i < 34; i++ {
		if tiles34[i] == 0 {
			continue
		}
		tiles34[i]--
		if CalculateShanten(tiles34) == targetShanten {
			// 向聽不變，捨牌正確
			children[i] = _search13(targetShanten, playerInfo, stopAtShanten)
		}
		tiles34[i]++
	}

	return &shantenSearchNode14{
		shanten:  targetShanten,
		children: children,
	}
}

// 3k+1 張牌，計算向聽數、進張（考慮了剩餘枚數），不計算改良
func CalculateShantenAndWaits13(tiles34 []int, leftTiles34 []int) (shanten int, waits Waits) {
	if len(leftTiles34) == 0 {
		leftTiles34 = InitLeftTiles34WithTiles34(tiles34)
	}

	shanten = CalculateShanten(tiles34)
	pi := &model.PlayerInfo{HandTiles34: tiles34, LeftTiles34: leftTiles34}
	node13 := _search13(shanten, pi, shanten) // 只搜索一層
	waits = node13.waits
	return
}

// 技巧：傳入的 shanten 若為當前手牌的向聽+1，則為向聽倒退
func searchShanten14(shanten int, playerInfo *model.PlayerInfo, stopAtShanten int) *shantenSearchNode14 {
	if shanten == shantenStateAgari {
		return &shantenSearchNode14{
			shanten:  shanten,
			children: map[int]*shantenSearchNode13{},
		}
	}
	return _search14(shanten, playerInfo, stopAtShanten)
}
