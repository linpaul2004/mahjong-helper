package util

import "fmt"

// 根據實際信息，某些牌的危險度遠低於無筋（如現物、NC），這些牌可以用來計算筋牌的危險度
// TODO: 早外産生的筋牌可能要單獨計算
func calcLowRiskTiles27(safeTiles34 []bool, leftTiles34 []int) []int {
	lowRiskTiles27 := make([]int, 27)
	const _true = 1
	for i, safe := range safeTiles34[:27] {
		if safe {
			lowRiskTiles27[i] = _true
		}
	}
	for i := 0; i < 3; i++ {
		// 2斷，當做打過1
		if leftTiles34[9*i+1] == 0 {
			lowRiskTiles27[9*i] = _true
		}
		// 3斷，當做打過12
		if leftTiles34[9*i+2] == 0 {
			lowRiskTiles27[9*i] = _true
			lowRiskTiles27[9*i+1] = _true
		}
		// 4斷，當做打過23
		if leftTiles34[9*i+3] == 0 {
			lowRiskTiles27[9*i+1] = _true
			lowRiskTiles27[9*i+2] = _true
		}
		// 6斷，當做打過78
		if leftTiles34[9*i+5] == 0 {
			lowRiskTiles27[9*i+6] = _true
			lowRiskTiles27[9*i+7] = _true
		}
		// 7斷，當做打過89
		if leftTiles34[9*i+6] == 0 {
			lowRiskTiles27[9*i+7] = _true
			lowRiskTiles27[9*i+8] = _true
		}
		// 8斷，當做打過9
		if leftTiles34[9*i+7] == 0 {
			lowRiskTiles27[9*i+8] = _true
		}
	}
	return lowRiskTiles27
}

// 根據傳入的捨牌，計算 mpz 各個牌的筋牌類型
func calcTileType27(discardTiles []int) []tileType {
	sujiType27 := make([]tileType, 27)

	safeTiles34 := make([]int, 34)
	// 0危險，1安全
	for _, tile := range discardTiles {
		safeTiles34[tile] = 1
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			idx := 9*i + j
			sujiType27[idx] = TileTypeTable[j][safeTiles34[idx+3]]
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			mixSafeTile := safeTiles34[idx-3]<<1 | safeTiles34[idx+3]
			sujiType27[idx] = TileTypeTable[j][mixSafeTile]
		}
		for j := 6; j < 9; j++ {
			idx := 9*i + j
			sujiType27[idx] = TileTypeTable[j][safeTiles34[idx-3]]
		}
	}

	return sujiType27
}

type RiskTiles34 []float64

