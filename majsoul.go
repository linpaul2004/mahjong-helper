package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"sort"
	"time"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/proto/lq"
)

type majsoulMessage struct {
	// 對應到服務器用戶數據庫中的ID，該值越小表示您的註冊時間越早
	AccountID int `json:"account_id"`

	// 友人列表
	Friends lq.FriendList `json:"friends"`

	// 新獲取到的牌譜基本信息列表
	RecordBaseInfoList []*majsoulRecordBaseInfo `json:"record_list"`

	// 分享的牌譜基本信息
	SharedRecordBaseInfo *majsoulRecordBaseInfo `json:"shared_record_base_info"`

	// 當前正在觀看的牌譜的 UUID
	CurrentRecordUUID string `json:"current_record_uuid"`

	// 當前正在觀看的牌譜的全部操作
	RecordActions []*majsoulRecordAction `json:"record_actions"`

	// 玩家在網頁上的（點擊）操作（網頁響應了的）
	RecordClickAction      string `json:"record_click_action"`
	RecordClickActionIndex int    `json:"record_click_action_index"`
	FastRecordTo           int    `json:"fast_record_to"` // 閉區間

	// 觀戰
	LiveBaseInfo   *majsoulLiveRecordBaseInfo `json:"live_head"`
	LiveFastAction *majsoulRecordAction       `json:"live_fast_action"`
	LiveAction     *majsoulRecordAction       `json:"live_action"`

	// 座位變更
	ChangeSeatTo *int `json:"change_seat_to"`

	// 遊戲重連時收到的數據
	SyncGameActions []*majsoulRecordAction `json:"sync_game_actions"`

	// ResAuthGame
	// {"seat_list":[x,x,x,x],"is_game_start":false,"game_config":{"category":1,"mode":{"mode":1,"ai":true,"detail_rule":{"time_fixed":60,"time_add":0,"dora_count":3,"shiduan":1,"init_point":25000,"fandian":30000,"bianjietishi":true,"ai_level":1,"fanfu":1}},"meta":{"room_id":18269}},"ready_id_list":[0,0,0]}
	IsGameStart *bool              `json:"is_game_start"` // false=新遊戲，true=重連
	SeatList    []int              `json:"seat_list"`
	ReadyIDList []int              `json:"ready_id_list"`
	GameConfig  *majsoulGameConfig `json:"game_config"`

	// NotifyPlayerLoadGameReady
	//ReadyIDList []int `json:"ready_id_list"`

	// ActionNewRound
	// {"chang":0,"ju":0,"ben":0,"tiles":["1m","3m","7m","3p","6p","7p","6s","1z","1z","2z","3z","4z","7z"],"dora":"6m","scores":[25000,25000,25000,25000],"liqibang":0,"al":false,"md5":"","left_tile_count":69}
	MD5   string      `json:"md5"`
	Chang *int        `json:"chang"`
	Ju    *int        `json:"ju"`
	Ben   *int        `json:"ben"`
	Tiles interface{} `json:"tiles"` // 一般情況下為 []interface{}, interface{} 即 string，但是暗槓的情況下，該值為一個 string
	Dora  string      `json:"dora"`

	// RecordNewRound
	Tiles0 []string `json:"tiles0"`
	Tiles1 []string `json:"tiles1"`
	Tiles2 []string `json:"tiles2"`
	Tiles3 []string `json:"tiles3"`

	// ActionDealTile
	// {"seat":1,"tile":"5m","left_tile_count":23,"operation":{"seat":1,"operation_list":[{"type":1}],"time_add":0,"time_fixed":60000},"zhenting":false}
	// 他家暗槓後的摸牌
	// {"seat":1,"left_tile_count":3,"doras":["7m","0p"],"zhenting":false}
	Seat          *int     `json:"seat"`
	Tile          string   `json:"tile"`
	Doras         []string `json:"doras"` // 暗槓摸牌了，同時翻出槓寶牌指示牌
	LeftTileCount *int     `json:"left_tile_count"`

	// ActionDiscardTile
	// {"seat":0,"tile":"5z","is_liqi":false,"moqie":true,"zhenting":false,"is_wliqi":false}
	// {"seat":0,"tile":"1z","is_liqi":false,"operation":{"seat":1,"operation_list":[{"type":3,"combination":["1z|1z"]}],"time_add":0,"time_fixed":60000},"moqie":false,"zhenting":false,"is_wliqi":false}
	// 吃 碰 和
	// {"seat":0,"tile":"6p","is_liqi":false,"operation":{"seat":1,"operation_list":[{"type":2,"combination":["7p|8p"]},{"type":3,"combination":["6p|6p"]},{"type":9}],"time_add":0,"time_fixed":60000},"moqie":false,"zhenting":true,"is_wliqi":false}
	IsLiqi    *bool     `json:"is_liqi"`
	IsWliqi   *bool     `json:"is_wliqi"`
	Moqie     *bool     `json:"moqie"`
	Operation *struct{} `json:"operation"`

	// ActionChiPengGang || ActionAnGangAddGang
	// 他家吃 {"seat":0,"type":0,"tiles":["2s","3s","4s"],"froms":[0,0,3],"zhenting":false}
	// 他家碰 {"seat":1,"type":1,"tiles":["1z","1z","1z"],"froms":[1,1,0],"operation":{"seat":1,"operation_list":[{"type":1,"combination":["1z"]}],"time_add":0,"time_fixed":60000},"zhenting":false,"tingpais":[{"tile":"4m","zhenting":false,"infos":[{"tile":"6s","haveyi":true},{"tile":"6p","haveyi":true}]},{"tile":"7m","zhenting":false,"infos":[{"tile":"6s","haveyi":true},{"tile":"6p","haveyi":true}]}]}
	// 他家大明槓 {"seat":2,"type":2,"tiles":["3z","3z","3z","3z"],"froms":[2,2,2,0],"zhenting":false}
	// 他家加槓 {"seat":2,"type":2,"tiles":"3z"}
	// 他家暗槓 {"seat":2,"type":3,"tiles":"3s"}
	Type  int   `json:"type"`
	Froms []int `json:"froms"`

	// ActionLiqi

	// ActionHule
	Hules []struct {
		Seat          int  `json:"seat"`
		Zimo          bool `json:"zimo"`
		PointRong     int  `json:"point_rong"`
		PointZimoQin  int  `json:"point_zimo_qin"`
		PointZimoXian int  `json:"point_zimo_xian"`
	} `json:"hules"`

	// ActionLiuJu
	// {"liujumanguan":false,"players":[{"tingpai":true,"hand":["3s","3s","4s","5s","6s","1z","1z","7z","7z","7z"],"tings":[{"tile":"1z","haveyi":true},{"tile":"3s","haveyi":true}]},{"tingpai":false},{"tingpai":false},{"tingpai":true,"hand":["4m","0m","6m","6m","6m","4s","4s","4s","5s","7s"],"tings":[{"tile":"6s","haveyi":true}]}],"scores":[{"old_scores":[23000,29000,24000,24000],"delta_scores":[1500,-1500,-1500,1500]}],"gameend":false}
	//Liujumanguan *bool `json:"liujumanguan"`
	//Players *struct{ } `json:"players"`
	//Gameend      *bool `json:"gameend"`

	// ActionBabei
}

