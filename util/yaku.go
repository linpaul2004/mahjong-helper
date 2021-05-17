package util

import (
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"sort"
)

// 門清限定
func (hi *_handInfo) daburii() bool {
	return hi.IsDaburii
}

// 門清限定
func (hi *_handInfo) riichi() bool {
	return !hi.IsDaburii && hi.IsRiichi
}

// 門清限定
func (hi *_handInfo) tsumo() bool {
	return !hi.IsNaki() && hi.IsTsumo
}

// 門清限定
func (hi *_handInfo) chiitoi() bool {
	return hi.divideResult.IsChiitoi
}

// 門清限定
func (hi *_handInfo) pinfu() bool {
	// 順子數必須為 4
	if len(hi.divideResult.ShuntsuFirstTiles) != 4 {
		return false
	}

	// 雀頭不能是役牌
	if hi.isYakuTile(hi.divideResult.PairTile) {
		return false
	}

	for _, tile := range hi.divideResult.ShuntsuFirstTiles {
		// 是兩面和牌
		if tile%9 < 6 && tile == hi.WinTile || tile%9 > 0 && tile+2 == hi.WinTile {
			return true
		}
	}

	// 沒有兩面和牌
	return false
}

// 門清限定
func (hi *_handInfo) ryanpeikou() bool {
	return hi.divideResult.IsRyanpeikou
}

// 門清限定
// 兩盃口時不算一盃口
func (hi *_handInfo) iipeikou() bool {
	return hi.divideResult.IsIipeikou
}

func (hi *_handInfo) sanshokuDoujun() bool {
	shuntsuFirstTiles := hi.allShuntsuFirstTiles
	if len(shuntsuFirstTiles) < 3 {
		return false
	}
	var sMan, sPin, sSou []int
	for _, s := range shuntsuFirstTiles {
		if isMan(s) {
			sMan = append(sMan, s)
		} else if isPin(s) {
			sPin = append(sPin, s)
		} else { // isSou
			sSou = append(sSou, s)
		}
	}
	for _, man := range sMan {
		for _, pin := range sPin {
			for _, sou := range sSou {
				if man == pin-9 && man == sou-18 {
					return true
				}
			}
		}
	}
	return false
}

func (hi *_handInfo) ittsuu() bool {
	hasNakiShuntsu := false
	for _, meld := range hi.Melds {
		if meld.MeldType == model.MeldTypeChi {
			hasNakiShuntsu = true
			break
		}
	}
	if !hasNakiShuntsu {
		// 沒有鳴順子就直接用
		return hi.divideResult.IsIttsuu
	}

	shuntsuFirstTiles := hi.allShuntsuFirstTiles
	if len(shuntsuFirstTiles) < 3 {
		return false
	}
	// （這裏沒用排序，因為下面用的是更為快速的比較）
	// 若有 123，找是否有同色的 456 和 789
	for _, tile := range shuntsuFirstTiles {
		if tile%9 == 0 { // has123
			has456 := false
			has789 := false
			for _, otherTile := range shuntsuFirstTiles {
				if otherTile == tile+3 {
					has456 = true
				} else if otherTile == tile+6 {
					has789 = true
				}
			}
			if has456 && has789 {
				return true
			}
		}
	}
	return false
}

func (hi *_handInfo) toitoi() bool {
	return len(hi.allKotsuTiles) == 4
}

// 榮和的刻子是明刻
// 注意 456666 這樣的榮和 6，算暗刻
func (hi *_handInfo) sanAnkou() bool {
	num, _ := hi.numAnkou()
	return num == 3
}

func (hi *_handInfo) sanshokuDoukou() bool {
	kotsuTiles := hi.allKotsuTiles
	if len(kotsuTiles) < 3 {
		return false
	}
	var kMan, kPin, kSou []int
	for _, tile := range kotsuTiles {
		if isMan(tile) {
			kMan = append(kMan, tile)
		} else if isPin(tile) {
			kPin = append(kPin, tile)
		} else if isSou(tile) {
			kSou = append(kSou, tile)
		}
	}
	for _, man := range kMan {
		for _, pin := range kPin {
			for _, sou := range kSou {
				if man == pin-9 && man == sou-18 {
					return true
				}
			}
		}
	}
	return false
}

