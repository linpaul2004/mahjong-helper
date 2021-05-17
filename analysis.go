package main

import (
	"github.com/EndlessCheng/mahjong-helper/util"
	"fmt"
	"strings"
	"github.com/fatih/color"
	"github.com/EndlessCheng/mahjong-helper/util/model"
)

func simpleBestDiscardTile(playerInfo *model.PlayerInfo) int {
	shanten, results14, incShantenResults14 := util.CalculateShantenWithImproves14(playerInfo)
	bestAttackDiscardTile := -1
	if len(results14) > 0 {
		bestAttackDiscardTile = results14[0].DiscardTile
	} else if len(incShantenResults14) > 0 {
		bestAttackDiscardTile = incShantenResults14[0].DiscardTile
	} else {
		return -1
	}
	if shanten == 1 && len(playerInfo.DiscardTiles) < 9 && len(results14) > 0 && len(incShantenResults14) > 0 && !playerInfo.IsNaki() { // 鳴牌時的向聽倒退暫不考慮
		if results14[0].Result13.Waits.AllCount() < 9 && results14[0].Result13.MixedWaitsScore < incShantenResults14[0].Result13.MixedWaitsScore {
			bestAttackDiscardTile = incShantenResults14[0].DiscardTile
		}
	}
	return bestAttackDiscardTile
}

// TODO: 重構至 model
func humanMeld(meld model.Meld) string {
	humanMeld := util.TilesToStr(meld.Tiles)
	if meld.MeldType == model.MeldTypeAnkan {
		return strings.ToUpper(humanMeld)
	}
	return humanMeld
}
func humanHands(playerInfo *model.PlayerInfo) string {
	humanHands := util.Tiles34ToStr(playerInfo.HandTiles34)
	if len(playerInfo.Melds) > 0 {
		humanHands += " " + model.SepMeld
		for i := len(playerInfo.Melds) - 1; i >= 0; i-- {
			humanHands += " " + humanMeld(playerInfo.Melds[i])
		}
	}
	return humanHands
}

func analysisPlayerWithRisk(playerInfo *model.PlayerInfo, mixedRiskTable riskTable) error {
	// 手牌
	humanTiles := humanHands(playerInfo)
	fmt.Println(humanTiles)
	fmt.Println(strings.Repeat("=", len(humanTiles)))

	countOfTiles := util.CountOfTiles34(playerInfo.HandTiles34)
	switch countOfTiles % 3 {
	case 1:
		result := util.CalculateShantenWithImproves13(playerInfo)
		fmt.Println("當前" + util.NumberToChineseShanten(result.Shanten) + "：")
		r := &analysisResult{
			discardTile34:  -1,
			result13:       result,
			mixedRiskTable: mixedRiskTable,
		}
		r.printWaitsWithImproves13_oneRow()
	case 2:
		// 分析手牌
		shanten, results14, incShantenResults14 := util.CalculateShantenWithImproves14(playerInfo)

		// 提示信息
		if shanten == -1 {
			color.HiRed("【已和牌】")
		} else if shanten == 0 {
			if len(results14) > 0 {
				r13 := results14[0].Result13
				if r13.RiichiPoint > 0 && r13.FuritenRate == 0 && r13.DamaPoint >= 5200 && r13.DamaWaits.AllCount() == r13.Waits.AllCount() {
					color.HiGreen("默聽打點充足：追求和率默聽，追求打點立直")
				}
				// 局收支相近時，提示：局收支相近，追求和率打xx，追求打點打xx
			}
		} else if shanten == 1 {
			// 早巡中巡門清時，提醒向聽倒退
			if len(playerInfo.DiscardTiles) < 9 && !playerInfo.IsNaki() {
				alertBackwardToShanten2(results14, incShantenResults14)
			}
		}

		// TODO: 接近流局時提示河底是哪家

		// 何切分析結果
		printResults14WithRisk(results14, mixedRiskTable)
		printResults14WithRisk(incShantenResults14, mixedRiskTable)
	default:
		err := fmt.Errorf("參數錯誤: %d 張牌", countOfTiles)
		if debugMode {
			panic(err)
		}
		return err
	}

	fmt.Println()
	return nil
}

