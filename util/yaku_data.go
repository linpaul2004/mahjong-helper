package util

import (
	"fmt"
	"sort"
)

var considerOldYaku bool

func SetConsiderOldYaku(b bool) {
	considerOldYaku = b
}

//

const (
	// https://en.wikipedia.org/wiki/Japanese_Mahjong_yaku
	// Special criteria
	YakuRiichi int = iota
	YakuChiitoi

	// Yaku based on luck
	YakuTsumo
	//YakuIppatsu
	//YakuHaitei
	//YakuHoutei
	//YakuRinshan
	//YakuChankan
	YakuDaburii

	// Yaku based on sequences
	YakuPinfu
	YakuRyanpeikou
	YakuIipeikou
	YakuSanshokuDoujun  // *
	YakuIttsuu          // *

	// Yaku based on triplets and/or quads
	YakuToitoi
	YakuSanAnkou
	YakuSanshokuDoukou
	YakuSanKantsu

	// Yaku based on terminal or honor tiles
	YakuTanyao
	YakuYakuhai
	YakuChanta     // * 必須有順子
	YakuJunchan    // * 必須有順子
	YakuHonroutou  // 七對也算
	YakuShousangen

	// Yaku based on suits
	YakuHonitsu   // *
	YakuChinitsu  // *

	// Yakuman
	//YakuKokushi
	//YakuKokushi13
	YakuSuuAnkou
	YakuSuuAnkouTanki
	YakuDaisangen
	YakuShousuushii
	YakuDaisuushii
	YakuTsuuiisou
	YakuChinroutou
	YakuRyuuiisou
	YakuChuuren
	YakuChuuren9
	YakuSuuKantsu
	//YakuTenhou
	//YakuChiihou

	// 古役
	YakuShiiaruraotai
	YakuUumensai
	YakuSanrenkou
	YakuIsshokusanjun

	// 古役役滿
	YakuDaisuurin
	YakuDaisharin
	YakuDaichikurin
	YakuDaichisei

	//_endYakuType  // 標記 enum 結束，方便計算有多少個 YakuType
)

//const maxYakuType = _endYakuType

var YakuNameMap = map[int]string{
	// Special criteria
	YakuRiichi:  "立直",
	YakuChiitoi: "七對",

	// Yaku based on luck
	YakuTsumo: "自摸",
	//YakuIppatsu: "一發",
	//YakuHaitei:  "海底",
	//YakuHoutei:  "河底",
	//YakuRinshan: "嶺上",
	//YakuChankan: "搶槓",
	YakuDaburii: "w立",

	// Yaku based on sequences
	YakuPinfu:          "平和",
	YakuRyanpeikou:     "兩盃口",
	YakuIipeikou:       "一盃口",
	YakuSanshokuDoujun: "三色",
	YakuIttsuu:         "一通", // 一氣

	// Yaku based on triplets and/or quads
	YakuToitoi:         "對對",
	YakuSanAnkou:       "三暗刻",
	YakuSanshokuDoukou: "三色同刻",
	YakuSanKantsu:      "三槓子",

	// Yaku based on terminal or honor tiles
	YakuTanyao:     "斷幺",
	YakuYakuhai:    "役牌",
	YakuChanta:     "混全",
	YakuJunchan:    "純全",
	YakuHonroutou:  "混老頭", // 七對也算
	YakuShousangen: "小三元",

	// Yaku based on suits
	YakuHonitsu:  "混一色",
	YakuChinitsu: "清一色",

	// Yakuman
	//YakuKokushi:       "國士",
	//YakuKokushi13:     "國士十三面",
	YakuSuuAnkou:      "四暗刻",
	YakuSuuAnkouTanki: "四暗刻單騎",
	YakuDaisangen:     "大三元",
	YakuShousuushii:   "小四喜",
	YakuDaisuushii:    "大四喜",
	YakuTsuuiisou:     "字一色",
	YakuChinroutou:    "清老頭",
	YakuRyuuiisou:     "綠一色",
	YakuChuuren:       "九蓮",
	YakuChuuren9:      "純正九蓮",
	YakuSuuKantsu:     "四槓子",
	//YakuTenhou:        "天和",
	//YakuChiihou:       "地和",
}

var OldYakuNameMap = map[int]string{
	YakuShiiaruraotai: "十二落擡",
	YakuUumensai:      "五門齊",
	YakuSanrenkou:     "三連刻",
	YakuIsshokusanjun: "一色三順",

	YakuDaisuurin:   "大數鄰",
	YakuDaisharin:   "大車輪",
	YakuDaichikurin: "大竹林",
	YakuDaichisei:   "大七星",
}

func YakuTypesToStr(yakuTypes []int) string {
	if len(yakuTypes) == 0 {
		return "[無役]"
	}
	names := []string{}
	for _, t := range yakuTypes {
		if name, ok := YakuNameMap[t]; ok {
			names = append(names, name)
		}
	}

	if considerOldYaku {
		for _, t := range yakuTypes {
			if name, ok := OldYakuNameMap[t]; ok {
				names = append(names, name)
			}
		}
	}

	return fmt.Sprint(names)
}

