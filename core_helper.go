package main

var debugMode = false

type gameMode int

const (
	// TODO: 感覺有點雜亂需要重構
	gameModeMatch       gameMode = iota // 對戰 - IsInit
	gameModeRecord                      // 解析牌譜
	gameModeRecordCache                 // 解析牌譜 - runMajsoulRecordAnalysisTask
	gameModeLive                        // 解析觀戰
)

const (
	dataSourceTypeTenhou = iota
	dataSourceTypeMajsoul
)

const (
	meldTypeChi    = iota // 吃
	meldTypePon           // 碰
	meldTypeAnkan         // 暗槓
	meldTypeMinkan        // 大明槓
	meldTypeKakan         // 加槓
)

// 負數變正數
func normalDiscardTiles(discardTiles []int) []int {
	newD := make([]int, len(discardTiles))
	copy(newD, discardTiles)
	for i, discardTile := range newD {
		if discardTile < 0 {
			newD[i] = ^discardTile
		}
	}
	return newD
}
