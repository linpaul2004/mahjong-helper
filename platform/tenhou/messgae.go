package tenhou

type message struct {
	Tag string `json:"tag" xml:"-"`

	//Name string `json:"name"` // id
	//Sex  string `json:"sx"`

	UserName string `json:"uname" xml:"-"`
	//RatingScale string `json:"ratingscale"`

	//N string `json:"n"`
	//J string `json:"j"`
	//G string `json:"g"`

	// round 開始 tag=INIT
	Seed   string `json:"seed" xml:"seed,attr"` // 本局信息：場數，場棒數，立直棒數，骰子A減一，骰子B減一，寶牌指示牌 1,0,0,3,2,92
	Ten    string `json:"ten" xml:"ten,attr"`   // 各家點數 280,230,240,250
	Dealer string `json:"oya" xml:"oya,attr"`   // 莊家 0=自家, 1=下家, 2=對家, 3=上家
	Hai    string `json:"hai" xml:"hai,attr"`   // 初始手牌 30,114,108,31,78,107,25,23,2,14,122,44,49
	Hai0   string `json:"-" xml:"hai0,attr"`
	Hai1   string `json:"-" xml:"hai1,attr"`
	Hai2   string `json:"-" xml:"hai2,attr"`
	Hai3   string `json:"-" xml:"hai3,attr"`

	// 摸牌 tag=T編號，如 T68

	// 副露 tag=N
	Who  string `json:"who" xml:"who,attr"` // 副露者 0=自家, 1=下家, 2=對家, 3=上家
	Meld string `json:"m" xml:"m,attr"`     // 副露編號 35914

	// 槓寶牌指示牌 tag=DORA
	// `json:"hai"` // 槓寶牌指示牌 39

	// 立直聲明 tag=REACH, step=1
	// `json:"who"` // 立直者
	Step string `json:"step" xml:"step,attr"` // 1

	// 立直成功，扣1000點 tag=REACH, step=2
	// `json:"who"` // 立直者
	// `json:"ten"` // 立直成功後的各家點數 250,250,240,250
	// `json:"step"` // 2

	// 自摸/有人放銃 tag=牌, t>=8
	T string `json:"t"` // 選項

	// 和牌 tag=AGARI
	// ba, hai, m, machi, ten, yaku, doraHai, who, fromWho, sc
	//Ba string `json:"ba"` // 0,0
	// `json:"hai"` // 和牌型 8,9,11,14,19,125,126,127
	// `json:"m"` // 副露編號 13527,50794
	//Machi string `json:"machi"` // (待ち) 自摸/榮和的牌 126
	// `json:"ten"` // 符數,點數,這張牌的來源 30,7700,0
	//Yaku        string `json:"yaku"`       // 役（編號，翻數） 18,1,20,1,34,2
	//DoraTile    string `json:"doraHai"`    // 寶牌 123
	//UraDoraTile string `json:"doraHaiUra"` // 裏寶牌 77
	// `json:"who"` // 和牌者
	//FromWho string `json:"fromWho"` // 自摸/榮和牌的來源
	//Score   string `json:"sc"`      // 各家增減分 260,-77,310,77,220,0,210,0

	// 遊戲結束 tag=PROF

	// 重連 tag=GO
	// type, lobby, gpid
	//Type  string `json:"type"`
	//Lobby string `json:"lobby"`
	//GPID  string `json:"gpid"`

	// 重連 tag=REINIT
	// `json:"seed"`
	// `json:"ten"`
	// `json:"oya"`
	// `json:"hai"`
	//Meld1    string `json:"m1"` // 各家副露編號 17450
	//Meld2    string `json:"m2"`
	//Meld3    string `json:"m3"`
	//Kawa0 string `json:"kawa0"` // 各家牌河 112,73,3,131,43,98,78,116
	//Kawa1 string `json:"kawa1"`
	//Kawa2 string `json:"kawa2"`
	//Kawa3 string `json:"kawa3"`
}