const (
	majsoulMeldTypeChi = iota
	majsoulMeldTypePon
	majsoulMeldTypeMinkanOrKakan
	majsoulMeldTypeAnkan
)

type majsoulRoundData struct {
	*roundData

	originJSON string
	msg        *majsoulMessage

	selfSeat int // 自家初始座位：0-第一局的東家 1-第一局的南家 2-第一局的西家 3-第一局的北家
}

func (d *majsoulRoundData) fatalParse(info string, msg string) {
	panic(fmt.Sprintln(info, len(msg), msg, []byte(msg)))
}

func (d *majsoulRoundData) normalTiles(tiles interface{}) (majsoulTiles []string) {
	_tiles, ok := tiles.([]interface{})
	if !ok {
		_tile, ok := tiles.(string)
		if !ok {
			panic(fmt.Sprintln("[normalTiles] 解析錯誤", tiles))
		}
		return []string{_tile}
	}

	majsoulTiles = make([]string, len(_tiles))
	for i, _tile := range _tiles {
		_t, ok := _tile.(string)
		if !ok {
			panic(fmt.Sprintln("[normalTiles] 解析錯誤", tiles))
		}
		majsoulTiles[i] = _t
	}
	return majsoulTiles
}

func (d *majsoulRoundData) parseWho(seat int) int {
	// 轉換成 0=自家, 1=下家, 2=對家, 3=上家
	// 對三麻四麻均適用
	who := (seat + d.dealer - d.roundNumber%4 + 4) % 4
	return who
}