func YakuTypesWithDoraToStr(yakuTypes map[int]struct{}, numDora int) string {
	if len(yakuTypes) == 0 {
		return "[無役]"
	}
	yt := []int{}
	for t := range yakuTypes {
		yt = append(yt, t)
	}
	sort.Ints(yt)
	names := []string{}
	for _, t := range yt {
		names = append(names, YakuNameMap[t])
	}
	// TODO: old yaku
	if numDora > 0 {
		names = append(names, fmt.Sprintf("寶牌%d", numDora))
	}
	return fmt.Sprint(names)
}

//

type _yakuHanMap map[int]int
type _yakumanTimesMap map[int]int

var YakuHanMap = _yakuHanMap{
	YakuRiichi:  1,
	YakuChiitoi: 2,

	YakuTsumo: 1,
	//YakuIppatsu: 1,
	//YakuHaitei:  1,
	//YakuHoutei:  1,
	//YakuRinshan: 1,
	//YakuChankan: 1,
	YakuDaburii: 2,

	YakuPinfu:          1,
	YakuRyanpeikou:     3,
	YakuIipeikou:       1,
	YakuSanshokuDoujun: 2,
	YakuIttsuu:         2,

	YakuToitoi:         2,
	YakuSanAnkou:       2,
	YakuSanshokuDoukou: 2,
	YakuSanKantsu:      2,

	YakuTanyao:     1,
	YakuYakuhai:    1,
	YakuChanta:     2,
	YakuJunchan:    3,
	YakuHonroutou:  2,
	YakuShousangen: 2,

	YakuHonitsu:  3,
	YakuChinitsu: 6,
}

var NakiYakuHanMap = _yakuHanMap{
	//YakuHaitei:  1,
	//YakuHoutei:  1,
	//YakuRinshan: 1,
	//YakuChankan: 1,

	YakuSanshokuDoujun: 1,
	YakuIttsuu:         1,

	YakuToitoi:         2,
	YakuSanAnkou:       2,
	YakuSanshokuDoukou: 2,
	YakuSanKantsu:      2,

	YakuTanyao:     1,
	YakuYakuhai:    1,
	YakuChanta:     1,
	YakuJunchan:    2,
	YakuHonroutou:  2,
	YakuShousangen: 2,

	YakuHonitsu:  2,
	YakuChinitsu: 5,
}

var OldYakuHanMap = _yakuHanMap{
	YakuUumensai:      2,
	YakuSanrenkou:     2,
	YakuIsshokusanjun: 3,
}

var OldNakiYakuHanMap = _yakuHanMap{
	YakuShiiaruraotai: 1, // 四副露大吊車
	YakuUumensai:      2,
	YakuSanrenkou:     2,
	YakuIsshokusanjun: 2,
}

// 計算 yakuTypes(非役滿) 累積的番數
func CalcYakuHan(yakuTypes []int, isNaki bool) (cntHan int) {
	var yakuHanMap _yakuHanMap
	if !isNaki {
		yakuHanMap = YakuHanMap
	} else {
		yakuHanMap = NakiYakuHanMap
	}

	for _, yakuType := range yakuTypes {
		if han, ok := yakuHanMap[yakuType]; ok {
			cntHan += han
		}
	}

	if considerOldYaku {
		if !isNaki {
			yakuHanMap = OldYakuHanMap
		} else {
			yakuHanMap = OldNakiYakuHanMap
		}

		for _, yakuType := range yakuTypes {
			if han, ok := yakuHanMap[yakuType]; ok {
				cntHan += han
			}
		}
	}

	return
}

//

var YakumanTimesMap = map[int]int{
	//YakuKokushi:       1,
	//YakuKokushi13:     2,
	YakuSuuAnkou:      1,
	YakuSuuAnkouTanki: 2,
	YakuDaisangen:     1,
	YakuShousuushii:   1,
	YakuDaisuushii:    2,
	YakuTsuuiisou:     1,
	YakuChinroutou:    1,
	YakuRyuuiisou:     1,
	YakuChuuren:       1,
	YakuChuuren9:      2,
	YakuSuuKantsu:     1,
	//YakuTenhou:        1,
	//YakuChiihou:       1,
}

var NakiYakumanTimesMap = map[int]int{
	YakuDaisangen:   1,
	YakuShousuushii: 1,
	YakuDaisuushii:  2,
	YakuTsuuiisou:   1,
	YakuChinroutou:  1,
	YakuRyuuiisou:   1,
	YakuSuuKantsu:   1,
}

var OldYakumanTimesMap = map[int]int{
	YakuDaisuurin:   1,
	YakuDaisharin:   1,
	YakuDaichikurin: 1,
	YakuDaichisei:   1, // 複合字一色，實際為兩倍役滿
}

// 計算役滿倍數
func CalcYakumanTimes(yakuTypes []int, isNaki bool) (times int) {
	var yakumanTimesMap _yakumanTimesMap
	if !isNaki {
		yakumanTimesMap = YakumanTimesMap
	} else {
		yakumanTimesMap = NakiYakumanTimesMap
	}

	for _, yakuman := range yakuTypes {
		if t, ok := yakumanTimesMap[yakuman]; ok {
			times += t
		}
	}

	if considerOldYaku && !isNaki {
		for _, yakuman := range yakuTypes {
			if t, ok := OldYakumanTimesMap[yakuman]; ok {
				times += t
			}
		}
	}

	return
}
