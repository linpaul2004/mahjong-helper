package model

const (
	MeldTypeChi    = iota // 吃
	MeldTypePon           // 碰
	MeldTypeAnkan         // 暗槓
	MeldTypeMinkan        // 大明槓
	MeldTypeKakan         // 加槓
)

type Meld struct {
	MeldType int // 鳴牌類型（吃、碰、暗槓、大明槓、加槓）

	// Tiles == sort(SelfTiles + CalledTile)
	Tiles      []int // 副露的牌
	SelfTiles  []int // 手牌中組成副露的牌（用於鳴牌分析）
	CalledTile int   // 被鳴的牌

	// TODO: 重構 ContainRedFive RedFiveFromOthers
	ContainRedFive    bool // 是否包含赤5
	RedFiveFromOthers bool // 赤5是否來自他家（用於獲取寶牌數）
}

// 是否為槓子
func (m *Meld) IsKan() bool {
	return m.MeldType == MeldTypeAnkan || m.MeldType == MeldTypeMinkan || m.MeldType == MeldTypeKakan
}
