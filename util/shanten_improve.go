package util

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"math"
	"sort"
)

// map[改良牌]進張（選擇進張數最大的）
type Improves map[int]Waits

// 3k+1 張手牌的分析結果
type Hand13AnalysisResult struct {
	// 原手牌
	Tiles34 []int

	// 剩餘牌
	LeftTiles34 []int

	// 是否已鳴牌（非門清狀態）
	// 用於判斷是否無役等
	IsNaki bool

	// 向聽數
	Shanten int

	// 進張
	// 考慮了剩餘枚數
	// 若某個進張牌 4 枚都可見，則該進張的 value 值為 0
	Waits Waits

	// 默聽時的進張
	DamaWaits Waits

	// TODO: 鳴牌進張：他家打出這張牌，可以鳴牌，且能讓向聽數前進
	//MeldWaits Waits

	// map[進張牌]向聽前進後的(最大)進張數
	NextShantenWaitsCountMap map[int]int

	// 向聽前進後的(最大)進張數的加權均值
	AvgNextShantenWaitsCount float64

	// 綜合了進張與向聽前進後進張的評分
	MixedWaitsScore float64

	// 改良：摸到這張牌雖不能讓向聽數前進，但可以讓進張變多
	// len(Improves) 即為改良的牌的種數
	Improves Improves

	// 改良情況數，這裏計算的是有多少種使進張增加的摸牌-切牌方式
	ImproveWayCount int

	// 摸到非進張牌時的進張數的加權均值（非改良+改良。對於非改良牌，其進張數為 Waits.AllCount()）
	// 這裏只考慮一巡的改良均值
	// TODO: 在考慮改良的情況下，如何計算向聽前進所需要的摸牌次數的期望值？蒙特卡羅方法？
	AvgImproveWaitsCount float64

	// 聽牌時的手牌和率
	// TODO: 未聽牌時的和率？
	AvgAgariRate float64

	// 振聽可能率（一向聽和聽牌時）
	FuritenRate float64

	// 役種
	YakuTypes map[int]struct{}

	// （鳴牌時）是否片聽
	IsPartWait bool

	// 寶牌個數（手牌+副露）
	DoraCount int

	// 非立直狀態下的打點期望（副露或默聽）
	DamaPoint float64

	// 立直狀態下的打點期望
	RiichiPoint float64

	// 局收支
	MixedRoundPoint float64

	// TODO: 赤牌改良提醒
}

// 進張和向聽前進後進張的評分
// 這裏粗略地近似為向聽前進兩次的概率
func (r *Hand13AnalysisResult) speedScore() float64 {
	if r.Waits.AllCount() == 0 || r.AvgNextShantenWaitsCount == 0 {
		return 0
	}
	leftCount := float64(CountOfTiles34(r.LeftTiles34))
	p2 := float64(r.Waits.AllCount()) / leftCount
	//p2 := r.AvgImproveWaitsCount / leftCount
	p1 := r.AvgNextShantenWaitsCount / leftCount
	//if r.AvgAgariRate > 0 { // TODO: 用和率需要考慮巡目
	//	p1 = r.AvgAgariRate / 100
	//}
	p2_, p1_ := 1-p2, 1-p1
	const leftTurns = 10.0 // math.Max(5.0, leftCount/4)
	sumP2 := p2_ * (1 - math.Pow(p2_, leftTurns)) / p2
	sumP1 := p1_ * (1 - math.Pow(p1_, leftTurns)) / p1
	result := p2 * p1 * (sumP2 - sumP1) / (p2_ - p1_)
	return result * 100
}

func (r *Hand13AnalysisResult) mixedRoundPoint() float64 {
	const weight = -1500
	if r.RiichiPoint > 0 {
		return r.AvgAgariRate/100*(r.RiichiPoint+1500) + weight
	}
	return r.AvgAgariRate/100*(r.DamaPoint+1500) + weight
}

