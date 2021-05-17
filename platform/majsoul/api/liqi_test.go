package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/proto/lq"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/tool"
	"github.com/satori/go.uuid"
	"os"
	"testing"
	"time"
)

func _genReqLogin(t *testing.T) *lq.ReqLogin {
	username, ok := os.LookupEnv("USERNAME")
	if !ok {
		t.Skip("未配置環境變量 USERNAME，退出")
	}

	password, ok := os.LookupEnv("PASSWORD")
	if !ok {
		t.Skip("未配置環境變量 PASSWORD，退出")
	}
	const key = "lailai" // from code.js
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(password))
	password = fmt.Sprintf("%x", mac.Sum(nil))

	// randomKey 最好是個固定值
	randomKey, ok := os.LookupEnv("RANDOM_KEY")
	if !ok {
		rawRandomKey, _ := uuid.NewV4()
		randomKey = rawRandomKey.String()
	}

	version, err := tool.GetMajsoulVersion(tool.ApiGetVersionZH)
	if err != nil {
		t.Fatal(err)
	}
	return &lq.ReqLogin{
		Account:   username,
		Password:  password,
		Reconnect: false,
		Device: &lq.ClientDeviceInfo{
			DeviceType: "pc",
			Os:         "",
			OsVersion:  "",
			Browser:    "safari",
		},
		RandomKey:         randomKey,          // 例如 aa566cfc-547e-4cc0-a36f-2ebe6269109b
		ClientVersion:     version.ResVersion, // 0.5.162.w
		GenAccessToken:    true,
		CurrencyPlatforms: []uint32{2}, // 1-inGooglePlay, 2-inChina
	}
}

func _genReqOauth2Login(t *testing.T, accessToken string) *lq.ReqOauth2Login {
	randomKey, ok := os.LookupEnv("RANDOM_KEY")
	if !ok {
		rawRandomKey, _ := uuid.NewV4()
		randomKey = rawRandomKey.String()
	}

	version, err := tool.GetMajsoulVersion(tool.ApiGetVersionZH)
	if err != nil {
		t.Fatal(err)
	}
	return &lq.ReqOauth2Login{
		Type:        0, // ? 懷疑是帳號/QQ/微信/微博
		AccessToken: accessToken,
		Reconnect:   false,
		Device: &lq.ClientDeviceInfo{
			DeviceType: "pc",
			Os:         "",
			OsVersion:  "",
			Browser:    "safari",
		},
		RandomKey:         randomKey,
		ClientVersion:     version.ResVersion,
		CurrencyPlatforms: []uint32{2}, // 1-inGooglePlay, 2-inChina
	}
}

func TestLogin(t *testing.T) {
	endpoint, err := tool.GetMajsoulWebSocketURL() // wss://mj-srv-7.majsoul.com:4131/
	if err != nil {
		t.Fatal(err)
	}
	t.Log("連接 endpoint: " + endpoint)
	c := NewWebSocketClient()
	if err := c.Connect(endpoint, tool.MajsoulOriginURL); err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	reqLogin := _genReqLogin(t)
	respLogin, err := c.Login(reqLogin)
	if err != nil {
		t.Skip("登錄失敗:", err)
	}
	t.Log("登錄成功:", respLogin)
	t.Log(respLogin.AccessToken)

	time.Sleep(time.Second)

	respLogout, err := c.Logout(&lq.ReqLogout{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("登出", respLogout)
}

func TestReLogin(t *testing.T) {
	endpoint, err := tool.GetMajsoulWebSocketURL() // wss://mj-srv-7.majsoul.com:4131/
	if err != nil {
		t.Fatal(err)
	}
	t.Log("連接 endpoint: " + endpoint)
	c := NewWebSocketClient()
	if err := c.Connect(endpoint, tool.MajsoulOriginURL); err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	accessToken, ok := os.LookupEnv("TOKEN")
	if !ok {
		t.Skip("未配置環境變量 TOKEN，退出")
	}
	reqOauth2Check := lq.ReqOauth2Check{
		// Type = 3 為 QQ
		Type:        0, // 帳號/QQ/微信/微博/ 海外的……?
		AccessToken: accessToken,
	}
	respOauth2Check, err := c.Oauth2Check(&reqOauth2Check)
	if err != nil {
		t.Skip("token 驗證失敗:", err)
	}
	t.Log(respOauth2Check)

	if !respOauth2Check.HasAccount {
		t.Skip("無效的 token")
	}

	reqOauth2Login := _genReqOauth2Login(t, accessToken)
	respLogin, err := c.Oauth2Login(reqOauth2Login)
	if err != nil {
		t.Skip("登錄失敗:", err)
	}
	t.Log("登錄成功:", respLogin)
	t.Log(respLogin.AccessToken)

	time.Sleep(time.Second)

	respLogout, err := c.Logout(&lq.ReqLogout{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("登出", respLogout)
}
