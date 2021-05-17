package main

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"github.com/fatih/color"
)

type DataParser interface {
	// 數據來源（是天鳳還是雀魂）
	GetDataSourceType() int

	// 獲取自家初始座位：0-第一局的東家 1-第一局的南家 2-第一局的西家 3-第一局的北家
	// 僅處理雀魂數據，天鳳返回 -1
	GetSelfSeat() int

	// 原始 JSON
	GetMessage() string

	// 解析前，根據消息內容來決定是否要進行後續解析
	SkipMessage() bool

	// 嘗試解析用戶名
	IsLogin() bool
	HandleLogin()

	// round 開始/重連
	// roundNumber: 場數（如東1為0，東2為1，...，南1為4，...，南4為7，...），對於三麻來說南1也是4
	// benNumber: 本場數
	// dealer: 莊家 0-3
	// doraIndicators: 寶牌指示牌
	// handTiles: 手牌
	// numRedFives: 按照 mps 的順序，赤5個數
	IsInit() bool
	ParseInit() (roundNumber int, benNumber int, dealer int, doraIndicators []int, handTiles []int, numRedFives []int)

	// 自家摸牌
	// tile: 0-33
	// isRedFive: 是否為赤5
	// kanDoraIndicator: 摸牌時，若為暗槓摸的嶺上牌，則可以翻出槓寶牌指示牌，否則返回 -1（目前恆為 -1，見 IsNewDora）
	IsSelfDraw() bool
	ParseSelfDraw() (tile int, isRedFive bool, kanDoraIndicator int)

	// 捨牌
	// who: 0=自家, 1=下家, 2=對家, 3=上家
	// isTsumogiri: 是否為摸切（who=0 時忽略該值）
	// isReach: 是否為立直宣言（isReach 對於天鳳來說恆為 false，見 IsReach）
	// canBeMeld: 是否可以鳴牌（who=0 時忽略該值）
	// kanDoraIndicator: 大明槓/加槓的槓寶牌指示牌，在切牌後出現，沒有則返回 -1（天鳳恆為-1，見 IsNewDora）
	IsDiscard() bool
	ParseDiscard() (who int, discardTile int, isRedFive bool, isTsumogiri bool, isReach bool, canBeMeld bool, kanDoraIndicator int)

	// 鳴牌（含暗槓、加槓）
	// kanDoraIndicator: 暗槓的槓寶牌指示牌，在他家暗槓時出現，沒有則返回 -1（天鳳恆為-1，見 IsNewDora）
	IsOpen() bool
	ParseOpen() (who int, meld *model.Meld, kanDoraIndicator int)

	// 立直聲明（IsReach 對於雀魂來說恆為 false，見 ParseDiscard）
	IsReach() bool
	ParseReach() (who int)

	// 振聽
	IsFuriten() bool

	// 本局是否和牌
	IsRoundWin() bool
	ParseRoundWin() (whos []int, points []int)

	// 是否流局
	// 四風連打 四家立直 四槓散了 九種九牌 三家和了 | 流局聽牌 流局未聽牌 | 流局滿貫
	// 三家和了
	IsRyuukyoku() bool
	ParseRyuukyoku() (type_ int, whos []int, points []int)

	// 拔北寶牌
	IsNukiDora() bool
	ParseNukiDora() (who int, isTsumogiri bool)

	// 這一項放在末尾處理
	// 槓寶牌（雀魂在暗槓後的摸牌時出現）
	// kanDoraIndicator: 0-33
	IsNewDora() bool
	ParseNewDora() (kanDoraIndicator int)
}

type playerInfo struct {
	name string // 自家/下家/對家/上家

	selfWindTile int // 自風

	melds                []*model.Meld // 副露
	meldDiscardsAtGlobal []int
	meldDiscardsAt       []int
	isNaki               bool // 是否鳴牌（暗槓不算鳴牌）

	// 注意負數（自摸切）要^
	discardTiles          []int // 該玩家的捨牌
	latestDiscardAtGlobal int   // 該玩家最近一次捨牌在 globalDiscardTiles 中的下標，初始為 -1
	earlyOutsideTiles     []int // 立直前的1-5巡的外側牌

	isReached  bool // 是否立直
	canIppatsu bool // 是否有一發

	reachTileAtGlobal int // 立直宣言牌在 globalDiscardTiles 中的下標，初始為 -1
	reachTileAt       int // 立直宣言牌在 discardTiles 中的下標，初始為 -1

	nukiDoraNum int // 拔北寶牌數
}

