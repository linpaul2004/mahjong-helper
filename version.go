package main

import (
	"net/http"
	"time"
	"fmt"
	"encoding/json"
	"github.com/fatih/color"
)

const versionDev = "dev"

// 編譯時自動寫入版本號
// go build -ldflags "-X main.version=$(git describe --abbrev=0 --tags)" -o mahjong-helper
var version = versionDev

func fetchLatestVersionTag() (latestVersionTag string, err error) {
	const apiGetLatestRelease = "https://api.github.com/repos/EndlessCheng/mahjong-helper/releases/latest"
	const timeout = 10 * time.Second

	c := &http.Client{Timeout: timeout}
	resp, err := c.Get(apiGetLatestRelease)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("[fetchLatestVersionTag] 返回 %s", resp.Status)
	}

	d := struct {
		TagName string `json:"tag_name"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return
	}

	return d.TagName, nil
}

func checkNewVersion(currentVersionTag string) {
	const latestReleasePage = "https://github.com/EndlessCheng/mahjong-helper/releases/latest"

	latestVersionTag, err := fetchLatestVersionTag()
	if err != nil {
		// 下次再說~
		return
	}

	if latestVersionTag > currentVersionTag {
		color.HiGreen("檢測到新版本: %s！請前往 %s 下載", latestVersionTag, latestReleasePage)
	}
}
