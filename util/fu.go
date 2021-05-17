package util

import (
	"github.com/EndlessCheng/mahjong-helper/util/model"
)

func roundUpFu(fu int) int {
	return ((fu-1)/10 + 1) * 10
}

// 根據手牌拆解結果，結合場況計算符數
func (hi *_handInfo) calcFu(isNaki bool) int {
	divideResult := hi.divideResult

	// 特殊：七對子計 25 符
	if divideResult.IsChiitoi {
		return 25
	}

	const baseFu = 20

	// 符底 20 符
	fu := baseFu

	// 暗刻加符
	_, ronKotsu := hi.numAnkou()
	for _, tile := range divideResult.KotsuTiles {
		var _fu int
		// 榮和刻子算明刻
		if ronKotsu && tile == hi.WinTile {
			_fu = 2
		} else {
			_fu = 4
		}
		if isYaochupai(tile) {
			_fu *= 2
		}
		fu += _fu
	}

	// 明刻、明槓、暗槓加符
	for _, meld := range hi.Melds {
		_fu := 0
		switch meld.MeldType {
		case model.MeldTypePon:
			_fu = 2
		case model.MeldTypeMinkan, model.MeldTypeKakan:
			_fu = 8
		case model.MeldTypeAnkan:
			_fu = 16
		}
		if _fu > 0 {
			if isYaochupai(meld.Tiles[0]) {
				_fu *= 2
			}
			fu += _fu
		}
	}

	// 雀頭加符（連風雀頭計 4 符）
	if hi.isYakuTile(divideResult.PairTile) {
		fu += 2
		if hi.isDoubleWindTile(divideResult.PairTile) {
			fu += 2
		}
	}

	if fu == baseFu {
		// 手牌全是順子，且雀頭不是役牌
		if isNaki {
			// 無論怎樣都不可能超過 30 符，直接返回
			return 30
		}
		// 門清狀態下需要檢測能否平和
		// 若沒有平和則一定是坎張、邊張、單騎和牌
		isPinfu := false
		for _, tile := range divideResult.ShuntsuFirstTiles {
			t9 := tile % 9
			if t9 < 6 && tile == hi.WinTile || t9 > 0 && tile+2 == hi.WinTile {
				isPinfu = true
				break
			}
		}
		if hi.IsTsumo {
			if isPinfu {
				// 門清自摸平和 20 符
				return 20
			}
			// 坎張、邊張、單騎自摸，30 符
			return 30
		} else {
			// 榮和
			if isPinfu {
				// 門清平和榮和 30 符
				return 30
			}
			// 坎張、邊張、單騎榮和，40 符
			return 40
		}
	}

	// 門清榮和加符
	if !isNaki && !hi.IsTsumo {
		fu += 10
	}

	// 自摸加符
	if hi.IsTsumo {
		fu += 2
	}

	// 邊張、坎張、單騎和牌加符
	// 考慮能否不為兩面和牌
	if divideResult.PairTile == hi.WinTile {
		fu += 2 // 單騎和牌加符
	} else {
		for _, tile := range divideResult.ShuntsuFirstTiles {
			if tile+1 == hi.WinTile {
				fu += 2 // 坎張和牌加符
				break
			}
			if tile%9 == 0 && tile+2 == hi.WinTile || tile%9 == 6 && tile == hi.WinTile {
				fu += 2 // 邊張和牌加符
				break
			}
		}
	}

	// 進位
	return roundUpFu(fu)
}
