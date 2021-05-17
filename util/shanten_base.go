package util

import (
	"fmt"
)

const (
	shantenStateAgari  = -1
	shantenStateTenpai = 0
)

// 參考 http://ara.moo.jp/mjhmr/shanten.htm
// 七對子向聽數 = 6-對子數+max(0,7-種類數)
func CalculateShantenOfChiitoi(tiles34 []int) int {
	shanten := 6
	numKind := 0
	for _, c := range tiles34 {
		if c == 0 {
			continue
		}
		if c >= 2 {
			shanten--
		}
		numKind++
	}
	shanten += MaxInt(0, 7-numKind)
	return shanten
}

type shanten struct {
	tiles         []int
	numberMelds   int
	numberTatsu   int
	numberPairs   int
	numberJidahai int // 13枚にしてから少なくとも打牌しなければならない字牌の數 -> これより向聴數は下がらない
	ankanTiles    int // 暗槓，28bit 位壓縮：27bit數牌|1bit字牌
	isolatedTiles int // 孤張，28bit 位壓縮：27bit數牌|1bit字牌
	minShanten    int
}

func (st *shanten) scanCharacterTiles(countOfTiles int) {
	ankanTiles := 0    // 暗槓，7bit 位壓縮
	isolatedTiles := 0 // 孤張，7bit 位壓縮

	for i, c := range st.tiles[27:] {
		if c == 0 {
			continue
		}
		switch c {
		case 1:
			isolatedTiles |= 1 << uint(i)
		case 2:
			st.numberPairs++
		case 3:
			st.numberMelds++
		case 4:
			st.numberMelds++
			st.numberJidahai++
			ankanTiles |= 1 << uint(i)
			isolatedTiles |= 1 << uint(i)
		}
	}

	if st.numberJidahai > 0 && countOfTiles%3 == 2 {
		st.numberJidahai--
	}

	if isolatedTiles > 0 {
		st.isolatedTiles |= 1 << 27
		if ankanTiles|isolatedTiles == ankanTiles {
			// 此孤張不能視作單騎做雀頭的材料
			st.ankanTiles |= 1 << 27
		}
	}
}

// 計算一般型（非七對子和國士無雙）的向聽數
// 參考 http://ara.moo.jp/mjhmr/shanten.htm
func (st *shanten) calcNormalShanten() int {
	_shanten := 8 - 2*st.numberMelds - st.numberTatsu - st.numberPairs
	numMentsuKouho := st.numberMelds + st.numberTatsu
	if st.numberPairs > 0 {
		numMentsuKouho += st.numberPairs - 1 // 有雀頭時面子候補-1
	} else if st.ankanTiles > 0 && st.isolatedTiles > 0 {
		if st.ankanTiles|st.isolatedTiles == st.ankanTiles { // 沒有雀頭，且除了暗槓外沒有孤張，這連單騎都算不上
			// 比如 5555m 應該算作一向聽
			_shanten++
		}
	}
	if numMentsuKouho > 4 { // 面子候補過多
		_shanten += numMentsuKouho - 4
	}
	if _shanten != shantenStateAgari && _shanten < st.numberJidahai {
		return st.numberJidahai
	}
	return _shanten
}

// 拆分出一個暗刻
func (st *shanten) increaseSet(k int) {
	st.tiles[k] -= 3
	st.numberMelds++
}

func (st *shanten) decreaseSet(k int) {
	st.tiles[k] += 3
	st.numberMelds--
}

// 拆分出一個雀頭
func (st *shanten) increasePair(k int) {
	st.tiles[k] -= 2
	st.numberPairs++
}

func (st *shanten) decreasePair(k int) {
	st.tiles[k] += 2
	st.numberPairs--
}

// 拆分出一個順子
func (st *shanten) increaseSyuntsu(k int) {
	st.tiles[k]--
	st.tiles[k+1]--
	st.tiles[k+2]--
	st.numberMelds++
}
func (st *shanten) decreaseSyuntsu(k int) {
	st.tiles[k]++
	st.tiles[k+1]++
	st.tiles[k+2]++
	st.numberMelds--
}