func newPlayerInfo(name string, selfWindTile int) *playerInfo {
	return &playerInfo{
		name:                  name,
		selfWindTile:          selfWindTile,
		latestDiscardAtGlobal: -1,
		reachTileAtGlobal:     -1,
		reachTileAt:           -1,
	}
}

func modifySanninPlayerInfoList(lst []*playerInfo, roundNumber int) []*playerInfo {
	windToIdxMap := map[int]int{}
	for i, pi := range lst {
		windToIdxMap[pi.selfWindTile] = i
	}

	idxS, idxW, idxN := windToIdxMap[28], windToIdxMap[29], windToIdxMap[30]
	switch roundNumber % 4 {
	case 0:
	case 1:
		// 北和西交換
		lst[idxN].selfWindTile, lst[idxW].selfWindTile = lst[idxW].selfWindTile, lst[idxN].selfWindTile
	case 2:
		// 北和西交換，再和南交換
		lst[idxN].selfWindTile, lst[idxW].selfWindTile, lst[idxS].selfWindTile = lst[idxW].selfWindTile, lst[idxS].selfWindTile, lst[idxN].selfWindTile
	default:
		panic("[modifySanninPlayerInfoList] 代碼有誤")
	}
	return lst
}

func (p *playerInfo) doraNum(doraList []int) (doraCount int) {
	for _, meld := range p.melds {
		for _, tile := range meld.Tiles {
			for _, doraTile := range doraList {
				if tile == doraTile {
					doraCount++
				}
			}
		}
		if meld.ContainRedFive {
			doraCount++
		}
	}
	if p.nukiDoraNum > 0 {
		doraCount += p.nukiDoraNum
		// 特殊：西為指示牌
		for _, doraTile := range doraList {
			if doraTile == 30 {
				doraCount += p.nukiDoraNum
			}
		}
	}
	return
}

//

type roundData struct {
	parser DataParser

	gameMode gameMode

	skipOutput bool

	// 玩家數，3 為三麻，4 為四麻
	playerNumber int

	// 場數（如東1為0，東2為1，...，南1為4，...）
	roundNumber int

	// 本場數，從 0 開始算
	benNumber int

	// 場風
	roundWindTile int

	// 莊家 0=自家, 1=下家, 2=對家, 3=上家
	// 請用 reset 設置
	dealer int

	// 寶牌指示牌
	doraIndicators []int

	// 自家手牌
	counts []int

	// 按照 mps 的順序記錄自家赤5數量，包含副露的赤5
	// 比如有 0p 和 0s 就是 [1, 0, 1]
	numRedFives []int

	// 牌山剩餘牌量
	leftCounts []int

	// 全局捨牌
	// 按捨牌順序，負數表示摸切(-)，非負數表示手切(+)
	// 可以理解成：- 表示不要/暗色，+ 表示進張/亮色
	globalDiscardTiles []int

	// 0=自家, 1=下家, 2=對家, 3=上家
	players []*playerInfo
}

func newRoundData(parser DataParser, roundNumber int, benNumber int, dealer int) *roundData {
	// 無論是三麻還是四麻，都視作四個人
	const playerNumber = 4
	roundWindTile := 27 + roundNumber/playerNumber
	playerWindTile := make([]int, playerNumber)
	for i := 0; i < playerNumber; i++ {
		playerWindTile[i] = 27 + (playerNumber-dealer+i)%playerNumber
	}
	return &roundData{
		parser:      parser,
		roundNumber: roundNumber,
		benNumber:   benNumber,

		roundWindTile:      roundWindTile,
		dealer:             dealer,
		counts:             make([]int, 34),
		leftCounts:         util.InitLeftTiles34(),
		globalDiscardTiles: []int{},
		players: []*playerInfo{
			newPlayerInfo("自家", playerWindTile[0]),
			newPlayerInfo("下家", playerWindTile[1]),
			newPlayerInfo("對家", playerWindTile[2]),
			newPlayerInfo("上家", playerWindTile[3]),
		},
	}
}

