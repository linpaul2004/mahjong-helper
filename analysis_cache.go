package main

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/fatih/color"
)

type analysisOpType int

const (
	analysisOpTypeTsumo     analysisOpType = iota
	analysisOpTypeChiPonKan  // 吃 碰 明槓
	analysisOpTypeKan        // 加槓 暗槓
)

// TODO: 提醒「此處應該副露，不應跳過」

type analysisCache struct {
	analysisOpType analysisOpType

	selfDiscardTile int
	//isSelfDiscardRedFive bool
	selfDiscardTileRisk float64
	isRiichiWhenDiscard bool
	meldType            int

	// 用手牌中的什麼牌去鳴牌，空就是跳過不鳴
	selfOpenTiles []int

	aiAttackDiscardTile      int
	aiDefenceDiscardTile     int
	aiAttackDiscardTileRisk  float64
	aiDefenceDiscardTileRisk float64

	tenpaiRate []float64 // TODO: 三家聽牌率
}

type roundAnalysisCache struct {
	isStart bool
	isEnd   bool
	cache   []*analysisCache

	analysisCacheBeforeChiPon *analysisCache
}

func (rc *roundAnalysisCache) print() {
	const (
		baseInfo  = "助手正在計算推薦捨牌，請稍等……（計算結果僅供參考）"
		emptyInfo = "--"
		sep       = "  "
	)

	done := rc != nil && rc.isEnd
	if !done {
		color.HiGreen(baseInfo)
	} else {
		// 檢查最後的是否自摸，若為自摸則去掉推薦
		if len(rc.cache) > 0 {
			latestCache := rc.cache[len(rc.cache)-1]
			if latestCache.selfDiscardTile == -1 {
				latestCache.aiAttackDiscardTile = -1
				latestCache.aiDefenceDiscardTile = -1
			}
		}
	}

	fmt.Print("巡目　　")
	if done {
		for i := range rc.cache {
			fmt.Printf("%s%2d", sep, i+1)
		}
	}
	fmt.Println()

	printTileInfo := func(tile int, risk float64, suffix string) {
		info := emptyInfo
		if tile != -1 {
			info = util.Mahjong[tile]
		}
		fmt.Print(sep)
		if info == emptyInfo || risk < 5 {
			fmt.Print(info)
		} else {
			color.New(getNumRiskColor(risk)).Print(info)
		}
		fmt.Print(suffix)
	}

	fmt.Print("自家切牌")
	if done {
		for i, c := range rc.cache {
			suffix := ""
			if c.isRiichiWhenDiscard {
				suffix = "[立直]"
			} else if c.selfDiscardTile == -1 && i == len(rc.cache)-1 {
				//suffix = "[自摸]"
				// TODO: 流局
			}
			printTileInfo(c.selfDiscardTile, c.selfDiscardTileRisk, suffix)
		}
	}
	fmt.Println()

	fmt.Print("進攻推薦")
	if done {
		for _, c := range rc.cache {
			printTileInfo(c.aiAttackDiscardTile, c.aiAttackDiscardTileRisk, "")
		}
	}
	fmt.Println()

	fmt.Print("防守推薦")
	if done {
		for _, c := range rc.cache {
			printTileInfo(c.aiDefenceDiscardTile, c.aiDefenceDiscardTileRisk, "")
		}
	}
	fmt.Println()

	fmt.Println()
}

// （摸牌後、鳴牌後的）實際捨牌
func (rc *roundAnalysisCache) addSelfDiscardTile(tile int, risk float64, isRiichiWhenDiscard bool) {
	latestCache := rc.cache[len(rc.cache)-1]
	latestCache.selfDiscardTile = tile
	latestCache.selfDiscardTileRisk = risk
	latestCache.isRiichiWhenDiscard = isRiichiWhenDiscard
}

// 摸牌時的切牌推薦
func (rc *roundAnalysisCache) addAIDiscardTileWhenDrawTile(attackTile int, defenceTile int, attackTileRisk float64, defenceDiscardTileRisk float64) {
	// 摸牌，巡目+1
	rc.cache = append(rc.cache, &analysisCache{
		analysisOpType:           analysisOpTypeTsumo,
		selfDiscardTile:          -1,
		aiAttackDiscardTile:      attackTile,
		aiDefenceDiscardTile:     defenceTile,
		aiAttackDiscardTileRisk:  attackTileRisk,
		aiDefenceDiscardTileRisk: defenceDiscardTileRisk,
	})
	rc.analysisCacheBeforeChiPon = nil
}

// 加槓 暗槓
func (rc *roundAnalysisCache) addKan(meldType int) {
	// latestCache 是摸牌
	latestCache := rc.cache[len(rc.cache)-1]
	latestCache.analysisOpType = analysisOpTypeKan
	latestCache.meldType = meldType
	// 槓完之後又會摸牌，巡目+1
}