// 調試用
func (r *Hand13AnalysisResult) String() string {
	s := fmt.Sprintf("%d 進張 %s\n%.2f 改良進張 [%d(%d) 種]",
		r.Waits.AllCount(),
		//r.Waits.AllCount()+r.MeldWaits.AllCount(),
		TilesToStrWithBracket(r.Waits.indexes()),
		r.AvgImproveWaitsCount,
		len(r.Improves),
		r.ImproveWayCount,
	)
	if len(r.DamaWaits) > 0 {
		s += fmt.Sprintf("（默聽進張 %s）", TilesToStrWithBracket(r.DamaWaits.indexes()))
	}
	if r.Shanten >= 1 {
		mixedScore := r.MixedWaitsScore
		//for i := 2; i <= r.Shanten; i++ {
		//	mixedScore /= 4
		//}
		s += fmt.Sprintf(" %.2f %s進張（%.2f 綜合分）",
			r.AvgNextShantenWaitsCount,
			NumberToChineseShanten(r.Shanten-1),
			mixedScore,
		)
	}
	if r.AvgAgariRate > 0 {
		s += fmt.Sprintf("[%.2f%% 和率] ", r.AvgAgariRate)
	}
	if r.MixedRoundPoint > 0 {
		s += fmt.Sprintf(" [局收支%d]", int(math.Round(r.MixedRoundPoint)))
	}
	if r.DamaPoint > 0 {
		s += fmt.Sprintf("[默聽%d]", int(math.Round(r.DamaPoint)))
	}
	if r.RiichiPoint > 0 {
		s += fmt.Sprintf("[立直%d]", int(math.Round(r.RiichiPoint)))
	}
	if r.Shanten >= 0 && r.Shanten <= 1 {
		if r.FuritenRate > 0 {
			if r.FuritenRate < 1 {
				s += "[可能振聽]"
			} else {
				s += "[振聽]"
			}
		}
	}
	if len(r.YakuTypes) > 0 {
		s += YakuTypesWithDoraToStr(r.YakuTypes, r.DoraCount)
	}
	return s
}