func newGame(parser DataParser) *roundData {
	return newRoundData(parser, 0, 0, 0)
}

// 新的一局
func (d *roundData) reset(roundNumber int, benNumber int, dealer int) {
	skipOutput := d.skipOutput
	gameMode := d.gameMode
	playerNumber := d.playerNumber
	newData := newRoundData(d.parser, roundNumber, benNumber, dealer)
	newData.skipOutput = skipOutput
	newData.gameMode = gameMode
	newData.playerNumber = playerNumber
	if playerNumber == 3 {
		// 三麻沒有 2-8m
		for i := 1; i <= 7; i++ {
			newData.leftCounts[i] = 0
		}
		newData.players = modifySanninPlayerInfoList(newData.players, roundNumber)
	}
	*d = *newData
}

func (d *roundData) newGame() {
	d.reset(0, 0, 0)
}

func (d *roundData) descLeftCounts(tile int) {
	d.leftCounts[tile]--
	if d.leftCounts[tile] < 0 {
		info := fmt.Sprintf("數據異常: %s 數量為 %d", util.MahjongZH[tile], d.leftCounts[tile])
		if debugMode {
			panic(info)
		} else {
			fmt.Println(info)
		}
	}
}

// 槓！
func (d *roundData) newDora(kanDoraIndicator int) {
	d.doraIndicators = append(d.doraIndicators, kanDoraIndicator)
	d.descLeftCounts(kanDoraIndicator)

	if d.skipOutput {
		return
	}

	color.Yellow("槓寶牌指示牌是 %s", util.MahjongZH[kanDoraIndicator])
}

// 根據寶牌指示牌計算出寶牌
func (d *roundData) doraList() (dl []int) {
	return model.DoraList(d.doraIndicators, d.playerNumber == 3)
}

func (d *roundData) printDiscards() {
	// 三麻的北家是不需要打印的
	for i := len(d.players) - 1; i >= 1; i-- {
		if player := d.players[i]; d.playerNumber != 3 || player.selfWindTile != 30 {
			player.printDiscards()
		}
	}
}

// 分析34種牌的危險度
// 可以用來判斷自家手牌的安全度，以及他家是否在進攻（多次切出危險度高的牌）
func (d *roundData) analysisTilesRisk() (riList riskInfoList) {
	riList = make(riskInfoList, len(d.players))
	for who := range riList {
		riList[who] = &riskInfo{
			playerNumber: d.playerNumber,
			safeTiles34:  make([]bool, 34),
		}
	}

	// 先利用振聽規則收集各家安牌
	for who, player := range d.players {
		if who == 0 {
			// TODO: 暫時不計算自家的
			continue
		}

		// 捨牌振聽産生的安牌
		for _, tile := range normalDiscardTiles(player.discardTiles) {
			riList[who].safeTiles34[tile] = true
		}
		if player.reachTileAtGlobal != -1 {
			// 立直後振聽産生的安牌
			for _, tile := range normalDiscardTiles(d.globalDiscardTiles[player.reachTileAtGlobal:]) {
				riList[who].safeTiles34[tile] = true
			}
		} else if player.latestDiscardAtGlobal != -1 {
			// 同巡振聽産生的安牌
			// 即該玩家在最近一次捨牌後，其他玩家的捨牌
			for _, tile := range normalDiscardTiles(d.globalDiscardTiles[player.latestDiscardAtGlobal:]) {
				riList[who].safeTiles34[tile] = true
			}
		}

		// 特殊：槓産生的安牌
		// 很難想象一個人會在有 678888 的時候去開槓（即使有這個可能，本程序也是不防的）
		for _, meld := range player.melds {
			if meld.IsKan() {
				riList[who].safeTiles34[meld.Tiles[0]] = true
			}
		}
	}

	// 計算各種數據
	for who, player := range d.players {
		if who == 0 {
			// TODO: 暫時不計算自家的
			continue
		}

		// 該玩家的巡目 = 為其切過的牌的數目
		turns := util.MinInt(len(player.discardTiles), util.MaxTurns)
		if turns == 0 {
			turns = 1
		}

		// TODO: 若某人一直摸切，然後突然手切了一張字牌，那他很有可能默聽/一向聽
		if player.isReached {
			riList[who].tenpaiRate = 100.0
			if player.reachTileAtGlobal < len(d.globalDiscardTiles) { // 天鳳可能有數據漏掉
				riList[who].isTsumogiriRiichi = d.globalDiscardTiles[player.reachTileAtGlobal] < 0
			}
		} else {
			rate := util.CalcTenpaiRate(player.melds, player.discardTiles, player.meldDiscardsAt)
			if d.playerNumber == 3 {
				rate = util.GetTenpaiRate3(rate)
			}
			riList[who].tenpaiRate = rate
		}

		// 估計該玩家榮和點數
		var ronPoint float64
		switch {
		case player.canIppatsu:
			// 立直一發巡的榮和點數
			ronPoint = util.RonPointRiichiIppatsu
		case player.isReached:
			// 立直非一發巡的榮和點數
			ronPoint = util.RonPointRiichiHiIppatsu
		case player.isNaki:
			// 副露時的榮和點數（非常粗略地估計）
			doraCount := player.doraNum(d.doraList())
			ronPoint = util.RonPointOtherNakiWithDora(doraCount)
		default:
			// 默聽時的榮和點數
			ronPoint = util.RonPointDama
		}
		// 親家*1.5
		if who == d.dealer {
			ronPoint *= 1.5
		}
		riList[who]._ronPoint = ronPoint

		// 根據該玩家的巡目、現物、立直後通過的牌、NC、Dora、早外、榮和點數來計算每張牌的危險度
		risk34 := util.CalculateRiskTiles34(turns, riList[who].safeTiles34, d.leftCounts, d.doraList(), d.roundWindTile, player.selfWindTile).
			FixWithEarlyOutside(player.earlyOutsideTiles).
			FixWithPoint(ronPoint)
		riList[who].riskTable = riskTable(risk34)

		// 計算剩餘筋牌
		if len(player.melds) < 4 {
			riList[who].leftNoSujiTiles = util.CalculateLeftNoSujiTiles(riList[who].safeTiles34, d.leftCounts)
		} else {
			// 大吊車：愚型聽牌
		}
	}

	return riList
}

