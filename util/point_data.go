package util

// 修正立直打點，即考慮自摸、裏寶和一發的實際打點
// 參考:「統計學」のマージャン戦術
func (pr *PointResult) fixedRiichiPoint(isFuriten bool) float64 {
	ronPoint := float64(pr.Point)
	if pr.isParent {
		ronPoint /= 1.5
	}

	pt := ronPoint - 100 // 保證親子落入同一區間
	switch {
	case pt <= 1300:
		ronPoint *= 2730.0 / 1300.0
	case pt <= 2000:
		ronPoint *= 3700.0 / 2000.0
	case pt <= 2600:
		ronPoint *= 4900.0 / 2600.0
	case pt <= 3900:
		ronPoint *= 6300.0 / 3900.0
	case pt <= 5200:
		ronPoint *= 7500.0 / 5200.0
	case pt <= 7700:
		ronPoint *= 9100.0 / 7700.0
	case pt <= 8000:
		if pr.han == 4 {
			ronPoint *= 9130.0 / 8000.0
		} else if pr.han == 5 {
			ronPoint *= 11200.0 / 8000.0
		}
	case pt <= 12000:
		if pr.han == 6 {
			ronPoint *= 13030.0 / 12000.0
		} else if pr.han == 7 {
			ronPoint *= 15000.0 / 12000.0
		}
	default:
		// TODO: 跳滿以上的立直打點調整
	}

	if isFuriten {
		// 振聽時由於只能自摸，打點要略高些
		const furitenRiichiPointMulti = 1.1
		ronPoint *= furitenRiichiPointMulti
	}

	if pr.isParent {
		ronPoint *= 1.5
	}
	return ronPoint
}

//

// 子家榮和點數均值
// 參考：「統計學」のマージャン戦術
// 親家按 x1.5 算
// TODO: 剩餘 dora 數對失點的影響
const (
	RonPointRiichiHiIppatsu = 5172.0 // 基準
	RonPointRiichiIppatsu   = 7445.0
	//RonPointHonitsu         = 6603.0
	//RonPointToitoi          = 7300.0
	RonPointOtherNaki = 3000.0 // *fixed
	// TODO: 考慮雙東的影響
	RonPointDama = 4536.0
)

// 簡單地判斷子家副露者的打點
// dora point han
// 0    3000  1-3
// 1    4200  2-4
// 2    5880  3-5
// 3    8232  4-6
// 4    10000 5-7
// 5    13000 6-8
// 親家按 x1.5 算
// TODO: 暗槓對打點的提升？
func RonPointOtherNakiWithDora(doraCount int) (point float64) {
	point = RonPointOtherNaki
	const doraMulti = 1.4 // TODO: 待調整？
	for i := 0; i < MinInt(3, doraCount); i++ {
		point *= doraMulti
	}

	doraCount -= 3
	if doraCount <= 0 {
		return point
	}

	const doraMulti2 = 1.3 // TODO: 待調整？
	for i := 0; i < MinInt(2, doraCount); i++ {
		point *= doraMulti2
	}

	return point
}