func (n *shantenSearchNode13) analysis(playerInfo *model.PlayerInfo, considerImprove bool) (result13 *Hand13AnalysisResult) {
	tiles34 := playerInfo.HandTiles34
	leftTiles34 := playerInfo.LeftTiles34
	shanten13 := n.shanten
	waits := n.waits
	waitsCount := waits.AllCount()

	nextShantenWaitsCountMap := map[int]int{} // map[進張牌]聽多少張牌
	improves := Improves{}
	improveWayCount := 0
	// 對於每張牌，摸到之後的手牌進張數（如果摸到的是 waits 中的牌，則進張數視作 waitsCount）
	maxImproveWaitsCount34 := make([]int, 34)
	for i := 0; i < 34; i++ {
		maxImproveWaitsCount34[i] = waitsCount // 初始化成基本進張
	}
	avgRoundPoint := 0.0
	roundPointWeight := 0
	yakuTypes := map[int]struct{}{}

	for i := 0; i < 34; i++ {
		// 從剩餘牌中摸牌
		if leftTiles34[i] == 0 {
			continue
		}
		leftTiles34[i]--
		tiles34[i]++

		if node14, ok := n.children[i]; ok && node14 != nil { // 摸到的是進張
			// 計算最大向聽前進後的進張
			maxNextShantenWaitsCount := 0
			for _, node13 := range node14.children {
				maxNextShantenWaitsCount = MaxInt(maxNextShantenWaitsCount, node13.waits.AllCount())
			}
			nextShantenWaitsCountMap[i] = maxNextShantenWaitsCount

			//const minRoundPoint = -1e10
			//maxRoundPoint := minRoundPoint

			if results14 := node14.analysis(playerInfo, false); len(results14) > 0 {
				bestResult14 := results14[0]

				// 加權：進張牌的剩餘枚數*局收支
				w := leftTiles34[i] + 1
				avgRoundPoint += float64(w) * bestResult14.Result13.MixedRoundPoint
				roundPointWeight += w

				// 添加役種
				for t := range bestResult14.Result13.YakuTypes {
					yakuTypes[t] = struct{}{}
				}
			}

			//for discardTile, node13 := range node14.children {
			//
			//
			//	// 切牌，然後分析 3k+1 張牌下的手牌情況
			//	// 若這張是5，在只有赤5的情況下才會切赤5（TODO: 考慮赤5騙37）
			//	_isRedFive := playerInfo.IsOnlyRedFive(discardTile)
			//	playerInfo.DiscardTile(discardTile, _isRedFive)
			//
			//	// 聽牌了
			//	if newShanten13 == 0 {
			//		// 聽牌一般切局收支最高的，這裏若為副露狀態用副露局收支，否則用立直局收支
			//		_avgAgariRate := CalculateAvgAgariRate(newWaits, playerInfo) / 100
			//		var _roundPoint float64
			//		if isNaki {
			//			// FIXME: 後付時，應該只計算役牌的和率
			//			_avgPoint, _ := CalcAvgPoint(*playerInfo, newWaits)
			//			if _avgPoint == 0 { // 無役
			//				_avgAgariRate = 0
			//			}
			//			_roundPoint = _avgAgariRate*(_avgPoint+1500) - 1500
			//		} else {
			//			_avgRiichiPoint, _ := CalcAvgRiichiPoint(*playerInfo, newWaits)
			//			_roundPoint = _avgAgariRate*(_avgRiichiPoint+1500) - 1500
			//		}
			//		maxRoundPoint = math.Max(maxRoundPoint, _roundPoint)
			//		// 計算可能的役種
			//		//fillYakuTypes(newShanten13, newWaits)
			//	}
			//
			//	playerInfo.UndoDiscardTile(discardTile, _isRedFive)
			//}
			//// 加權：進張牌的剩餘枚數*局收支
			//w := leftTiles34[i] + 1
			////avgAgariRate += maxAgariRate * float64(w)
			//if maxRoundPoint > minRoundPoint {
			//	avgRoundPoint += float64(w) * maxRoundPoint
			//	roundPointWeight += w
			//}
			//fmt.Println(i, maxAvgRiichiRonPoint)
			//avgRiichiPoint += maxAvgRiichiRonPoint * float64(w)
		} else if considerImprove { // 摸到的不是進張，但可能有改良
			for j := 0; j < 34; j++ {
				if tiles34[j] == 0 || j == i {
					continue
				}
				// 切牌，然後分析 3k+1 張牌下的手牌情況
				// 若這張是5，在只有赤5的情況下才會切赤5（TODO: 考慮赤5騙37）
				_isRedFive := playerInfo.IsOnlyRedFive(j)
				playerInfo.DiscardTile(j, _isRedFive)
				// 正確的切牌
				if newShanten13, improveWaits := CalculateShantenAndWaits13(tiles34, leftTiles34); newShanten13 == shanten13 {
					// 若進張數變多，則為改良
					// TODO: 若打點上升，也算改良
					if improveWaitsCount := improveWaits.AllCount(); improveWaitsCount > waitsCount {
						improveWayCount++
						if improveWaitsCount > maxImproveWaitsCount34[i] {
							maxImproveWaitsCount34[i] = improveWaitsCount
							// improves 選的是進張數最大的改良
							improves[i] = improveWaits
						}
						//fmt.Println(fmt.Sprintf("    摸 %s 切 %s 改良:", MahjongZH[i], MahjongZH[j]), improveWaitsCount, TilesToStrWithBracket(improveWaits.indexes()))
					}
				}
				playerInfo.UndoDiscardTile(j, _isRedFive)
			}
		}

		tiles34[i]--
		leftTiles34[i]++
	}

	_tiles34 := make([]int, 34)
	copy(_tiles34, tiles34)
	result13 = &Hand13AnalysisResult{
		Tiles34:                  _tiles34,
		LeftTiles34:              leftTiles34,
		IsNaki:                   playerInfo.IsNaki(),
		Shanten:                  shanten13,
		Waits:                    waits,
		DamaWaits:                Waits{},
		NextShantenWaitsCountMap: nextShantenWaitsCountMap,
		Improves:                 improves,
		ImproveWayCount:          improveWayCount,
		AvgImproveWaitsCount:     float64(waitsCount),
		YakuTypes:                yakuTypes,
		DoraCount:                playerInfo.CountDora(),
	}

	// 計算局收支、打點、和率和役種
	if waitsCount > 0 {
		//avgAgariRate /= float64(waitsCount)
		if roundPointWeight > 0 {
			avgRoundPoint /= float64(roundPointWeight)
			//if shanten13 == 1 {
			//	avgRoundPoint /= 6 // TODO: 待調整？
			//} else if shanten13 == 2 {
			//	avgRoundPoint /= 18 // TODO: 待調整？
			//}
		}
		//avgRiichiPoint /= float64(waitsCount)
		if shanten13 == shantenStateTenpai {
			// TODO: 考慮默聽時的自摸
			avgRonPoint, pointResults := CalcAvgPoint(*playerInfo, waits)
			result13.DamaPoint = avgRonPoint
			// 計算默聽進張
			for _, pr := range pointResults {
				result13.DamaWaits[pr.winTile] = leftTiles34[pr.winTile]
			}

			if !result13.IsNaki {
				avgRiichiPoint, riichiPointResults := CalcAvgRiichiPoint(*playerInfo, waits)
				result13.RiichiPoint = avgRiichiPoint
				result13.AvgAgariRate = CalculateAvgAgariRate(waits, playerInfo)
				for _, pr := range riichiPointResults {
					for _, yakuType := range pr.yakuTypes {
						result13.YakuTypes[yakuType] = struct{}{}
					}
				}
			} else {
				// 副露時，考慮到存在某些侍牌無法和牌（如後付、片聽），不計算這些侍牌的和率
				agariRate := 0.0
				for _, pr := range pointResults { // pointResults 不包含無法和牌的情況
					agariRate = agariRate + pr.agariRate - agariRate*pr.agariRate/100
					for _, yakuType := range pr.yakuTypes {
						result13.YakuTypes[yakuType] = struct{}{}
					}
				}
				result13.AvgAgariRate = agariRate

				// 是否片聽
				result13.IsPartWait = len(pointResults) < len(waits.AvailableTiles())
			}
		}
	}

	// 三向聽七對子特殊提醒
	if len(playerInfo.Melds) == 0 && shanten13 == 3 && CountPairsOfTiles34(tiles34)+shanten13 == 6 {
		// 對於三向聽，除非進張很差才會考慮七對子
		if waitsCount <= 21 {
			result13.YakuTypes[YakuChiitoi] = struct{}{}
		}
	}

	// 對於聽牌及一向聽，判斷是否有振聽可能
	if shanten13 <= 1 {
		for _, discardTile := range playerInfo.DiscardTiles {
			if _, ok := waits[discardTile]; ok {
				result13.FuritenRate = 0.5 // TODO: 待完善
				if shanten13 == shantenStateTenpai {
					result13.FuritenRate = 1
				}
			}
		}
	}

	// 計算局收支
	//if shanten13 <= 1 {
	//result13.DamaPoint = avgRonPoint
	//if !result13.IsNaki {
	//	result13.RiichiPoint = avgRiichiPoint
	//}
	// 振聽時若能立直則只考慮立直
	//if result13.FuritenRate == 1 && result13.RiichiPoint > 0 {
	//	result13.DamaPoint = 0
	//}
	if shanten13 == shantenStateTenpai {
		result13.MixedRoundPoint = result13.mixedRoundPoint()
	} else {
		result13.MixedRoundPoint = avgRoundPoint
	}
	//}

	// 計算手牌速度
	if len(nextShantenWaitsCountMap) > 0 {
		nextShantenWaitsSum := 0
		weight := 0
		for tile, c := range nextShantenWaitsCountMap {
			w := leftTiles34[tile]
			nextShantenWaitsSum += w * c
			weight += w
		}
		result13.AvgNextShantenWaitsCount = float64(nextShantenWaitsSum) / float64(weight)
	}
	if len(improves) > 0 {
		improveWaitsSum := 0
		weight := 0
		for i := 0; i < 34; i++ {
			w := leftTiles34[i]
			improveWaitsSum += w * maxImproveWaitsCount34[i]
			weight += w
		}
		result13.AvgImproveWaitsCount = float64(improveWaitsSum) / float64(weight)
	}
	result13.MixedWaitsScore = result13.speedScore()

	// 特殊處理，方便提示向聽倒退！
	if shanten13 == 2 {
		result13.MixedWaitsScore /= 4 // TODO: 待調整
	}

	return
}