func (hi *_handInfo) sanKantsu() bool {
	if len(hi.Melds) < 3 {
		return false
	}
	return hi.numKantsu() == 3
}

func (hi *_handInfo) tanyao() bool {
	if len(hi.Melds) == 0 {
		// 沒副露時簡單判斷，這考慮了七對子的情況
		for _, tile := range YaochuTiles {
			if hi.HandTiles34[tile] > 0 {
				return false
			}
		}
		return true
	}

	// 所有雀頭和面子都不能包含幺九牌
	dr := hi.divideResult
	if isYaochupai(dr.PairTile) {
		return false
	}
	for _, tile := range hi.allShuntsuFirstTiles {
		if isYaochupai(tile) || isYaochupai(tile+2) {
			return false
		}
	}
	for _, tile := range hi.allKotsuTiles {
		if isYaochupai(tile) {
			return false
		}
	}
	return true
}

// 返回役牌個數，連風算兩個
func (hi *_handInfo) numYakuhai() (cnt int) {
	for _, tile := range hi.allKotsuTiles {
		if hi.isYakuTile(tile) {
			cnt++
			if hi.isDoubleWindTile(tile) {
				cnt++
			}
		}
	}
	return
}

func (hi *_handInfo) _chantai() bool {
	// 必須要有順子
	shuntsuFirstTiles := hi.allShuntsuFirstTiles
	if len(shuntsuFirstTiles) == 0 {
		return false
	}
	// 所有雀頭和面子都要包含幺九牌
	if !isYaochupai(hi.divideResult.PairTile) {
		return false
	}
	for _, tile := range shuntsuFirstTiles {
		if !isYaochupai(tile) && !isYaochupai(tile + 2) {
			return false
		}
	}
	for _, tile := range hi.allKotsuTiles {
		if !isYaochupai(tile) {
			return false
		}
	}
	return true
}

func (hi *_handInfo) chanta() bool {
	return hi.containHonor() && hi._chantai()
}

func (hi *_handInfo) junchan() bool {
	return !hi.containHonor() && hi._chantai()
}

func (hi *_handInfo) honroutou() bool {
	if !hi.containHonor() {
		return false
	}
	if len(hi.Melds) == 0 {
		// 沒副露時簡單判斷，這考慮了七對子的情況
		cnt := 0
		for _, tile := range YaochuTiles {
			cnt += hi.HandTiles34[tile]
		}
		return cnt == 14
	}

	// 不能有順子
	if len(hi.allShuntsuFirstTiles) > 0 {
		return false
	}
	if !isYaochupai(hi.divideResult.PairTile) {
		return false
	}
	for _, tile := range hi.allKotsuTiles {
		if !isYaochupai(tile) {
			return false
		}
	}
	return true
}

func (hi *_handInfo) shousangen() bool {
	// 檢查雀頭
	if hi.divideResult.PairTile < 31 {
		return false
	}
	// 檢查三元牌刻子個數
	cnt := 0
	for _, tile := range hi.allKotsuTiles {
		if tile >= 31 {
			cnt++
		}
	}
	return cnt == 2
}

func (hi *_handInfo) _numSuit() int {
	cntMan := 0
	cntPin := 0
	cntSou := 0
	cnt := func(tile int) {
		if isMan(tile) {
			cntMan++
		} else if isPin(tile) {
			cntPin++
		} else if isSou(tile) {
			cntSou++
		}
	}

	if hi.divideResult.IsChiitoi {
		// 七對子特殊判斷
		for i, c := range hi.HandTiles34[:27] {
			if c > 0 {
				cnt(i)
			}
		}
	} else {
		cnt(hi.divideResult.PairTile)
		for _, tile := range hi.allShuntsuFirstTiles {
			cnt(tile)
		}
		for _, tile := range hi.allKotsuTiles {
			cnt(tile)
		}
	}

	numSuit := 0
	if cntMan > 0 {
		numSuit++
	}
	if cntPin > 0 {
		numSuit++
	}
	if cntSou > 0 {
		numSuit++
	}
	return numSuit
}