// 拆分出一個兩面/邊張搭子
func (st *shanten) increaseTatsuFirst(k int) {
	st.tiles[k]--
	st.tiles[k+1]--
	st.numberTatsu++
}
func (st *shanten) decreaseTatsuFirst(k int) {
	st.tiles[k]++
	st.tiles[k+1]++
	st.numberTatsu--
}

// 拆分出一個坎張搭子
func (st *shanten) increaseTatsuSecond(k int) {
	st.tiles[k]--
	st.tiles[k+2]--
	st.numberTatsu++
}
func (st *shanten) decreaseTatsuSecond(k int) {
	st.tiles[k]++
	st.tiles[k+2]++
	st.numberTatsu--
}

// 拆分出一個孤張（浮牌）
func (st *shanten) increaseIsolatedTile(k int) {
	st.tiles[k]--
	st.isolatedTiles |= 1 << uint(k)
}
func (st *shanten) decreaseIsolatedTile(k int) {
	st.tiles[k]++
	st.isolatedTiles &^= 1 << uint(k)
}

func (st *shanten) run(depth int) {
	if st.minShanten == shantenStateAgari {
		return
	}

	// skip
	for ; depth < 27 && st.tiles[depth] == 0; depth++ {
	}

	if depth >= 27 {
		_shanten := st.calcNormalShanten()
		st.minShanten = MinInt(st.minShanten, _shanten)
		return
	}

	// i := depth % 9
	// 快速取模
	i := depth
	if i > 8 {
		i -= 9
	}
	if i > 8 {
		i -= 9
	}

	// 手牌拆解
	switch st.tiles[depth] {
	case 1:
		// 孤立牌は２つ以上取る必要は無い -> 雀頭のほうが向聴數は下がる -> ３枚 -> 雀頭＋孤立は雀頭から取る
		// 孤立牌は合計８枚以上取る必要は無い
		if i < 6 && st.tiles[depth+1] == 1 && st.tiles[depth+2] > 0 && st.tiles[depth+3] < 4 {
			// 延べ単
			// 順子
			st.increaseSyuntsu(depth)
			st.run(depth + 2)
			st.decreaseSyuntsu(depth)
		} else {
			// 浮牌
			st.increaseIsolatedTile(depth)
			st.run(depth + 1)
			st.decreaseIsolatedTile(depth)

			if i < 7 && st.tiles[depth+2] > 0 {
				if st.tiles[depth+1] != 0 {
					// 順子
					st.increaseSyuntsu(depth)
					st.run(depth + 1)
					st.decreaseSyuntsu(depth)
				}
				// 坎張搭子
				st.increaseTatsuSecond(depth)
				st.run(depth + 1)
				st.decreaseTatsuSecond(depth)
			}
			if i < 8 && st.tiles[depth+1] > 0 {
				// 兩面/邊張搭子
				st.increaseTatsuFirst(depth)
				st.run(depth + 1)
				st.decreaseTatsuFirst(depth)
			}
		}
	case 2:
		// 雀頭
		st.increasePair(depth)
		st.run(depth + 1)
		st.decreasePair(depth)

		if i < 7 && st.tiles[depth+1] > 0 && st.tiles[depth+2] > 0 {
			// 順子
			st.increaseSyuntsu(depth)
			st.run(depth)
			st.decreaseSyuntsu(depth)
		}
	case 3:
		// 暗刻
		st.increaseSet(depth)
		st.run(depth + 1)
		st.decreaseSet(depth)

		st.increasePair(depth)
		if i < 7 && st.tiles[depth+1] > 0 && st.tiles[depth+2] > 0 {
			// 雀頭+順子
			st.increaseSyuntsu(depth)
			st.run(depth + 1)
			st.decreaseSyuntsu(depth)
		} else {
			if i < 7 && st.tiles[depth+2] > 0 {
				// 雀頭+坎張搭子
				st.increaseTatsuSecond(depth)
				st.run(depth + 1)
				st.decreaseTatsuSecond(depth)
			}
			if i < 8 && st.tiles[depth+1] > 0 {
				// 雀頭+兩面/邊張搭子
				st.increaseTatsuFirst(depth)
				st.run(depth + 1)
				st.decreaseTatsuFirst(depth)
			}
		}
		st.decreasePair(depth)

		if i < 7 && st.tiles[depth+1] >= 2 && st.tiles[depth+2] >= 2 {
			// 一盃口
			st.increaseSyuntsu(depth)
			st.increaseSyuntsu(depth)
			st.run(depth)
			st.decreaseSyuntsu(depth)
			st.decreaseSyuntsu(depth)
		}
	case 4:
		st.increaseSet(depth)
		if i < 7 && st.tiles[depth+2] > 0 {
			if st.tiles[depth+1] > 0 {
				// 暗刻+順子
				st.increaseSyuntsu(depth)
				st.run(depth + 1)
				st.decreaseSyuntsu(depth)
			}
			// 暗刻+坎張搭子
			st.increaseTatsuSecond(depth)
			st.run(depth + 1)
			st.decreaseTatsuSecond(depth)
		}
		if i < 8 && st.tiles[depth+1] > 0 {
			// 暗刻+兩面/邊張搭子
			st.increaseTatsuFirst(depth)
			st.run(depth + 1)
			st.decreaseTatsuFirst(depth)
		}
		// 暗刻+孤張
		st.increaseIsolatedTile(depth)
		st.run(depth + 1)
		st.decreaseIsolatedTile(depth)
		st.decreaseSet(depth)

		st.increasePair(depth)
		if i < 7 && st.tiles[depth+2] > 0 {
			if st.tiles[depth+1] > 0 {
				// 雀頭+順子
				st.increaseSyuntsu(depth)
				st.run(depth)
				st.decreaseSyuntsu(depth)
			}
			// 雀頭+坎張搭子
			st.increaseTatsuSecond(depth)
			st.run(depth + 1)
			st.decreaseTatsuSecond(depth)
		}
		if i < 8 && st.tiles[depth+1] > 0 {
			// 雀頭+兩面/邊張搭子
			st.increaseTatsuFirst(depth)
			st.run(depth + 1)
			st.decreaseTatsuFirst(depth)
		}
		st.decreasePair(depth)
	}
}