func _stopShanten(shanten int) int {
	if shanten >= 3 {
		return shanten - 1
	}
	return shanten - 2
}

// 3k+1 張牌，計算向聽數、進張、改良等（考慮了剩餘枚數）
func CalculateShantenWithImproves13(playerInfo *model.PlayerInfo) (r *Hand13AnalysisResult) {
	if len(playerInfo.LeftTiles34) == 0 {
		playerInfo.FillLeftTiles34()
	}

	shanten := CalculateShanten(playerInfo.HandTiles34)
	shantenSearchRoot := _search13(shanten, playerInfo, _stopShanten(shanten))
	return shantenSearchRoot.analysis(playerInfo, true)
}

//

const (
	honorRiskRoundWind = 4
	honorRiskYaku      = 3
	honorRiskOtakaze   = 2
	honorRiskSelfWind  = 1
)

type tileValue float64

const (
	doraValue                tileValue = 10000
	doraFirstNeighbourValue  tileValue = 1000
	doraSecondNeighbourValue tileValue = 100
	honoredValue             tileValue = 15
)

func calculateIsolatedTileValue(tile int, playerInfo *model.PlayerInfo) tileValue {
	value := tileValue(100)

	// 是否為寶牌
	for _, doraTile := range playerInfo.DoraTiles {
		if tile == doraTile {
			value += doraValue
			//} else if doraTile < 27 {
			//	if tile/3 != doraTile/3 {
			//		continue
			//	}
			//	t9 := tile % 9
			//	dt9 := doraTile % 9
			//	if t9+1 == dt9 || t9-1 == dt9 {
			//		value += doraFirstNeighbourValue
			//	} else if t9+2 == dt9 || t9-2 == dt9 {
			//		value += doraSecondNeighbourValue
			//	}
		}
	}

	if tile >= 27 {
		if tile == playerInfo.SelfWindTile || tile == playerInfo.RoundWindTile || tile >= 31 {
			// 役牌
			value += honoredValue
			if playerInfo.SelfWindTile == playerInfo.RoundWindTile && tile == playerInfo.SelfWindTile {
				value += honoredValue // 連風
			} else if tile == playerInfo.SelfWindTile {
				value++ // 自風 +1
			} else if tile == playerInfo.RoundWindTile {
				value-- // 場風 -1
			}
			if tile == 31 {
				value -= 0.1
			}
			if tile == 32 {
				value -= 0.2
			}
		} else {
			// 客風
			for i := 1; i <= 3; i++ {
				otakazeTile := playerInfo.SelfWindTile + i
				if otakazeTile > 30 {
					otakazeTile -= 4
				}
				if tile == otakazeTile {
					// 下家 -3  對家 -2  上家 -1
					value -= tileValue(4 - i)
					break
				}
			}
		}
		left := playerInfo.LeftTiles34[tile]
		if left == 2 {
			value *= 0.9
		} else if left == 1 {
			value *= 0.2
		} else if left == 0 {
			value = 0
		}
	}

	return value
}

