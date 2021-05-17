package util

import "github.com/EndlessCheng/mahjong-helper/util/model"

type _handInfo struct {
	*model.PlayerInfo
	divideResult *DivideResult // 手牌解析結果

	// *在計算役種前，緩存自己的順子牌和刻子牌，這樣能減少大量重複計算
	allShuntsuFirstTiles []int
	allKotsuTiles        []int
}

// 未排序。用於算一通、三色
func (hi *_handInfo) getAllShuntsuFirstTiles() []int {
	shuntsuFirstTiles := append([]int{}, hi.divideResult.ShuntsuFirstTiles...)
	for _, meld := range hi.Melds {
		if meld.MeldType == model.MeldTypeChi {
			shuntsuFirstTiles = append(shuntsuFirstTiles, meld.Tiles[0])
		}
	}
	return shuntsuFirstTiles
}

// 未排序。用於算對對、三色同刻
func (hi *_handInfo) getAllKotsuTiles() []int {
	kotsuTiles := append([]int{}, hi.divideResult.KotsuTiles...)
	for _, meld := range hi.Melds {
		if meld.MeldType != model.MeldTypeChi {
			kotsuTiles = append(kotsuTiles, meld.Tiles[0])
		}
	}
	return kotsuTiles
}

// 是否包含字牌（調用前需要設置刻子牌）
func (hi *_handInfo) containHonor() bool {
	// 七對子特殊處理
	if hi.divideResult.IsChiitoi {
		for _, c := range hi.HandTiles34[27:] {
			if c > 0 {
				return true
			}
		}
		return false
	}

	if hi.divideResult.PairTile >= 27 {
		return true
	}
	for _, tile := range hi.allKotsuTiles {
		if tile >= 27 {
			return true
		}
	}
	return false
}

// 是否為役牌，用於算役種（役牌、平和）、雀頭加符
func (hi *_handInfo) isYakuTile(tile int) bool {
	return tile >= 31 || tile == hi.RoundWindTile || tile == hi.SelfWindTile
}

// 是否為連風牌
func (hi *_handInfo) isDoubleWindTile(tile int) bool {
	return hi.RoundWindTile == hi.SelfWindTile && tile == hi.RoundWindTile
}

// 暗槓個數，用於算三暗刻、四暗刻
func (hi *_handInfo) numAnkan() (cnt int) {
	for _, meld := range hi.Melds {
		if meld.MeldType == model.MeldTypeAnkan {
			cnt++
		}
	}
	return
}

// 槓子個數，用於算三槓子、四槓子
func (hi *_handInfo) numKantsu() (cnt int) {
	for _, meld := range hi.Melds {
		if meld.IsKan() {
			cnt++
		}
	}
	return
}

// 暗刻個數，用於算三暗刻、四暗刻、符數（如 456666 榮和 6，這裏算一個暗刻）
// 即手牌暗刻和暗槓
func (hi *_handInfo) numAnkou() (cnt int, isMinkou bool) {
	num := len(hi.divideResult.KotsuTiles) + hi.numAnkan()
	// 自摸直接返回，無需討論是否榮和了刻子
	if hi.IsTsumo {
		return num, false
	}
	// 榮和的牌在雀頭裏
	if hi.WinTile == hi.divideResult.PairTile {
		return num, false
	}
	// 榮和的牌在順子裏
	for _, tile := range hi.divideResult.ShuntsuFirstTiles {
		if hi.WinTile >= tile && hi.WinTile <= tile+2 {
			return num, false
		}
	}
	// 榮和的牌在刻子裏，該刻子算明刻
	return num - 1, true
}

// 計算在指定牌中的刻子個數
func (hi *_handInfo) _countSpecialKotsu(specialTilesL, specialTilesLR int) (cnt int) {
	for _, tile := range hi.allKotsuTiles {
		if tile >= specialTilesL && tile <= specialTilesLR {
			cnt++
		}
	}
	return
}