// TODO: 特殊處理w立直
func (d *roundData) isPlayerDaburii(who int) bool {
	// w立直成立的前提是沒有任何玩家副露
	for _, p := range d.players {
		if len(p.melds) > 0 {
			return false
		}
		// 對於三麻來說，還不能有拔北
		if p.nukiDoraNum > 0 {
			return false
		}
	}
	return d.players[who].reachTileAt == 0
}

// 自家的 PlayerInfo
func (d *roundData) newModelPlayerInfo() *model.PlayerInfo {
	const wannpaiTilesCount = 14
	leftDrawTilesCount := util.CountOfTiles34(d.leftCounts) - (wannpaiTilesCount - len(d.doraIndicators))
	for _, player := range d.players[1:] {
		leftDrawTilesCount -= 13 - 3*len(player.melds)
	}
	if d.playerNumber == 3 {
		leftDrawTilesCount += 13
	}

	melds := []model.Meld{}
	for _, m := range d.players[0].melds {
		melds = append(melds, *m)
	}

	const self = 0
	selfPlayer := d.players[self]

	return &model.PlayerInfo{
		HandTiles34: d.counts,
		Melds:       melds,
		DoraTiles:   d.doraList(),
		NumRedFives: d.numRedFives,

		RoundWindTile: d.roundWindTile,
		SelfWindTile:  selfPlayer.selfWindTile,
		IsParent:      d.dealer == self,
		//IsDaburii:     d.isPlayerDaburii(self), // FIXME PLS，應該在立直時就判斷
		IsRiichi: selfPlayer.isReached,

		DiscardTiles: normalDiscardTiles(selfPlayer.discardTiles),
		LeftTiles34:  d.leftCounts,

		LeftDrawTilesCount: leftDrawTilesCount,

		NukiDoraNum: selfPlayer.nukiDoraNum,
	}
}