func calculateTileValue(tile int, playerInfo *model.PlayerInfo) (value tileValue) {
	// 是否為寶牌或寶牌周邊
	for _, doraTile := range playerInfo.DoraTiles {
		if tile == doraTile {
			value += doraValue
		} else if doraTile < 27 {
			if tile/3 != doraTile/3 {
				continue
			}
			t9 := tile % 9
			dt9 := doraTile % 9
			if t9+1 == dt9 || t9-1 == dt9 {
				value += doraFirstNeighbourValue
			} else if t9+2 == dt9 || t9-2 == dt9 {
				value += doraSecondNeighbourValue
			}
		}
	}
	return
}

type Hand14AnalysisResult struct {
	// 需要切的牌
	DiscardTile int

	// 切的是否為寶牌
	IsDiscardDoraTile bool

	// 切的牌的價值（寶牌或寶牌周邊）
	DiscardTileValue tileValue

	// 切的牌是否為幺九浮牌
	isIsolatedYaochuDiscardTile bool

	// 切牌後的手牌分析結果
	Result13 *Hand13AnalysisResult

	DiscardHonorTileRisk int

	// 剩餘可以摸的牌數
	LeftDrawTilesCount int

	// 副露信息（沒有副露就是 nil）
	// 比如用 23m 吃了牌，OpenTiles 就是 [1,2]
	OpenTiles []int
}

func (r *Hand14AnalysisResult) String() string {
	meldInfo := ""
	if len(r.OpenTiles) > 0 {
		meldType := "吃"
		if r.OpenTiles[0] == r.OpenTiles[1] {
			meldType = "碰"
		}
		meldInfo = fmt.Sprintf("用 %s%s %s，", string([]rune(MahjongZH[r.OpenTiles[0]])[:1]), MahjongZH[r.OpenTiles[1]], meldType)
	}
	return meldInfo + fmt.Sprintf("切 %s: %s", MahjongZH[r.DiscardTile], r.Result13.String())
}

type Hand14AnalysisResultList []*Hand14AnalysisResult

