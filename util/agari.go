package util

import "fmt"

func _calcKey(tiles34 []int) (key int) {
	bitPos := -1

	// 數牌
	idx := -1
	for i := 0; i < 3; i++ {
		prevInHand := false // 上一張牌是否在手牌中
		for j := 0; j < 9; j++ {
			idx++
			if c := tiles34[idx]; c > 0 {
				prevInHand = true
				bitPos++
				switch c {
				case 2:
					key |= 0x3 << uint(bitPos)
					bitPos += 2
				case 3:
					key |= 0xF << uint(bitPos)
					bitPos += 4
				case 4:
					key |= 0x3F << uint(bitPos)
					bitPos += 6
				}
			} else {
				if prevInHand {
					prevInHand = false
					key |= 0x1 << uint(bitPos)
					bitPos++
				}
			}
		}
		if prevInHand {
			key |= 0x1 << uint(bitPos)
			bitPos++
		}
	}

	// 字牌
	for i := 27; i < 34; i++ {
		if c := tiles34[i]; c > 0 {
			bitPos++
			switch c {
			case 2:
				key |= 0x3 << uint(bitPos)
				bitPos += 2
			case 3:
				key |= 0xF << uint(bitPos)
				bitPos += 4
			case 4:
				key |= 0x3F << uint(bitPos)
				bitPos += 6
			}
			key |= 0x1 << uint(bitPos)
			bitPos++
		}
	}

	return
}

// 3k+2 張牌，是否和牌（不檢測國士無雙）
func IsAgari(tiles34 []int) bool {
	key := _calcKey(tiles34)
	_, isAgari := winTable[key]
	return isAgari
}

//

// 3k+2 張牌的某種拆解結果
type DivideResult struct {
	PairTile          int   // 雀頭牌
	KotsuTiles        []int // 刻子牌（注意 len(KotsuTiles) 為自摸時的暗刻數，榮和時的暗刻數需要另加邏輯判斷）
	ShuntsuFirstTiles []int // 順子牌的第一張（如 678s 的 6s）

	// 由於生成 winTable 的代碼是不考慮具體是什麼牌的，
	// 所以只能判斷如七對子、九蓮寶燈、一氣通貫、兩盃口、一盃口等和「形狀」有關的役，
	// 像國士無雙、斷幺、全帶、三色、綠一色等，和具體的牌/位置有關的役是判斷不出的，需要另加邏輯判斷
	IsChiitoi       bool // 七對子
	IsChuurenPoutou bool // 九蓮寶燈
	IsIttsuu        bool // 一氣通貫（注意：未考慮副露！）
	IsRyanpeikou    bool // 兩盃口（IsRyanpeikou == true 時 IsIipeikou == false）
	IsIipeikou      bool // 一盃口
}

// 調試用
func (d *DivideResult) String() string {
	if d.IsChiitoi {
		return "[七對子]"
	}

	output := ""

	humanTilesList := []string{TilesToStr([]int{d.PairTile, d.PairTile})}
	for _, kotsuTile := range d.KotsuTiles {
		humanTilesList = append(humanTilesList, TilesToStr([]int{kotsuTile, kotsuTile, kotsuTile}))
	}
	for _, shuntsuFirstTile := range d.ShuntsuFirstTiles {
		humanTilesList = append(humanTilesList, TilesToStr([]int{shuntsuFirstTile, shuntsuFirstTile + 1, shuntsuFirstTile + 2}))
	}
	output += fmt.Sprint(humanTilesList)

	if d.IsChuurenPoutou {
		output += "[九蓮寶燈]"
	}
	if d.IsIttsuu {
		output += "[一氣通貫]"
	}
	if d.IsRyanpeikou {
		output += "[兩盃口]"
	}
	if d.IsIipeikou {
		output += "[一盃口]"
	}

	return output
}

// 3k+2 張牌，返回所有可能的拆解，沒有拆解表示未和牌（不檢測國士無雙）
// http://hp.vector.co.jp/authors/VA046927/mjscore/mjalgorism.html
// http://hp.vector.co.jp/authors/VA046927/mjscore/AgariIndex.java
func DivideTiles34(tiles34 []int) (divideResults []*DivideResult) {
	tiles14 := make([]int, 14)
	tiles14TailIndex := 0

	key := 0
	bitPos := -1

	// 數牌
	idx := -1
	for i := 0; i < 3; i++ {
		prevInHand := false // 上一張牌是否在手牌中
		for j := 0; j < 9; j++ {
			idx++
			if c := tiles34[idx]; c > 0 {
				tiles14[tiles14TailIndex] = idx
				tiles14TailIndex++

				prevInHand = true
				bitPos++
				switch c {
				case 2:
					key |= 0x3 << uint(bitPos)
					bitPos += 2
				case 3:
					key |= 0xF << uint(bitPos)
					bitPos += 4
				case 4:
					key |= 0x3F << uint(bitPos)
					bitPos += 6
				}
			} else {
				if prevInHand {
					prevInHand = false
					key |= 0x1 << uint(bitPos)
					bitPos++
				}
			}
		}
		if prevInHand {
			key |= 0x1 << uint(bitPos)
			bitPos++
		}
	}

	// 字牌
	for i := 27; i < 34; i++ {
		if c := tiles34[i]; c > 0 {
			tiles14[tiles14TailIndex] = i
			tiles14TailIndex++

			bitPos++
			switch c {
			case 2:
				key |= 0x3 << uint(bitPos)
				bitPos += 2
			case 3:
				key |= 0xF << uint(bitPos)
				bitPos += 4
			case 4:
				key |= 0x3F << uint(bitPos)
				bitPos += 6
			}
			key |= 0x1 << uint(bitPos)
			bitPos++
		}
	}

	results, ok := winTable[key]
	if !ok {
		return
	}

	// 3bit  0: 刻子數(0～4)
	// 3bit  3: 順子數(0～4)
	// 4bit  6: 雀頭位置(1～13)
	// 4bit 10: 面子位置1(0～13) 刻子在前，順子在後
	// 4bit 14: 面子位置2(0～13)
	// 4bit 18: 面子位置3(0～13)
	// 4bit 22: 面子位置4(0～13)
	// 1bit 26: 七對子
	// 1bit 27: 九蓮寶燈
	// 1bit 28: 一氣通貫
	// 1bit 29: 兩盃口
	// 1bit 30: 一盃口
	for _, r := range results {
		// 雀頭
		pairTile := tiles14[(r>>6)&0xF]

		// 刻子
		numKotsu := r & 0x7
		kotsuTiles := make([]int, numKotsu)
		for i := range kotsuTiles {
			kotsuTiles[i] = tiles14[(r>>uint(10+i*4))&0xF]
		}

		// 順子的第一張牌
		numShuntsu := (r >> 3) & 0x7
		shuntsuFirstTiles := make([]int, numShuntsu)
		for i := range shuntsuFirstTiles {
			shuntsuFirstTiles[i] = tiles14[(r>>uint(10+(numKotsu+i)*4))&0xF]
		}

		divideResults = append(divideResults, &DivideResult{
			PairTile:          pairTile,
			KotsuTiles:        kotsuTiles,
			ShuntsuFirstTiles: shuntsuFirstTiles,
			IsChiitoi:         r&(1<<26) != 0,
			IsChuurenPoutou:   r&(1<<27) != 0,
			IsIttsuu:          r&(1<<28) != 0,
			IsRyanpeikou:      r&(1<<29) != 0,
			IsIipeikou:        r&(1<<30) != 0,
		})
	}

	return
}
