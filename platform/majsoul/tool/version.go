package tool

type majsoulVersion struct {
	Code       string `json:"code"`    // code.js 路徑 v0.5.81.w/code.js
	ResVersion string `json:"version"` // 資源文件版本  0.5.82.w（注意開頭沒有 v）
}

func GetMajsoulVersion(apiGetVersionURL string) (version *majsoulVersion, err error) {
	version = &majsoulVersion{}
	err = get(apiGetVersionURL, version)
	return
}