// 按照特定規則排序
// 若 improveFirst 為 true，則優先按照 AvgImproveWaitsCount 排序（對於三向聽及以上來說）
func (l Hand14AnalysisResultList) Sort(improveFirst bool) {
	if len(l) <= 1 {
		return
	}

	shanten := l[0].Result13.Shanten

	sort.Slice(l, func(i, j int) bool {
		ri, rj := l[i].Result13, l[j].Result13
		riWaitsCount, rjWaitsCount := ri.Waits.AllCount(), rj.Waits.AllCount()

		// 首先，無論怎樣，進張數為 0，無條件排在後面，也不看改良
		// 進張數都為 0 才看改良
		if riWaitsCount == 0 || rjWaitsCount == 0 {
			if riWaitsCount == 0 && rjWaitsCount == 0 {
				return ri.AvgImproveWaitsCount > rj.AvgImproveWaitsCount
			}
			return riWaitsCount > rjWaitsCount
		}

		switch shanten {
		case 0:
			// 聽牌的話：局收支 - 和率
			// 局收支，有明顯差異
			if !InDelta(ri.MixedRoundPoint, rj.MixedRoundPoint, 100) {
				return ri.MixedRoundPoint > rj.MixedRoundPoint
			}
			// 和率優先
			if !Equal(ri.AvgAgariRate, rj.AvgAgariRate) {
				return ri.AvgAgariRate > rj.AvgAgariRate
			}
		case 1:
			// 一向聽：進張*局收支
			var riScore, rjScore float64
			if shanten >= 2 && improveFirst {
				// 對於兩向聽，若需要改良的話以改良為主
				//riScore = float64(ri.AvgImproveWaitsCount) * ri.MixedRoundPoint
				//rjScore = float64(rj.AvgImproveWaitsCount) * rj.MixedRoundPoint
				break
			} else {
				// 負數要調整
				wi := float64(riWaitsCount)
				if ri.MixedRoundPoint < 0 {
					wi = 1 / wi
				}
				wj := float64(rjWaitsCount)
				if rj.MixedRoundPoint < 0 {
					wj = 1 / wj
				}
				riScore = wi * ri.MixedRoundPoint
				rjScore = wj * rj.MixedRoundPoint
			}
			if !Equal(riScore, rjScore) {
				return riScore > rjScore
			}
		}

		if shanten >= 2 {
			// 兩向聽及以上時，若存在幺九浮牌，則根據價值來單獨比較浮牌
			if l[i].isIsolatedYaochuDiscardTile && l[j].isIsolatedYaochuDiscardTile {
				// 優先切掉價值最低的浮牌，這裏直接比較浮點數
				if l[i].DiscardTileValue != l[j].DiscardTileValue {
					return l[i].DiscardTileValue < l[j].DiscardTileValue
				}
			} else if l[i].isIsolatedYaochuDiscardTile && l[i].DiscardTileValue < 500 {
				return true
			} else if l[j].isIsolatedYaochuDiscardTile && l[j].DiscardTileValue < 500 {
				return false
			}
		}

		//if improveFirst {
		//	// 優先按照 AvgImproveWaitsCount 排序
		//	if !Equal(ri.AvgImproveWaitsCount, rj.AvgImproveWaitsCount) {
		//		return ri.AvgImproveWaitsCount > rj.AvgImproveWaitsCount
		//	}
		//}

		// 排序規則：綜合評分（速度） - 進張 - 前進後的進張 - 和率 - 改良 - 價值低 - 好牌先走
		// 必須注意到的一點是，隨著遊戲的進行，進張會被他家打出，所以進張是有減少的趨勢的
		// 對於一向聽，考慮到未聽牌之前要聽的牌會被他家打出而造成聽牌時的枚數降低，所以聽牌枚數比和率更重要
		// 對比當前進張與前進後的進張，在二者綜合評分相近的情況下（注意這個前提），由於進張越多聽牌速度越快，聽牌時的進張數也就越接近預期進張數，所以進張越多越好（再次強調是在二者綜合評分相近的情況下）

		if !Equal(ri.MixedWaitsScore, rj.MixedWaitsScore) {
			return ri.MixedWaitsScore > rj.MixedWaitsScore
		}

		if riWaitsCount != rjWaitsCount {
			return riWaitsCount > rjWaitsCount
		}

		if !Equal(ri.AvgNextShantenWaitsCount, rj.AvgNextShantenWaitsCount) {
			return ri.AvgNextShantenWaitsCount > rj.AvgNextShantenWaitsCount
		}

		// shanten == 1
		if !Equal(ri.AvgAgariRate, rj.AvgAgariRate) {
			return ri.AvgAgariRate > rj.AvgAgariRate
		}

		if !Equal(ri.AvgImproveWaitsCount, rj.AvgImproveWaitsCount) {
			return ri.AvgImproveWaitsCount > rj.AvgImproveWaitsCount
		}

		if l[i].DiscardTileValue != l[j].DiscardTileValue {
			// 價值低的先走
			return l[i].DiscardTileValue < l[j].DiscardTileValue
		}

		// 好牌先走
		idxI, idxJ := l[i].DiscardTile, l[j].DiscardTile
		if idxI < 27 && idxJ < 27 {
			idxI %= 9
			if idxI > 4 {
				idxI = 8 - idxI
			}
			idxJ %= 9
			if idxJ > 4 {
				idxJ = 8 - idxJ
			}
			return idxI > idxJ
		}
		if idxI < 27 || idxJ < 27 {
			// 數牌先走
			return idxI < idxJ
		}
		// 場風 - 三元牌 - 他家客風 - 自風
		return l[i].DiscardHonorTileRisk > l[j].DiscardHonorTileRisk

		//// 改良種類、方式多的優先
		//if len(ri.Improves) != len(rj.Improves) {
		//	return len(ri.Improves) > len(rj.Improves)
		//}
		//if ri.ImproveWayCount != rj.ImproveWayCount {
		//	return ri.ImproveWayCount > rj.ImproveWayCount
		//}
	})
}