// 吃 碰 明槓
func (rc *roundAnalysisCache) addChiPonKan(meldType int) {
	if meldType == meldTypeMinkan {
		// 暫時忽略明槓，巡目不+1，留給摸牌時+1
		return
	}
	// 巡目+1
	var newCache *analysisCache
	if rc.analysisCacheBeforeChiPon != nil {
		newCache = rc.analysisCacheBeforeChiPon // 見 addPossibleChiPonKan
		newCache.analysisOpType = analysisOpTypeChiPonKan
		newCache.meldType = meldType
		rc.analysisCacheBeforeChiPon = nil
	} else {
		// 此處代碼應該不會觸發
		if debugMode {
			panic("rc.analysisCacheBeforeChiPon == nil")
		}
		newCache = &analysisCache{
			analysisOpType:       analysisOpTypeChiPonKan,
			selfDiscardTile:      -1,
			aiAttackDiscardTile:  -1,
			aiDefenceDiscardTile: -1,
			meldType:             meldType,
		}
	}
	rc.cache = append(rc.cache, newCache)
}

// 吃 碰 槓 跳過
func (rc *roundAnalysisCache) addPossibleChiPonKan(attackTile int, attackTileRisk float64) {
	rc.analysisCacheBeforeChiPon = &analysisCache{
		analysisOpType:          analysisOpTypeChiPonKan,
		selfDiscardTile:         -1,
		aiAttackDiscardTile:     attackTile,
		aiDefenceDiscardTile:    -1,
		aiAttackDiscardTileRisk: attackTileRisk,
	}
}

//

type gameAnalysisCache struct {
	// 局數 本場數
	wholeGameCache [][]*roundAnalysisCache

	majsoulRecordUUID string

	selfSeat int
}

func newGameAnalysisCache(majsoulRecordUUID string, selfSeat int) *gameAnalysisCache {
	cache := make([][]*roundAnalysisCache, 3*4) // 最多到西四
	for i := range cache {
		cache[i] = make([]*roundAnalysisCache, 100) // 最多連莊
	}
	return &gameAnalysisCache{
		wholeGameCache:    cache,
		majsoulRecordUUID: majsoulRecordUUID,
		selfSeat:          selfSeat,
	}
}

//

// TODO: 重構成 struct
var (
	_analysisCacheList = make([]*gameAnalysisCache, 4)
	_currentSeat       int
)

func resetAnalysisCache() {
	_analysisCacheList = make([]*gameAnalysisCache, 4)
}

func setAnalysisCache(analysisCache *gameAnalysisCache) {
	_analysisCacheList[analysisCache.selfSeat] = analysisCache
	_currentSeat = analysisCache.selfSeat
}

func getAnalysisCache(seat int) *gameAnalysisCache {
	if seat == -1 {
		return nil
	}
	return _analysisCacheList[seat]
}

func getCurrentAnalysisCache() *gameAnalysisCache {
	return getAnalysisCache(_currentSeat)
}

func (c *gameAnalysisCache) runMajsoulRecordAnalysisTask(actions majsoulRoundActions) error {
	// 從第一個 action 中取出局和場
	if len(actions) == 0 {
		return fmt.Errorf("數據異常：此局數據為空")
	}

	newRoundAction := actions[0]
	data := newRoundAction.Action
	roundNumber := 4*(*data.Chang) + *data.Ju
	ben := *data.Ben
	roundCache := c.wholeGameCache[roundNumber][ben] // TODO: 建議用原子操作
	if roundCache == nil {
		roundCache = &roundAnalysisCache{isStart: true}
		if debugMode {
			fmt.Println("助手正在計算推薦捨牌…… 創建 roundCache")
		}
		c.wholeGameCache[roundNumber][ben] = roundCache
	} else if roundCache.isStart {
		if debugMode {
			fmt.Println("無需重複計算")
		}
		return nil
	}

	// 遍歷自家捨牌，找到捨牌前的操作
	// 若為摸牌操作，計算出此時的 AI 進攻捨牌和防守捨牌
	// 若為鳴牌操作，計算出此時的 AI 進攻捨牌（無進攻捨牌則設為 -1），防守捨牌設為 -1
	// TODO: 玩家跳過，但是 AI 覺得應鳴牌？
	majsoulRoundData := &majsoulRoundData{selfSeat: c.selfSeat} // 注意這裏是用的一個新的 majsoulRoundData 去計算的，不會有數據沖突
	majsoulRoundData.roundData = newGame(majsoulRoundData)
	majsoulRoundData.roundData.gameMode = gameModeRecordCache
	majsoulRoundData.skipOutput = true
	for i, action := range actions[:len(actions)-1] {
		if c.majsoulRecordUUID != getMajsoulCurrentRecordUUID() {
			if debugMode {
				fmt.Println("用戶退出該牌譜")
			}
			// 提前退出，減少不必要的計算
			return nil
		}
		if debugMode {
			fmt.Println("助手正在計算推薦捨牌…… action", i)
		}
		majsoulRoundData.msg = action.Action
		majsoulRoundData.analysis()
	}
	roundCache.isEnd = true

	if c.majsoulRecordUUID != getMajsoulCurrentRecordUUID() {
		if debugMode {
			fmt.Println("用戶退出該牌譜")
		}
		return nil
	}

	clearConsole()
	roundCache.print()

	return nil
}
