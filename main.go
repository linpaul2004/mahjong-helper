package main

import (
	"flag"
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"github.com/fatih/color"
	"math/rand"
	"strings"
	"time"
)

var (
	considerOldYaku bool

	isMajsoul     bool
	isTenhou      bool
	isAnalysis    bool
	isInteractive bool

	showImproveDetail      bool
	showAgariAboveShanten1 bool
	showScore              bool
	showAllYakuTypes       bool

	humanDoraTiles string

	port int
)

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.BoolVar(&considerOldYaku, "old", false, "允許古役")
	flag.BoolVar(&isMajsoul, "majsoul", false, "雀魂助手")
	flag.BoolVar(&isTenhou, "tenhou", false, "天鳳助手")
	flag.BoolVar(&isAnalysis, "analysis", false, "分析模式")
	flag.BoolVar(&isInteractive, "interactive", false, "交互模式")
	flag.BoolVar(&isInteractive, "i", false, "同 -interactive")
	flag.BoolVar(&showImproveDetail, "detail", false, "顯示改良細節")
	flag.BoolVar(&showAgariAboveShanten1, "agari", false, "顯示聽牌前的估計和率")
	flag.BoolVar(&showAgariAboveShanten1, "a", false, "同 -agari")
	flag.BoolVar(&showScore, "score", false, "顯示局收支")
	flag.BoolVar(&showScore, "s", false, "同 -score")
	flag.BoolVar(&showAllYakuTypes, "yaku", false, "顯示所有役種")
	flag.BoolVar(&showAllYakuTypes, "y", false, "同 -yaku")
	flag.StringVar(&humanDoraTiles, "dora", "", "指定哪些牌是寶牌")
	flag.StringVar(&humanDoraTiles, "d", "", "同 -dora")
	flag.IntVar(&port, "port", 12121, "指定服務端口")
	flag.IntVar(&port, "p", 12121, "同 -port")
}

const (
	platformTenhou  = 0
	platformMajsoul = 1

	defaultPlatform = platformMajsoul
)

var platforms = map[int][]string{
	platformTenhou: {
		"天鳳",
		"Web",
		"4K",
	},
	platformMajsoul: {
		"雀魂",
		"國際中文服",
		"日服",
		"國際服",
	},
}

const readmeURL = "https://github.com/EndlessCheng/mahjong-helper/blob/master/README.md"
const issueURL = "https://github.com/EndlessCheng/mahjong-helper/issues"
const issueCommonQuestions = "https://github.com/EndlessCheng/mahjong-helper/issues/104"
const qqGroupNum = "375865038"

func welcome() int {
	fmt.Println("使用說明：" + readmeURL)
	fmt.Println("問題反饋：" + issueURL)
	fmt.Println("吐槽群：" + qqGroupNum)
	fmt.Println()

	fmt.Println("請輸入數字，選擇對應網站：")
	for i, cnt := 0, 0; cnt < len(platforms); i++ {
		if platformInfo, ok := platforms[i]; ok {
			info := platformInfo[0] + " [" + strings.Join(platformInfo[1:], ",") + "]"
			fmt.Printf("%d - %s\n", i, info)
			cnt++
		}
	}

	choose := defaultPlatform
	fmt.Scanln(&choose) // 直接回車也無妨
	platformInfo, ok := platforms[choose]
	var platformName string
	if ok {
		platformName = platformInfo[0]
	}
	if !ok {
		choose = defaultPlatform
		platformName = platforms[choose][0]
	}

	clearConsole()
	color.HiGreen("已選擇 - %s", platformName)

	if choose == platformMajsoul {
		if len(gameConf.MajsoulAccountIDs) == 0 {
			color.HiYellow(`
提醒：首次啟用時，請開啟一局人機對戰，或者重登遊戲。
該步驟用於獲取您的帳號 ID，便於在遊戲開始時獲取自風，否則程序將無法解析後續數據。

若助手無響應，請確認您已按步驟安裝完成。
相關鏈接 ` + issueCommonQuestions)
		}
	}

	return choose
}

func main() {
	flag.Parse()

	color.HiGreen("日本麻將助手 %s (by EndlessCheng)", version)
	if version != versionDev {
		go checkNewVersion(version)
	}

	util.SetConsiderOldYaku(considerOldYaku)

	humanTiles := strings.Join(flag.Args(), " ")
	humanTilesInfo := &model.HumanTilesInfo{
		HumanTiles:     humanTiles,
		HumanDoraTiles: humanDoraTiles,
	}

	var err error
	switch {
	case isMajsoul:
		err = runServer(true, port)
	case isTenhou || isAnalysis:
		err = runServer(true, port)
	case isInteractive: // 交互模式
		err = interact(humanTilesInfo)
	case len(flag.Args()) > 0: // 靜態分析
		_, err = analysisHumanTiles(humanTilesInfo)
	default: // 服務器模式
		choose := welcome()
		isHTTPS := choose == platformMajsoul
		err = runServer(isHTTPS, port)
	}
	if err != nil {
		errorExit(err)
	}
}