func (l *Hand14AnalysisResultList) filterOutDiscard(cantDiscardTile int) {
	newResults := Hand14AnalysisResultList{}
	for _, r := range *l {
		if r.DiscardTile != cantDiscardTile {
			newResults = append(newResults, r)
		}
	}
	*l = newResults
}

func (l Hand14AnalysisResultList) addOpenTile(openTiles []int) {
	for _, r := range l {
		r.OpenTiles = openTiles
	}
}

func (n *shantenSearchNode14) analysis(playerInfo *model.PlayerInfo, considerImprove bool) (results Hand14AnalysisResultList) {
	for discardTile, node13 := range n.children {
		isRedFive := playerInfo.IsOnlyRedFive(discardTile)

		// 切牌，然後分析 3k+1 張牌下的手牌情況
		// 若這張是5，在只有赤5的情況下才會切赤5（TODO: 考慮赤5騙37）
		playerInfo.DiscardTile(discardTile, isRedFive)
		result13 := node13.analysis(playerInfo, considerImprove)

		// 記錄切牌後的分析結果
		r14 := &Hand14AnalysisResult{
			DiscardTile:        discardTile,
			IsDiscardDoraTile:  InInts(discardTile, playerInfo.DoraTiles),
			Result13:           result13,
			LeftDrawTilesCount: playerInfo.LeftDrawTilesCount,
		}
		results = append(results, r14)

		if n.shanten >= 2 {
			if isYaochupai(discardTile) && isIsolatedTile(discardTile, playerInfo.HandTiles34) {
				r14.isIsolatedYaochuDiscardTile = true
				r14.DiscardTileValue = calculateIsolatedTileValue(discardTile, playerInfo)
			} else {
				r14.DiscardTileValue = calculateTileValue(discardTile, playerInfo)
			}
		}

		if discardTile >= 27 {
			switch discardTile {
			case playerInfo.RoundWindTile:
				r14.DiscardHonorTileRisk = honorRiskRoundWind
			case 31, 32, 33:
				r14.DiscardHonorTileRisk = honorRiskYaku
			case playerInfo.SelfWindTile:
				r14.DiscardHonorTileRisk = honorRiskSelfWind
			default:
				r14.DiscardHonorTileRisk = honorRiskOtakaze
			}
		}

		playerInfo.UndoDiscardTile(discardTile, isRedFive)
	}

	// 下面這一邏輯被「綜合速度」取代
	//improveFirst := func(l []*Hand14AnalysisResult) bool {
	//	if !considerImprove || len(l) <= 1 {
	//		return false
	//	}
	//
	//	shanten := l[0].Result13.Shanten
	//	// 一向聽及以下著眼於進張，改良其次
	//	if shanten <= 1 {
	//		return false
	//	}
	//
	//	// 判斷七對和一般型的向聽數是否相同，若七對更小則改良優先
	//	tiles34 := playerInfo.HandTiles34
	//	shantenChiitoi := CalculateShantenOfChiitoi(tiles34)
	//	shantenNormal := CalculateShantenOfNormal(tiles34, CountOfTiles34(tiles34))
	//	return shantenChiitoi < shantenNormal
	//}
	//improveFst := improveFirst(results)

	results.Sort(false)

	return
}

// 3k+2 張牌，計算向聽數、進張、改良、向聽倒退等
func CalculateShantenWithImproves14(playerInfo *model.PlayerInfo) (shanten int, results Hand14AnalysisResultList, incShantenResults Hand14AnalysisResultList) {
	if len(playerInfo.LeftTiles34) == 0 {
		playerInfo.FillLeftTiles34()
	}

	shanten = CalculateShanten(playerInfo.HandTiles34)
	stopAtShanten := _stopShanten(shanten)
	shantenSearchRoot := searchShanten14(shanten, playerInfo, stopAtShanten)
	results = shantenSearchRoot.analysis(playerInfo, true)
	incShantenSearchRoot := searchShanten14(shanten+1, playerInfo, stopAtShanten+1)
	incShantenResults = incShantenSearchRoot.analysis(playerInfo, true)
	return
}