func (hi *_handInfo) honitsu() bool {
	return hi.containHonor() && hi._numSuit() == 1
}

func (hi *_handInfo) chinitsu() bool {
	return !hi.containHonor() && hi._numSuit() == 1
}

type yakuChecker func(*_handInfo) bool

var yakuCheckerMap = map[int]yakuChecker{
	YakuDaburii:        (*_handInfo).daburii,
	YakuRiichi:         (*_handInfo).riichi,
	YakuChiitoi:        (*_handInfo).chiitoi,
	YakuTsumo:          (*_handInfo).tsumo,
	YakuPinfu:          (*_handInfo).pinfu,
	YakuRyanpeikou:     (*_handInfo).ryanpeikou,
	YakuIipeikou:       (*_handInfo).iipeikou,
	YakuSanshokuDoujun: (*_handInfo).sanshokuDoujun,
	YakuIttsuu:         (*_handInfo).ittsuu,
	YakuToitoi:         (*_handInfo).toitoi,
	YakuSanAnkou:       (*_handInfo).sanAnkou,
	YakuSanshokuDoukou: (*_handInfo).sanshokuDoukou,
	YakuSanKantsu:      (*_handInfo).sanKantsu,
	YakuTanyao:         (*_handInfo).tanyao,
	YakuChanta:         (*_handInfo).chanta,
	YakuJunchan:        (*_handInfo).junchan,
	YakuHonroutou:      (*_handInfo).honroutou,
	YakuShousangen:     (*_handInfo).shousangen,
	YakuHonitsu:        (*_handInfo).honitsu,
	YakuChinitsu:       (*_handInfo).chinitsu,
}

// 檢測不是役滿的役種
// 結果未排序
// *計算前必須設置順子牌和刻子牌
func findNormalYaku(hi *_handInfo, isNaki bool) (yakuTypes []int) {
	var yakuHanMap _yakuHanMap
	if !isNaki {
		yakuHanMap = YakuHanMap
	} else {
		yakuHanMap = NakiYakuHanMap
	}

	for yakuType := range yakuHanMap {
		if checker, ok := yakuCheckerMap[yakuType]; ok {
			if checker(hi) {
				yakuTypes = append(yakuTypes, yakuType)
			}
		}
	}

	if considerOldYaku {
		if !isNaki {
			yakuHanMap = OldYakuHanMap
		} else {
			yakuHanMap = OldNakiYakuHanMap
		}

		for yakuType := range yakuHanMap {
			if checker, ok := oldYakuCheckerMap[yakuType]; ok {
				if checker(hi) {
					yakuTypes = append(yakuTypes, yakuType)
				}
			}
		}
	}

	// 役牌單獨算（連風算兩個）
	numYakuhai := hi.numYakuhai()
	for i := 0; i < numYakuhai; i++ {
		yakuTypes = append(yakuTypes, YakuYakuhai)
	}

	return
}

// 尋找役種
// 結果未排序
func findYakuTypes(hi *_handInfo, isNaki bool) (yakuTypes []int) {
	// *計算役種前必須設置順子牌和刻子牌
	hi.allShuntsuFirstTiles = hi.getAllShuntsuFirstTiles()
	hi.allKotsuTiles = hi.getAllKotsuTiles()

	if considerOldYaku {
		sort.Ints(hi.allShuntsuFirstTiles)
		sort.Ints(hi.allKotsuTiles)
	}

	// 先檢測是否有役滿，存在役滿直接 return
	if yakuTypes = findYakumanTypes(hi, isNaki); len(yakuTypes) > 0 {
		return
	}

	return findNormalYaku(hi, isNaki)
}
