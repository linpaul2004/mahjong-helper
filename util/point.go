package util

import (
	"github.com/EndlessCheng/mahjong-helper/util/model"
)

// TODO: 考慮大三元和大四喜的包牌？

func roundUpPoint(point int) int {
	if point == 0 {
		return 0
	}
	return ((point-1)/100 + 1) * 100
}

func calcBasicPoint(han int, fu int, yakumanTimes int) (basicPoint int) {
	switch {
	case yakumanTimes > 0: // (x倍)役滿
		basicPoint = 8000 * yakumanTimes
	case han >= 13: // 累計役滿
		basicPoint = 8000
	case han >= 11: // 三倍滿
		basicPoint = 6000
	case han >= 8: // 倍滿
		basicPoint = 4000
	case han >= 6: // 跳滿
		basicPoint = 3000
	default:
		basicPoint = fu * (1 << uint(2+han))
		if basicPoint > 2000 { // 滿貫
			basicPoint = 2000
		}
	}
	return
}

// 番數 符數 役滿倍數 是否為親家
// 返回榮和點數
func CalcPointRon(han int, fu int, yakumanTimes int, isParent bool) (point int) {
	basicPoint := calcBasicPoint(han, fu, yakumanTimes)
	if isParent {
		point = 6 * basicPoint
	} else {
		point = 4 * basicPoint
	}
	return roundUpPoint(point)
}

// 番數 符數 役滿倍數 是否為親家
// 返回自摸時的子家支付點數和親家支付點數
func CalcPointTsumo(han int, fu int, yakumanTimes int, isParent bool) (childPoint int, parentPoint int) {
	basicPoint := calcBasicPoint(han, fu, yakumanTimes)
	if isParent {
		childPoint = 2 * basicPoint
	} else {
		childPoint = basicPoint
		parentPoint = 2 * basicPoint
	}
	return roundUpPoint(childPoint), roundUpPoint(parentPoint)
}

// 番數 符數 役滿倍數 是否為親家
// 返回自摸時的點數
func CalcPointTsumoSum(han int, fu int, yakumanTimes int, isParent bool) int {
	childPoint, parentPoint := CalcPointTsumo(han, fu, yakumanTimes, isParent)
	if isParent {
		return 3 * childPoint
	}
	return 2*childPoint + parentPoint
}

//

type PointResult struct {
	Point      int
	FixedPoint float64 // 和牌時的期望點數

	han          int
	fu           int
	yakumanTimes int
	isParent     bool

	divideResult *DivideResult
	winTile      int
	yakuTypes    []int
	agariRate    float64 // 無役時的和率為 0
}

// 已和牌，計算自摸或榮和時的點數（不考慮裏寶、一發等情況）
// 無役時返回的點數為 0（和率也為 0）
// 調用前請設置 IsTsumo WinTile
func CalcPoint(playerInfo *model.PlayerInfo) (result *PointResult) {
	result = &PointResult{}
	isNaki := playerInfo.IsNaki()
	var han, fu int
	numDora := playerInfo.CountDora()
	for _, divideResult := range DivideTiles34(playerInfo.HandTiles34) {
		_hi := &_handInfo{
			PlayerInfo:   playerInfo,
			divideResult: divideResult,
		}
		yakuTypes := findYakuTypes(_hi, isNaki)
		if len(yakuTypes) == 0 {
			// 此手牌拆解下無役
			continue
		}
		yakumanTimes := CalcYakumanTimes(yakuTypes, isNaki)
		if yakumanTimes == 0 {
			han = CalcYakuHan(yakuTypes, isNaki)
			han += numDora
			fu = _hi.calcFu(isNaki)
		}
		var pt int
		if _hi.IsTsumo {
			pt = CalcPointTsumoSum(han, fu, yakumanTimes, _hi.IsParent)
		} else {
			pt = CalcPointRon(han, fu, yakumanTimes, _hi.IsParent)
		}
		_result := &PointResult{
			pt,
			float64(pt),
			han,
			fu,
			yakumanTimes,
			_hi.IsParent,
			divideResult,
			_hi.WinTile,
			yakuTypes,
			0.0, // 後面會補上
		}
		// 高點法
		if pt > result.Point {
			result = _result
		} else if pt == result.Point {
			if han > result.han {
				result = _result
			}
		}
	}
	return
}

// 已聽牌，根據 playerInfo 提供的信息計算加權和率後的平均點數
// 無役時返回 0
// 有役時返回平均點數（立直時考慮自摸、一發和裏寶）和各種待牌下的對應點數
func CalcAvgPoint(playerInfo model.PlayerInfo, waits Waits) (avgPoint float64, pointResults []*PointResult) {
	isFuriten := playerInfo.IsFuriten(waits)
	if isFuriten {
		// 振聽只能自摸，但是振聽立直時考慮了這一點，所以只在默聽或鳴牌時考慮
		if !playerInfo.IsRiichi {
			playerInfo.IsTsumo = true
		}
	}

	tileAgariRate := CalculateAgariRateOfEachTile(waits, &playerInfo)
	sum := 0.0
	weight := 0.0
	for tile, left := range waits {
		if left == 0 {
			continue
		}
		playerInfo.HandTiles34[tile]++
		playerInfo.WinTile = tile
		result := CalcPoint(&playerInfo) // 非振聽時，這裏算出的是榮和的點數
		playerInfo.HandTiles34[tile]--
		if result.Point == 0 {
			// 不考慮部分無役（如後付、片聽）
			continue
		}
		pt := float64(result.Point)
		if playerInfo.IsRiichi {
			// 如果立直了，需要考慮自摸、一發和裏寶
			pt = result.fixedRiichiPoint(isFuriten)
			result.FixedPoint = pt
		}
		w := tileAgariRate[tile]
		sum += pt * w
		weight += w
		result.agariRate = w
		pointResults = append(pointResults, result)
	}
	if weight > 0 {
		avgPoint = sum / weight
	}
	return
}

// 計算立直時的平均點數（考慮自摸、一發和裏寶）和各種待牌下的對應點數
// 已鳴牌時返回 0
// TODO: 剩餘不到 4 張無法立直
// TODO: 不足 1000 點無法立直
func CalcAvgRiichiPoint(playerInfo model.PlayerInfo, waits Waits) (avgRiichiPoint float64, pointResults []*PointResult) {
	if playerInfo.IsNaki() {
		return 0, nil
	}
	playerInfo.IsRiichi = true
	return CalcAvgPoint(playerInfo, waits)
}