// 計算最小向聽數，鳴牌方式
func calculateMeldShanten(tiles34 []int, calledTile int, isRedFive bool, allowChi bool) (minShanten int, meldCombinations []model.Meld) {
	// 是否能碰
	if tiles34[calledTile] >= 2 {
		meldCombinations = append(meldCombinations, model.Meld{
			MeldType:          model.MeldTypePon,
			Tiles:             []int{calledTile, calledTile, calledTile},
			SelfTiles:         []int{calledTile, calledTile},
			CalledTile:        calledTile,
			RedFiveFromOthers: isRedFive,
		})
	}
	// 是否能吃
	if allowChi && calledTile < 27 {
		checkChi := func(tileA, tileB int) {
			if tiles34[tileA] > 0 && tiles34[tileB] > 0 {
				_tiles := []int{tileA, tileB, calledTile}
				sort.Ints(_tiles)
				meldCombinations = append(meldCombinations, model.Meld{
					MeldType:          model.MeldTypeChi,
					Tiles:             _tiles,
					SelfTiles:         []int{tileA, tileB},
					CalledTile:        calledTile,
					RedFiveFromOthers: isRedFive,
				})
			}
		}
		t9 := calledTile % 9
		if t9 >= 2 {
			checkChi(calledTile-2, calledTile-1)
		}
		if t9 >= 1 && t9 <= 7 {
			checkChi(calledTile-1, calledTile+1)
		}
		if t9 <= 6 {
			checkChi(calledTile+1, calledTile+2)
		}
	}

	// 計算所有鳴牌下的最小向聽數
	minShanten = 99
	for _, c := range meldCombinations {
		tiles34[c.SelfTiles[0]]--
		tiles34[c.SelfTiles[1]]--
		minShanten = MinInt(minShanten, CalculateShanten(tiles34))
		tiles34[c.SelfTiles[0]]++
		tiles34[c.SelfTiles[1]]++
	}

	return
}

// TODO 鳴牌的情況判斷（待重構）
// 編程時注意他家切掉的這張牌是否算到剩餘數中
//if isOpen {
//if newShanten, combinations, shantens := calculateMeldShanten(tiles34, i, true); newShanten < shanten {
//	// 向聽前進了，說明鳴牌成功，則換的這張牌為鳴牌進張
//	// 計算進張數：若能碰則 =剩餘數*3，否則 =剩餘數
//	meldWaits[i] = leftTile - tiles34[i]
//	for i, comb := range combinations {
//		if comb[0] == comb[1] && shantens[i] == newShanten {
//			meldWaits[i] *= 3
//			break
//		}
//	}
//}
//}

// 計算鳴牌下的何切分析
// calledTile 他家出的牌，嘗試鳴這張牌
// isRedFive 這張牌是否為赤5
// allowChi 是否允許吃這張牌
func CalculateMeld(playerInfo *model.PlayerInfo, calledTile int, isRedFive bool, allowChi bool) (minShanten int, results Hand14AnalysisResultList, incShantenResults Hand14AnalysisResultList) {
	if len(playerInfo.LeftTiles34) == 0 {
		playerInfo.FillLeftTiles34()
	}

	minShanten, meldCombinations := calculateMeldShanten(playerInfo.HandTiles34, calledTile, isRedFive, allowChi)

	for _, c := range meldCombinations {
		// 嘗試鳴這張牌
		playerInfo.AddMeld(c)
		_shanten, _results, _incShantenResults := CalculateShantenWithImproves14(playerInfo)
		playerInfo.UndoAddMeld()

		// 去掉現物食替的情況
		_results.filterOutDiscard(calledTile)
		_incShantenResults.filterOutDiscard(calledTile)

		// 去掉筋食替的情況
		if c.MeldType == model.MeldTypeChi {
			cannotDiscardTile := -1
			if c.SelfTiles[0] < calledTile && c.SelfTiles[1] < calledTile && calledTile%9 >= 3 {
				cannotDiscardTile = calledTile - 3
			} else if c.SelfTiles[0] > calledTile && c.SelfTiles[1] > calledTile && calledTile%9 <= 5 {
				cannotDiscardTile = calledTile + 3
			}
			if cannotDiscardTile != -1 {
				_results.filterOutDiscard(cannotDiscardTile)
				_incShantenResults.filterOutDiscard(cannotDiscardTile)
			}
		}

		// 添加副露信息，用於輸出
		_results.addOpenTile(c.SelfTiles)
		_incShantenResults.addOpenTile(c.SelfTiles)

		// 整理副露結果
		if _shanten == minShanten {
			results = append(results, _results...)
			incShantenResults = append(incShantenResults, _incShantenResults...)
		} else if _shanten == minShanten+1 {
			incShantenResults = append(incShantenResults, _results...)
		}
	}

	results.Sort(false)
	incShantenResults.Sort(false)

	return
}
