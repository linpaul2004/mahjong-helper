package util

import "github.com/EndlessCheng/mahjong-helper/util/model"

// 沒有立直時，根據玩家的副露、手切來判斷其聽牌率 (0-100)
// TODO: 傳入 *model.PlayerInfo
func CalcTenpaiRate(melds []*model.Meld, discardTiles []int, meldDiscardsAt []int) float64 {
	isNaki := false
	for _, meld := range melds {
		if meld.MeldType != model.MeldTypeAnkan {
			isNaki = true
		}
	}

	if !isNaki {
		// 默聽聽牌率近似為巡目數
		turn := len(discardTiles)
		return float64(turn)
	}

	if len(melds) == 4 {
		return 100
	}

	_tenpaiRate := tenpaiRate[len(melds)]

	turn := MinInt(len(discardTiles), len(_tenpaiRate)-1)
	_tenpaiRateWithTurn := _tenpaiRate[turn]

	// 計算上一次副露後的手切數
	// 注意連續開槓時，副露數 len(melds) 是不等於副露時的切牌數 len(meldDiscardsAt) 的
	countTedashi := 0
	if len(meldDiscardsAt) > 0 {
		latestDiscardAt := meldDiscardsAt[len(meldDiscardsAt)-1]
		if len(discardTiles) > latestDiscardAt {
			for _, disTile := range discardTiles[latestDiscardAt+1:] {
				if disTile >= 0 {
					countTedashi++
				}
			}
		}
	}
	countTedashi = MinInt(countTedashi, len(_tenpaiRateWithTurn)-1)

	return _tenpaiRateWithTurn[countTedashi]
}