// 根據手牌計算一般型（不考慮七對國士）的向聽數
// 3k+1 和 3k+2 張牌都行
func CalculateShantenOfNormal(tiles34 []int, countOfTiles int) int {
	st := shanten{
		numberMelds: (14 - countOfTiles) / 3,
		minShanten:  8, // 不考慮國士無雙和七對子的最大向聽
		tiles:       tiles34,
	}

	st.scanCharacterTiles(countOfTiles)

	for i, c := range st.tiles[:27] {
		if c == 4 {
			st.ankanTiles |= 1 << uint(i)
		}
	}

	st.run(0)

	return st.minShanten
}

// 根據手牌計算向聽數（不考慮國士）
// 3k+1 和 3k+2 張牌都行
func CalculateShanten(tiles34 []int) int {
	countOfTiles := CountOfTiles34(tiles34) // 若入參帶 countOfTiles，能節省約 5% 的時間
	if countOfTiles > 14 {
		panic(fmt.Sprintln("[CalculateShanten] 參數錯誤 >14", tiles34, countOfTiles))
	}
	minShanten := CalculateShantenOfNormal(tiles34, countOfTiles)
	if countOfTiles >= 13 { // 考慮七對子
		minShanten = MinInt(minShanten, CalculateShantenOfChiitoi(tiles34))
	}
	return minShanten
}