func (d *majsoulRoundData) mustParseMajsoulTile(humanTile string) (tile34 int, isRedFive bool) {
	tile34, isRedFive, err := util.StrToTile34(humanTile)
	if err != nil {
		panic(err)
	}
	return
}

func (d *majsoulRoundData) mustParseMajsoulTiles(majsoulTiles []string) (tiles []int, numRedFive int) {
	tiles = make([]int, len(majsoulTiles))
	for i, majsoulTile := range majsoulTiles {
		var isRedFive bool
		tiles[i], isRedFive = d.mustParseMajsoulTile(majsoulTile)
		if isRedFive {
			numRedFive++
		}
	}
	return
}

func (d *majsoulRoundData) isNewDora(doras []string) bool {
	return len(doras) > len(d.doraIndicators)
}

func (d *majsoulRoundData) GetDataSourceType() int {
	return dataSourceTypeMajsoul
}

func (d *majsoulRoundData) GetSelfSeat() int {
	return d.selfSeat
}

func (d *majsoulRoundData) GetMessage() string {
	return d.originJSON
}

func (d *majsoulRoundData) SkipMessage() bool {
	msg := d.msg

	// 沒有帳號 skip
	if gameConf.currentActiveMajsoulAccountID == -1 {
		return true
	}

	// TODO: 重構
	if msg.SeatList != nil {
		// 特判古役模式
		isGuyiMode := msg.GameConfig.isGuyiMode()
		util.SetConsiderOldYaku(isGuyiMode)
		if isGuyiMode {
			color.HiGreen("古役模式已開啟")
			time.Sleep(2 * time.Second)
		}
	} else {
		// msg.SeatList 必須為 nil
		if msg.ReadyIDList != nil {
			// 打印準備信息
			fmt.Printf("等待玩家準備 (%d/%d) %v\n", len(msg.ReadyIDList), d.playerNumber, msg.ReadyIDList)
		}
	}

	return false
}

func (d *majsoulRoundData) IsLogin() bool {
	msg := d.msg
	return msg.AccountID > 0 || msg.SeatList != nil
}

func (d *majsoulRoundData) HandleLogin() {
	msg := d.msg

	if accountID := msg.AccountID; accountID > 0 {
		gameConf.addMajsoulAccountID(accountID)
		if accountID != gameConf.currentActiveMajsoulAccountID {
			printAccountInfo(accountID)
			gameConf.setMajsoulAccountID(accountID)
		}
		return
	}

	// 從對戰 ID 列表中獲取帳號 ID
	if seatList := msg.SeatList; seatList != nil {
		// 嘗試從中找到緩存帳號 ID
		for _, accountID := range seatList {
			if accountID > 0 && gameConf.isIDExist(accountID) {
				// 找到了，更新當前使用的帳號 ID
				if gameConf.currentActiveMajsoulAccountID != accountID {
					printAccountInfo(accountID)
					gameConf.setMajsoulAccountID(accountID)
				}
				return
			}
		}

		// 未找到緩存 ID
		if gameConf.currentActiveMajsoulAccountID > 0 {
			color.HiRed("尚未獲取到您的帳號 ID，請您刷新網頁，或開啟一局人機對戰（錯誤信息：您的帳號 ID %d 不在對戰列表 %v 中）", gameConf.currentActiveMajsoulAccountID, msg.SeatList)
			return
		}

		// 判斷是否為人機對戰，若為人機對戰，則獲取帳號 ID
		if !util.InInts(0, msg.SeatList) {
			return
		}
		for _, accountID := range msg.SeatList {
			if accountID > 0 {
				gameConf.addMajsoulAccountID(accountID)
				printAccountInfo(accountID)
				gameConf.setMajsoulAccountID(accountID)
				return
			}
		}
	}
}

func (d *majsoulRoundData) IsInit() bool {
	msg := d.msg
	// ResAuthGame || ActionNewRound RecordNewRound
	return msg.IsGameStart != nil || msg.MD5 != ""
}