func (d *roundData) analysis() error {
	if !debugMode {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("內部錯誤：", err)
			}
		}()
	}

	if debugMode {
		if msg := d.parser.GetMessage(); len(msg) > 0 {
			const printLimit = 500
			if len(msg) > printLimit {
				msg = msg[:printLimit]
			}
			fmt.Println("收到", msg)
		}
	}

	// 先獲取用戶信息
	if d.parser.IsLogin() {
		d.parser.HandleLogin()
	}

	if d.parser.SkipMessage() {
		return nil
	}

	// 若自家立直，則進入看戲模式
	// TODO: 見逃判斷
	if !d.parser.IsInit() && !d.parser.IsRoundWin() && !d.parser.IsRyuukyoku() && d.players[0].isReached {
		return nil
	}

	if debugMode {
		fmt.Println("當前座位為", d.parser.GetSelfSeat())
	}

	var currentRoundCache *roundAnalysisCache
	if analysisCache := getAnalysisCache(d.parser.GetSelfSeat()); analysisCache != nil {
		currentRoundCache = analysisCache.wholeGameCache[d.roundNumber][d.benNumber]
	}

	switch {
	case d.parser.IsInit():
		// round 開始/重連
		if !debugMode && !d.skipOutput {
			clearConsole()
		}

		roundNumber, benNumber, dealer, doraIndicators, hands, numRedFives := d.parser.ParseInit()
		switch d.parser.GetDataSourceType() {
		case dataSourceTypeTenhou:
			d.reset(roundNumber, benNumber, dealer)
			d.gameMode = gameModeMatch // TODO: 牌譜模式？
		case dataSourceTypeMajsoul:
			if dealer != -1 { // 先就坐，還沒洗牌呢~
				// 設置第一局的 dealer
				d.reset(0, 0, dealer)
				d.gameMode = gameModeMatch
				fmt.Printf("遊戲即將開始，您分配到的座位是：")
				color.HiGreen(util.MahjongZH[d.players[0].selfWindTile])
				return nil
			} else {
				// 根據 selfSeat 和當前的 roundNumber 計算當前局的 dealer
				newDealer := (4 - d.parser.GetSelfSeat() + roundNumber) % 4
				// 新的一局
				d.reset(roundNumber, benNumber, newDealer)
			}
		default:
			panic("not impl!")
		}

		// 由於 reset 了，重新獲取 currentRoundCache
		if analysisCache := getAnalysisCache(d.parser.GetSelfSeat()); analysisCache != nil {
			currentRoundCache = analysisCache.wholeGameCache[d.roundNumber][d.benNumber]
		}

		d.doraIndicators = doraIndicators
		for _, dora := range doraIndicators {
			d.descLeftCounts(dora)
		}
		for _, tile := range hands {
			d.counts[tile]++
			d.descLeftCounts(tile)
		}
		d.numRedFives = numRedFives

		playerInfo := d.newModelPlayerInfo()

		// 牌譜分析模式下，記錄捨牌推薦
		if d.gameMode == gameModeRecordCache && len(hands) == 14 {
			currentRoundCache.addAIDiscardTileWhenDrawTile(simpleBestDiscardTile(playerInfo), -1, 0, 0)
		}

		if d.skipOutput {
			return nil
		}

		// 牌譜模式下，打印捨牌推薦
		if d.gameMode == gameModeRecord {
			currentRoundCache.print()
		}

		color.New(color.FgHiGreen).Printf("%s", util.MahjongZH[d.roundWindTile])
		fmt.Printf("%d局開始，自風為", roundNumber%4+1)
		color.New(color.FgHiGreen).Printf("%s", util.MahjongZH[d.players[0].selfWindTile])
		fmt.Println()
		info := fmt.Sprintln(util.TilesToMahjongZHInterface(d.doraIndicators)...)
		info = info[:len(info)-1]
		color.HiYellow("寶牌指示牌是 " + info)
		fmt.Println()
		// TODO: 顯示地和概率
		return analysisPlayerWithRisk(playerInfo, nil)
	case d.parser.IsOpen():
		// 某家鳴牌（含暗槓、加槓）
		who, meld, kanDoraIndicator := d.parser.ParseOpen()
		meldType := meld.MeldType
		meldTiles := meld.Tiles
		calledTile := meld.CalledTile

		// 任何形式的鳴牌都能破除一發
		for _, player := range d.players {
			player.canIppatsu = false
		}

		// 槓寶牌指示牌
		if kanDoraIndicator != -1 {
			d.newDora(kanDoraIndicator)
		}

		player := d.players[who]

		// 不是暗槓則標記該玩家鳴牌了
		if meldType != meldTypeAnkan {
			player.isNaki = true
		}

		// 加槓單獨處理
		if meldType == meldTypeKakan {
			if who != 0 {
				// （不是自家時）修改牌山剩餘量
				d.descLeftCounts(calledTile)
			} else {
				// 自家加槓成功，修改手牌
				d.counts[calledTile]--
				// 由於均為自家操作，寶牌數是不變的

				// 牌譜分析模式下，記錄加槓操作
				if d.gameMode == gameModeRecordCache {
					currentRoundCache.addKan(meldType)
				}
			}
			// 修改原副露
			for _, _meld := range player.melds {
				// 找到原有的碰副露
				if _meld.Tiles[0] == calledTile {
					_meld.MeldType = meldTypeKakan
					_meld.Tiles = append(_meld.Tiles, calledTile)
					_meld.ContainRedFive = meld.ContainRedFive
					break
				}
			}

			if debugMode {
				if who == 0 {
					if handsCount := util.CountOfTiles34(d.counts); handsCount%3 != 1 {
						return fmt.Errorf("手牌錯誤：%d 張牌 %v", handsCount, d.counts)
					}
				}
			}

			break
		}

		// 修改玩家副露數據
		d.players[who].melds = append(d.players[who].melds, meld)

		if who != 0 {
			// （不是自家時）修改牌山剩餘量
			// 先增後減
			if meldType != meldTypeAnkan {
				d.leftCounts[calledTile]++
			}
			for _, tile := range meldTiles {
				d.descLeftCounts(tile)
			}
		} else {
			// 自家，修改手牌
			if meldType == meldTypeAnkan {
				d.counts[meldTiles[0]] = 0

				// 牌譜分析模式下，記錄暗槓操作
				if d.gameMode == gameModeRecordCache {
					currentRoundCache.addKan(meldType)
				}
			} else {
				d.counts[calledTile]++
				for _, tile := range meldTiles {
					d.counts[tile]--
				}
				if meld.RedFiveFromOthers {
					tileType := meldTiles[0] / 9
					d.numRedFives[tileType]++
				}

				// 牌譜分析模式下，記錄吃碰明槓操作
				if d.gameMode == gameModeRecordCache {
					currentRoundCache.addChiPonKan(meldType)
				}
			}

			if debugMode {
				if meldType == meldTypeMinkan || meldType == meldTypeAnkan {
					if handsCount := util.CountOfTiles34(d.counts); handsCount%3 != 1 {
						return fmt.Errorf("手牌錯誤：%d 張牌 %v", handsCount, d.counts)
					}
				} else {
					if handsCount := util.CountOfTiles34(d.counts); handsCount%3 != 2 {
						return fmt.Errorf("手牌錯誤：%d 張牌 %v", handsCount, d.counts)
					}
				}
			}
		}
	case d.parser.IsReach():
		// 立直宣告
		// 如果是他家立直，進入攻守判斷模式
		who := d.parser.ParseReach()
		d.players[who].isReached = true
		d.players[who].canIppatsu = true
		//case "AGARI", "RYUUKYOKU":
		//	// 某人和牌或流局，round 結束
		//case "PROF":
		//	// 遊戲結束
		//case "BYE":
		//	// 某人退出
		//case "REJOIN", "GO":
		//	// 重連
	case d.parser.IsFuriten():
		// 振聽
		if d.skipOutput {
			return nil
		}
		color.HiYellow("振聽")
		//case "U", "V", "W":
		//	//（下家,對家,上家 不要其上家的牌）摸牌
		//case "HELO", "RANKING", "TAIKYOKU", "UN", "LN", "SAIKAI":
		//	// 其他
	case d.parser.IsSelfDraw():
		if !debugMode && !d.skipOutput {
			clearConsole()
		}
		// 自家（從牌山 d.leftCounts）摸牌（至手牌 d.counts）
		tile, isRedFive, kanDoraIndicator := d.parser.ParseSelfDraw()
		d.descLeftCounts(tile)
		d.counts[tile]++
		if isRedFive {
			d.numRedFives[tile/9]++
		}
		if kanDoraIndicator != -1 {
			d.newDora(kanDoraIndicator)
		}

		playerInfo := d.newModelPlayerInfo()

		// 安全度分析
		riskTables := d.analysisTilesRisk()
		mixedRiskTable := riskTables.mixedRiskTable()

		// 牌譜分析模式下，記錄捨牌推薦
		if d.gameMode == gameModeRecordCache {
			bestAttackDiscardTile := simpleBestDiscardTile(playerInfo)
			bestDefenceDiscardTile := mixedRiskTable.getBestDefenceTile(playerInfo.HandTiles34)
			bestAttackDiscardTileRisk, bestDefenceDiscardTileRisk := 0.0, 0.0
			if bestDefenceDiscardTile >= 0 {
				bestAttackDiscardTileRisk = mixedRiskTable[bestAttackDiscardTile]
				bestDefenceDiscardTileRisk = mixedRiskTable[bestDefenceDiscardTile]
			}
			currentRoundCache.addAIDiscardTileWhenDrawTile(bestAttackDiscardTile, bestDefenceDiscardTile, bestAttackDiscardTileRisk, bestDefenceDiscardTileRisk)
		}

		if d.skipOutput {
			return nil
		}

		// 牌譜模式下，打印捨牌推薦
		if d.gameMode == gameModeRecord {
			currentRoundCache.print()
		}

		// 打印他家捨牌信息
		d.printDiscards()
		fmt.Println()

		// 打印手牌對各家的安全度
		riskTables.printWithHands(d.counts, d.leftCounts)

		// 打印何切推薦
		// TODO: 根據是否聽牌/一向聽、打點、巡目、和率等進行攻守判斷
		return analysisPlayerWithRisk(playerInfo, mixedRiskTable)
	case d.parser.IsDiscard():
		who, discardTile, isRedFive, isTsumogiri, isReach, canBeMeld, kanDoraIndicator := d.parser.ParseDiscard()

		if kanDoraIndicator != -1 {
			d.newDora(kanDoraIndicator)
		}

		player := d.players[who]
		if isReach {
			player.isReached = true
			player.canIppatsu = true
		}

		if who == 0 {
			// 特殊處理自家捨牌的情況
			riskTables := d.analysisTilesRisk()
			mixedRiskTable := riskTables.mixedRiskTable()

			// 自家（從手牌 d.counts）捨牌（至牌河 d.globalDiscardTiles）
			d.counts[discardTile]--

			d.globalDiscardTiles = append(d.globalDiscardTiles, discardTile)
			player.discardTiles = append(player.discardTiles, discardTile)
			player.latestDiscardAtGlobal = len(d.globalDiscardTiles) - 1

			if isRedFive {
				d.numRedFives[discardTile/9]--
			}

			// 牌譜分析模式下，記錄自家捨牌
			if d.gameMode == gameModeRecordCache {
				currentRoundCache.addSelfDiscardTile(discardTile, mixedRiskTable[discardTile], isReach)
			}

			if debugMode {
				if handsCount := util.CountOfTiles34(d.counts); handsCount%3 != 1 {
					return fmt.Errorf("手牌錯誤：%d 張牌 %v", handsCount, d.counts)
				}
			}

			return nil
		}

		// 他家捨牌
		d.descLeftCounts(discardTile)

		_disTile := discardTile
		if isTsumogiri {
			_disTile = ^_disTile
		}
		d.globalDiscardTiles = append(d.globalDiscardTiles, _disTile)
		player.discardTiles = append(player.discardTiles, _disTile)
		player.latestDiscardAtGlobal = len(d.globalDiscardTiles) - 1

		// 標記外側牌
		if !player.isReached && len(player.discardTiles) <= 5 {
			player.earlyOutsideTiles = append(player.earlyOutsideTiles, util.OutsideTiles(discardTile)...)
		}

		if player.isReached && player.reachTileAtGlobal == -1 {
			// 標記立直宣言牌
			player.reachTileAtGlobal = len(d.globalDiscardTiles) - 1
			player.reachTileAt = len(player.discardTiles) - 1
			// 若該玩家摸切立直，打印提示信息
			if isTsumogiri && !d.skipOutput {
				color.HiYellow("%s 摸切立直！", player.name)
			}
		} else if len(player.meldDiscardsAt) != len(player.melds) {
			// 標記鳴牌的捨牌
			// 注意這裏會標記到暗槓後的捨牌上
			// 注意對於連續開槓的情況，len(player.meldDiscardsAt) 和 len(player.melds) 是不等的
			player.meldDiscardsAt = append(player.meldDiscardsAt, len(player.discardTiles)-1)
			player.meldDiscardsAtGlobal = append(player.meldDiscardsAtGlobal, len(d.globalDiscardTiles)-1)
		}

		// 若玩家在立直後摸牌捨牌，則沒有一發
		if player.reachTileAt < len(player.discardTiles)-1 {
			player.canIppatsu = false
		}

		playerInfo := d.newModelPlayerInfo()

		// 安全度分析
		riskTables := d.analysisTilesRisk()
		mixedRiskTable := riskTables.mixedRiskTable()

		// 牌譜分析模式下，記錄可能的鳴牌
		if d.gameMode == gameModeRecordCache {
			allowChi := who == 3
			_, results14, incShantenResults14 := util.CalculateMeld(playerInfo, discardTile, isRedFive, allowChi)
			bestAttackDiscardTile := -1
			if len(results14) > 0 {
				bestAttackDiscardTile = results14[0].DiscardTile
			} else if len(incShantenResults14) > 0 {
				bestAttackDiscardTile = incShantenResults14[0].DiscardTile
			}
			if bestAttackDiscardTile != -1 {
				bestDefenceDiscardTile := mixedRiskTable.getBestDefenceTile(playerInfo.HandTiles34)
				bestAttackDiscardTileRisk := 0.0
				if bestDefenceDiscardTile >= 0 {
					bestAttackDiscardTileRisk = mixedRiskTable[bestAttackDiscardTile]
				}
				currentRoundCache.addPossibleChiPonKan(bestAttackDiscardTile, bestAttackDiscardTileRisk)
			}
		}

		if d.skipOutput {
			return nil
		}

		// 上家捨牌時若無法鳴牌則跳過顯示
		//if d.gameMode == gameModeMatch && who == 3 && !canBeMeld {
		//	return nil
		//}

		if !debugMode {
			clearConsole()
		}

		// 牌譜模式下，打印捨牌推薦
		if d.gameMode == gameModeRecord {
			currentRoundCache.print()
		}

		// 打印他家捨牌信息
		d.printDiscards()
		fmt.Println()
		riskTables.printWithHands(d.counts, d.leftCounts)

		if d.gameMode == gameModeMatch && !canBeMeld {
			return nil
		}

		// 為了方便解析牌譜，這裏盡可能地解析副露
		// TODO: 提醒: 消除海底/避免河底
		allowChi := d.playerNumber != 3 && who == 3 && playerInfo.LeftDrawTilesCount > 0
		return analysisMeld(playerInfo, discardTile, isRedFive, allowChi, mixedRiskTable)
	case d.parser.IsRoundWin():
		// TODO: 解析天鳳牌譜 - 注意 skipOutput

		if !debugMode {
			clearConsole()
		}
		fmt.Println("和牌，本局結束")
		whos, points := d.parser.ParseRoundWin()
		if len(whos) == 3 {
			color.HiYellow("鳳 凰 級 避 銃")
			if d.parser.GetDataSourceType() == dataSourceTypeMajsoul {
				color.HiYellow("（快醒醒，這是雀魂）")
			}
		}
		for i, who := range whos {
			fmt.Println(d.players[who].name, points[i])
		}
	case d.parser.IsRyuukyoku():
		// TODO
		d.parser.ParseRyuukyoku()
	case d.parser.IsNukiDora():
		who, isTsumogiri := d.parser.ParseNukiDora()
		player := d.players[who]
		player.nukiDoraNum++
		if who != 0 {
			// 減少北的數量
			d.descLeftCounts(30)
			// TODO
			_ = isTsumogiri
		} else {
			// 減少自己手牌中北的數量
			d.counts[30]--
		}
		// 消除一發
		for _, player := range d.players {
			player.canIppatsu = false
		}
	case d.parser.IsNewDora():
		// 槓寶牌
		// 1. 剩餘牌減少
		// 2. 打點提高
		kanDoraIndicator := d.parser.ParseNewDora()
		d.newDora(kanDoraIndicator)
	default:
	}

	return nil
}