// 分析鳴牌
// playerInfo: 自家信息
// targetTile34: 他家捨牌
// isRedFive: 此捨牌是否為赤5
// allowChi: 是否能吃
// mixedRiskTable: 危險度表
func analysisMeld(playerInfo *model.PlayerInfo, targetTile34 int, isRedFive bool, allowChi bool, mixedRiskTable riskTable) error {
	if handsCount := util.CountOfTiles34(playerInfo.HandTiles34); handsCount%3 != 1 {
		return fmt.Errorf("手牌錯誤：%d 張牌 %v", handsCount, playerInfo.HandTiles34)
	}
	// 原始手牌分析
	result := util.CalculateShantenWithImproves13(playerInfo)
	// 副露分析
	shanten, results14, incShantenResults14 := util.CalculateMeld(playerInfo, targetTile34, isRedFive, allowChi)
	if len(results14) == 0 && len(incShantenResults14) == 0 {
		return nil // fmt.Errorf("輸入錯誤：無法鳴這張牌")
	}

	// 鳴牌
	humanTiles := humanHands(playerInfo)
	handsTobeNaki := humanTiles + " " + model.SepTargetTile + " " + util.Tile34ToStr(targetTile34) + "?"
	fmt.Println(handsTobeNaki)
	fmt.Println(strings.Repeat("=", len(handsTobeNaki)))

	// 原始手牌分析結果
	fmt.Println("當前" + util.NumberToChineseShanten(result.Shanten) + "：")
	r := &analysisResult{
		discardTile34:  -1,
		result13:       result,
		mixedRiskTable: mixedRiskTable,
	}
	r.printWaitsWithImproves13_oneRow()

	// 提示信息
	// TODO: 局收支相近時，提示：局收支相近，追求和率打xx，追求打點打xx
	if shanten == -1 {
		color.HiRed("【已和牌】")
	} else if shanten <= 1 {
		// 鳴牌後聽牌或一向聽，提示型聽
		if len(results14) > 0 && results14[0].LeftDrawTilesCount > 0 && results14[0].LeftDrawTilesCount <= 16 {
			color.HiGreen("考慮型聽？")
		}
	}

	// TODO: 接近流局時提示河底是哪家

	// 鳴牌何切分析結果
	printResults14WithRisk(results14, mixedRiskTable)
	printResults14WithRisk(incShantenResults14, mixedRiskTable)
	return nil
}

func analysisHumanTiles(humanTilesInfo *model.HumanTilesInfo) (playerInfo *model.PlayerInfo, err error) {
	defer func() {
		if er := recover(); er != nil {
			err = er.(error)
		}
	}()

	if err = humanTilesInfo.SelfParse(); err != nil {
		return
	}

	tiles34, numRedFives, err := util.StrToTiles34(humanTilesInfo.HumanTiles)
	if err != nil {
		return
	}

	tileCount := util.CountOfTiles34(tiles34)
	if tileCount > 14 {
		return nil, fmt.Errorf("輸入錯誤：%d 張牌", tileCount)
	}

	if tileCount%3 == 0 {
		color.HiYellow("%s 是 %d 張牌\n助手隨機補了一張牌", humanTilesInfo.HumanTiles, tileCount)
		util.RandomAddTile(tiles34)
	}

	melds := []model.Meld{}
	for _, humanMeld := range humanTilesInfo.HumanMelds {
		tiles, _numRedFives, er := util.StrToTiles(humanMeld)
		if er != nil {
			return nil, er
		}
		isUpper := humanMeld[len(humanMeld)-1] <= 'Z'
		var meldType int
		switch {
		case len(tiles) == 3 && tiles[0] != tiles[1]:
			meldType = model.MeldTypeChi
		case len(tiles) == 3 && tiles[0] == tiles[1]:
			meldType = model.MeldTypePon
		case len(tiles) == 4 && isUpper:
			meldType = model.MeldTypeAnkan
		case len(tiles) == 4 && !isUpper:
			meldType = model.MeldTypeMinkan
		default:
			return nil, fmt.Errorf("輸入錯誤: %s", humanMeld)
		}
		containRedFive := false
		for i, c := range _numRedFives {
			if c > 0 {
				containRedFive = true
				numRedFives[i] += c
			}
		}
		melds = append(melds, model.Meld{
			MeldType:       meldType,
			Tiles:          tiles,
			ContainRedFive: containRedFive,
		})
	}

	playerInfo = model.NewSimplePlayerInfo(tiles34, melds)
	playerInfo.NumRedFives = numRedFives

	if humanTilesInfo.HumanDoraTiles != "" {
		playerInfo.DoraTiles, _, err = util.StrToTiles(humanTilesInfo.HumanDoraTiles)
		if err != nil {
			return
		}
	}

	if humanTilesInfo.HumanTargetTile != "" {
		if tileCount%3 == 2 {
			return nil, fmt.Errorf("輸入錯誤: %s 是 %d 張牌", humanTilesInfo.HumanTiles, tileCount)
		}
		targetTile34, isRedFive, er := util.StrToTile34(humanTilesInfo.HumanTargetTile)
		if er != nil {
			return nil, er
		}
		if er := analysisMeld(playerInfo, targetTile34, isRedFive, true, nil); er != nil {
			return nil, er
		}
		return
	}

	playerInfo.IsTsumo = humanTilesInfo.IsTsumo
	err = analysisPlayerWithRisk(playerInfo, nil)
	return
}