func (d *majsoulRoundData) ParseInit() (roundNumber int, benNumber int, dealer int, doraIndicators []int, handTiles []int, numRedFives []int) {
	msg := d.msg

	if playerNumber := len(msg.SeatList); playerNumber >= 3 {
		d.playerNumber = playerNumber
		// 獲取自家初始座位：0-第一局的東家 1-第一局的南家 2-第一局的西家 3-第一局的北家
		for i, accountID := range msg.SeatList {
			if accountID == gameConf.currentActiveMajsoulAccountID {
				d.selfSeat = i
				break
			}
		}
		// dealer: 0=自家, 1=下家, 2=對家, 3=上家
		dealer = (4 - d.selfSeat) % 4
		return
	} else if len(msg.Tiles2) > 0 {
		if len(msg.Tiles3) > 0 {
			d.playerNumber = 4
		} else {
			d.playerNumber = 3
		}
	}
	dealer = -1

	roundNumber = 4*(*msg.Chang) + *msg.Ju
	benNumber = *msg.Ben
	if msg.Dora != "" {
		doraIndicator, _ := d.mustParseMajsoulTile(msg.Dora)
		doraIndicators = append(doraIndicators, doraIndicator)
	} else {
		for _, dora := range msg.Doras {
			doraIndicator, _ := d.mustParseMajsoulTile(dora)
			doraIndicators = append(doraIndicators, doraIndicator)
		}
	}
	numRedFives = make([]int, 3)

	var majsoulTiles []string
	if msg.Tiles != nil { // 實戰
		majsoulTiles = d.normalTiles(msg.Tiles)
	} else { // 牌譜、觀戰
		majsoulTiles = [][]string{msg.Tiles0, msg.Tiles1, msg.Tiles2, msg.Tiles3}[d.selfSeat]
	}
	for _, majsoulTile := range majsoulTiles {
		tile, isRedFive := d.mustParseMajsoulTile(majsoulTile)
		handTiles = append(handTiles, tile)
		if isRedFive {
			numRedFives[tile/9]++
		}
	}

	return
}

func (d *majsoulRoundData) IsSelfDraw() bool {
	msg := d.msg
	// ActionDealTile RecordDealTile
	return msg.Seat != nil && msg.Tile != "" && msg.Moqie == nil && d.parseWho(*msg.Seat) == 0
}

func (d *majsoulRoundData) ParseSelfDraw() (tile int, isRedFive bool, kanDoraIndicator int) {
	msg := d.msg
	tile, isRedFive = d.mustParseMajsoulTile(msg.Tile)
	kanDoraIndicator = -1
	if d.isNewDora(msg.Doras) {
		kanDoraIndicator, _ = d.mustParseMajsoulTile(msg.Doras[len(msg.Doras)-1])
	}
	return
}

func (d *majsoulRoundData) IsDiscard() bool {
	msg := d.msg
	// ActionDiscardTile RecordDiscardTile
	return msg.IsLiqi != nil
}

func (d *majsoulRoundData) ParseDiscard() (who int, discardTile int, isRedFive bool, isTsumogiri bool, isReach bool, canBeMeld bool, kanDoraIndicator int) {
	msg := d.msg
	who = d.parseWho(*msg.Seat)
	discardTile, isRedFive = d.mustParseMajsoulTile(msg.Tile)
	isTsumogiri = *msg.Moqie
	isReach = *msg.IsLiqi
	if msg.IsWliqi != nil && !isReach { // 兼容雀魂早期牌譜（無 IsWliqi 字段）
		isReach = *msg.IsWliqi
	}
	canBeMeld = msg.Operation != nil // 注意：觀戰模式下無此選項
	kanDoraIndicator = -1
	if d.isNewDora(msg.Doras) {
		kanDoraIndicator, _ = d.mustParseMajsoulTile(msg.Doras[len(msg.Doras)-1])
	}
	return
}

func (d *majsoulRoundData) IsOpen() bool {
	msg := d.msg
	// ActionChiPengGang RecordChiPengGang || ActionAnGangAddGang RecordAnGangAddGang
	return msg.Tiles != nil && len(d.normalTiles(msg.Tiles)) <= 4
}