// 根據巡目（對於對手而言）、現物、立直後通過的牌、NC、Dora，來計算基礎銃率
// 至於早外、OC 和讀牌交給後續的計算
// turns: 巡目，這裏是對於對手而言的，也就是該玩家捨牌的次數
// safeTiles34: 現物及立直後通過的牌
// leftTiles34: 各個牌在山中剩餘的枚數
// roundWindTile: 場風
// playerWindTile: 自風
func CalculateRiskTiles34(turns int, safeTiles34 []bool, leftTiles34 []int, doraTiles []int, roundWindTile int, playerWindTile int) (risk34 RiskTiles34) {
	risk34 = make(RiskTiles34, 34)

	// 只對 dora 牌的危險度進行調整（綜合了放銃率和失點）
	// double dora 等的危險度會進一步升高
	doraMulti := func(tile int, tileType tileType) float64 {
		multi := 1.0
		for _, dora := range doraTiles {
			if tile == dora {
				multi *= FixedDoraRiskRateMulti[tileType]
			}
		}
		return multi
	}

	// 各個數牌的和牌方式
	// 19 - 兩面 對碰單騎
	// 28 - 兩面 坎張 對碰單騎
	// 37 - 兩面 坎張 邊張 對碰單騎
	// 456- 兩面x2 坎張 對碰單騎

	// 首先，根據現物和 No Chance 計算有沒有兩面的可能
	// 生成用來計算筋牌的「安牌」
	lowRiskTiles27 := calcLowRiskTiles27(safeTiles34, leftTiles34)
	// 利用「安牌」計算無筋、筋、半筋、雙筋的銃率
	// TODO: 特殊處理宣言牌的筋牌、宣言牌的同色牌的銃率
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			idx := 9*i + j
			t := TileTypeTable[j][lowRiskTiles27[idx+3]]
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
			if j == 0 && safeTiles34[idx+3] && leftTiles34[idx] == 0 {
				// (1) 兩面 對碰單騎 都不可能 -> 安牌
				risk34[idx] = 0
			}
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			mixSafeTile := lowRiskTiles27[idx-3]<<1 | lowRiskTiles27[idx+3]
			t := TileTypeTable[j][mixSafeTile]
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		}
		for j := 6; j < 9; j++ {
			idx := 9*i + j
			t := TileTypeTable[j][lowRiskTiles27[idx-3]]
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
			if j == 8 && safeTiles34[idx-3] && leftTiles34[idx] == 0 {
				// (9) 兩面 對碰單騎 都不可能 -> 安牌
				risk34[idx] = 0
			}
		}
		// 5斷，37視作安牌筋
		if leftTiles34[9*i+4] == 0 {
			t := tileTypeSuji37
			risk34[9*i+2] = RiskRate[turns][t] * doraMulti(9*i+2, t)
			risk34[9*i+6] = RiskRate[turns][t] * doraMulti(9*i+6, t)
		}
	}
	for i := 27; i < 34; i++ {
		if leftTiles34[i] > 0 {
			// 該玩家的役牌 = 場風/其自風/白/發/中
			isYakuHai := i == roundWindTile || i == playerWindTile || i >= 31
			t := HonorTileType[boolToInt(isYakuHai)][leftTiles34[i]-1]
			risk34[i] = RiskRate[turns][t] * doraMulti(i, t)
		} else {
			// 剩餘數為 0 可以視作安牌（忽略國士）
			risk34[i] = 0
		}
	}

	// TODO: 降級
	// 如 1m 為壁，2m 變成無筋 19 等級，3m 變成無筋 28 等級

	// 根據 No Chance 計算有沒有兩面的可能，完善上面的計算
	// 更新銃率表：No Chance 的危險度
	// 12和筋1差不多（2比1多10%）
	// 3和筋2差不多
	// 456和兩筋差不多（存疑？）
	ncSafeTile34 := CalcNCSafeTiles(leftTiles34)
	for _, ncSafeTile := range ncSafeTile34 {
		idx := ncSafeTile.Tile34
		switch idx%9 + 1 {
		case 1, 9:
			t := tileTypeSuji19
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		case 2, 8:
			t := tileTypeSuji19
			risk34[idx] = RiskRate[turns][t] * 1.1 * doraMulti(idx, t)
		case 3, 7:
			t := tileTypeSuji28
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		case 4, 6:
			t := tileTypeDoubleSuji46
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		case 5:
			t := tileTypeDoubleSuji5
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		default:
			panic(fmt.Errorf("[CalculateRiskTiles34] 代碼有誤: ncSafeTile = %d", ncSafeTile.Tile34))
		}
	}

	// 根據現物和 No Chance 計算是否只輸對碰單騎，在這種情況下安全度和筋 19 差不多；若剩餘枚數為 0 可直接視作現物（忽略國士）
	// 更新銃率表：Double No Chance 的危險度
	dncSafeTiles := CalcDNCSafeTilesWithDiscards(leftTiles34, safeTiles34)
	for _, dncSafeTile := range dncSafeTiles {
		tile := dncSafeTile.Tile34
		if leftTiles34[tile] > 0 {
			t := tileTypeSuji19
			risk34[tile] = RiskRate[turns][t] * doraMulti(tile, t)
			// 非19仍然有點斷幺的危險，危險度 *1.1
			if t9 := tile % 9; t9 > 0 && t9 < 8 {
				risk34[tile] *= 1.1
			}
		} else {
			risk34[tile] = 0
		}
	}

	// 更新銃率表：現物的銃率為 0
	for i, isSafe := range safeTiles34 {
		if isSafe {
			risk34[i] = 0
		}
	}

	return
}

// 對 5 巡前的外側牌的危險度進行調整
// 粗略調整為 *0.4（參考：科學する麻雀）
func (l RiskTiles34) FixWithEarlyOutside(discardTiles []int) RiskTiles34 {
	for _, dTile := range discardTiles {
		l[dTile] *= 0.4
	}
	return l
}

func (l RiskTiles34) FixWithGlobalMulti(multi float64) RiskTiles34 {
	for i := range l {
		l[i] *= multi
	}
	return l
}

// 根據副露情況對危險度進行修正
func (l RiskTiles34) FixWithPoint(ronPoint float64) RiskTiles34 {
	return l.FixWithGlobalMulti(ronPoint / RonPointRiichiHiIppatsu)
}

// 計算剩餘的無筋 123789 牌
// 總計 18 種。剩餘無筋牌數量越少，該無筋牌越危險
func CalculateLeftNoSujiTiles(safeTiles34 []bool, leftTiles34 []int) (leftNoSujiTiles []int) {
	isNoSujiTiles27 := make([]bool, 27)

	for i := 0; i < 3; i++ {
		// 根據 456 中張是否為安牌來判斷相應筋牌是否安全
		for j := 3; j < 6; j++ {
			if !safeTiles34[9*i+j] {
				isNoSujiTiles27[9*i+j-3] = true
				isNoSujiTiles27[9*i+j+3] = true
			}
		}
		// 5斷，37視作安牌筋
		if leftTiles34[9*i+4] == 0 {
			isNoSujiTiles27[9*i+2] = false
			isNoSujiTiles27[9*i+6] = false
		}
	}

	// 根據打過 4 張的壁牌更新 isNoSujiTiles27
	for i, c := range leftTiles34[:27] {
		if c == 0 {
			isNoSujiTiles27[i] = false
		}
	}

	// 根據 No Chance 的安牌更新 isNoSujiTiles27
	lowRiskTiles27 := calcLowRiskTiles27(safeTiles34, leftTiles34)
	const _true = 1
	for i, isSafe := range lowRiskTiles27 {
		if isSafe == _true {
			isNoSujiTiles27[i] = false
		}
	}

	for i, isNoSujiTile := range isNoSujiTiles27 {
		if isNoSujiTile {
			leftNoSujiTiles = append(leftNoSujiTiles, i)
		}
	}

	return
}

// TODO:（待定）有早外的半筋（早巡打過8m時，3m的半筋6m）
// TODO:（待定）利用赤寶牌計算銃率
// TODO: 寶牌周邊牌的危險度要增加一點