func (d *majsoulRoundData) ParseOpen() (who int, meld *model.Meld, kanDoraIndicator int) {
	msg := d.msg

	who = d.parseWho(*msg.Seat)

	kanDoraIndicator = -1
	if d.isNewDora(msg.Doras) { // 暗槓（有時會在玩家摸牌後才發送 doras，可能是因為需要考慮搶暗槓的情況）
		kanDoraIndicator, _ = d.mustParseMajsoulTile(msg.Doras[len(msg.Doras)-1])
	}

	var meldType, calledTile int

	majsoulTiles := d.normalTiles(msg.Tiles)
	isSelfKan := len(majsoulTiles) == 1 // 自家加槓或暗槓
	if isSelfKan {
		majsoulTile := majsoulTiles[0]
		majsoulTiles = []string{majsoulTile, majsoulTile, majsoulTile, majsoulTile}
	}
	meldTiles, numRedFive := d.mustParseMajsoulTiles(majsoulTiles)
	containRedFive := numRedFive > 0
	if len(majsoulTiles) == 4 && meldTiles[0] < 27 && meldTiles[0]%9 == 4 {
		// 槓5意味著一定有赤5
		containRedFive = true
	}

	if isSelfKan {
		calledTile = meldTiles[0]
		// 用 msg.Type 判斷是加槓還是暗槓
		// 也可以通過是否有相關碰副露來判斷是加槓還是暗槓
		if msg.Type == majsoulMeldTypeMinkanOrKakan {
			meldType = meldTypeKakan // 加槓
		} else if msg.Type == majsoulMeldTypeAnkan {
			meldType = meldTypeAnkan // 暗槓
		}
		meld = &model.Meld{
			MeldType:       meldType,
			Tiles:          meldTiles,
			CalledTile:     calledTile,
			ContainRedFive: containRedFive,
		}
		return
	}

	var rawCalledTile string
	for i, seat := range msg.Froms {
		fromWho := d.parseWho(seat)
		if fromWho != who {
			rawCalledTile = majsoulTiles[i]
		}
	}
	if rawCalledTile == "" {
		panic("數據解析異常: 未找到 rawCalledTile")
	}
	calledTile, redFiveFromOthers := d.mustParseMajsoulTile(rawCalledTile)

	if len(meldTiles) == 3 {
		if meldTiles[0] == meldTiles[1] {
			meldType = meldTypePon // 碰
		} else {
			meldType = meldTypeChi // 吃
			sort.Ints(meldTiles)
		}
	} else if len(meldTiles) == 4 {
		meldType = meldTypeMinkan // 大明槓
	} else {
		panic("鳴牌數據解析失敗！")
	}
	meld = &model.Meld{
		MeldType:          meldType,
		Tiles:             meldTiles,
		CalledTile:        calledTile,
		ContainRedFive:    containRedFive,
		RedFiveFromOthers: redFiveFromOthers,
	}
	return
}

func (d *majsoulRoundData) IsReach() bool {
	return false
}

func (d *majsoulRoundData) ParseReach() (who int) {
	return 0
}

func (d *majsoulRoundData) IsFuriten() bool {
	return false
}

func (d *majsoulRoundData) IsRoundWin() bool {
	msg := d.msg
	// ActionHule RecordHule
	return msg.Hules != nil
}

func (d *majsoulRoundData) ParseRoundWin() (whos []int, points []int) {
	msg := d.msg

	for _, result := range msg.Hules {
		who := d.parseWho(result.Seat)
		whos = append(whos, d.parseWho(result.Seat))
		point := result.PointRong
		if result.Zimo {
			if who == d.dealer {
				point = 3 * result.PointZimoXian
			} else {
				point = result.PointZimoQin + 2*result.PointZimoXian
			}
			if d.playerNumber == 3 {
				// 自摸損（一個子家）
				point -= result.PointZimoXian
			}
		}
		points = append(points, point)
	}
	return
}

func (d *majsoulRoundData) IsRyuukyoku() bool {
	// TODO
	// ActionLiuJu RecordLiuJu
	return false
}

func (d *majsoulRoundData) ParseRyuukyoku() (type_ int, whos []int, points []int) {
	// TODO
	return
}

// 拔北寶牌
func (d *majsoulRoundData) IsNukiDora() bool {
	msg := d.msg
	// ActionBaBei RecordBaBei
	return msg.Seat != nil && msg.Moqie != nil && msg.Tile == ""
}

func (d *majsoulRoundData) ParseNukiDora() (who int, isTsumogiri bool) {
	msg := d.msg
	return d.parseWho(*msg.Seat), *msg.Moqie
}

// 在最後處理該項
func (d *majsoulRoundData) IsNewDora() bool {
	msg := d.msg
	// ActionDealTile
	return d.isNewDora(msg.Doras)
}

func (d *majsoulRoundData) ParseNewDora() (kanDoraIndicator int) {
	msg := d.msg

	kanDoraIndicator, _ = d.mustParseMajsoulTile(msg.Doras[len(msg.Doras)-1])
	return
}
