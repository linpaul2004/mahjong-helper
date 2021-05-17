package majsoul

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/api"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/proto/lq"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/tool"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"reflect"
)

const (
	RecordTypeAll         uint32 = 0
	RecordTypeFriend      uint32 = 1
	RecordTypeLevel       uint32 = 2
	RecordTypeCompetition uint32 = 4
	// 收藏的牌譜用 FetchGameRecordsDetail 接口獲取
	// 該接口傳入的 UUID 在登錄後調用 FetchCollectedGameRecordList 獲得
)

func genReqLogin(username string, password string) (*lq.ReqLogin, error) {
	const key = "lailai" // from code.js
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(password))
	password = fmt.Sprintf("%x", mac.Sum(nil))

	// randomKey 最好是個固定值
	randomKey, ok := os.LookupEnv("RANDOM_KEY")
	if !ok {
		randomKey = uuid.NewV4().String()
	}

	version, err := tool.GetMajsoulVersion(tool.ApiGetVersionZH)
	if err != nil {
		return nil, err
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
	}, nil
}

// TODO: add token
func DownloadRecords(username string, password string, recordType uint32) error {
	c := api.NewWebSocketClient()
	if err := c.ConnectMajsoul(); err != nil {
		return err
	}
	defer c.Close()

	// 登錄
	reqLogin, err := genReqLogin(username, password)
	if err != nil {
		return err
	}
	if _, err := c.Login(reqLogin); err != nil {
		return err
	}
	defer c.Logout(&lq.ReqLogout{})

	// 分頁獲取牌譜列表
	// TODO: 若之前下載過，可以判斷：上次是否下載完成->只下載本地最新文件之後的牌譜
	recordList := []*lq.RecordGame{}
	const pageSize = 10
	for i := uint32(1); ; i += pageSize {
		reqGameRecordList := lq.ReqGameRecordList{
			Start: i,
			Count: pageSize,
			Type:  recordType,
		}
		respGameRecordList, err := c.FetchGameRecordList(&reqGameRecordList)
		if err != nil {
			return err
		}
		recordList = append(recordList, respGameRecordList.RecordList...)
		if len(respGameRecordList.RecordList) < pageSize {
			break
		}
	}

	// TODO: 若牌譜數量巨大，可以使用協程增加下載速度
	for i, gameRecord := range recordList {
		fmt.Printf("%d/%d %s\n", i+1, len(recordList), gameRecord.Uuid)

		// 獲取具體牌譜內容
		reqGameRecord := lq.ReqGameRecord{
			GameUuid: gameRecord.Uuid,
		}
		respGameRecord, err := c.FetchGameRecord(&reqGameRecord)
		if err != nil {
			return err
		}

		// 解析
		data := respGameRecord.Data
		if len(data) == 0 {
			dataURL := respGameRecord.DataUrl
			if dataURL == "" {
				fmt.Fprintln(os.Stderr, "數據異常: dataURL 為空")
				continue
			}
			data, err = tool.Fetch(dataURL)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
		}
		detailRecords := lq.GameDetailRecords{}
		if err := api.UnwrapMessage(data, &detailRecords); err != nil {
			return err
		}

		type messageWithType struct {
			Name string        `json:"name"`
			Data proto.Message `json:"data"`
		}
		details := []messageWithType{}
		for _, detailRecord := range detailRecords.GetRecords() {
			name, data, err := api.UnwrapData(detailRecord)
			if err != nil {
				return err
			}

			name = name[1:] // 移除開頭的 .
			mt := proto.MessageType(name)
			if mt == nil {
				return fmt.Errorf("未找到 %s，請檢查代碼！", name)
			}
			messagePtr := reflect.New(mt.Elem())
			if err := proto.Unmarshal(data, messagePtr.Interface().(proto.Message)); err != nil {
				return err
			}

			details = append(details, messageWithType{
				Name: name[3:], // 移除開頭的 lq.
				Data: messagePtr.Interface().(proto.Message),
			})
		}

		// 保存至本地（JSON 格式）
		parseResult := struct {
			Head    *lq.RecordGame    `json:"head"`
			Details []messageWithType `json:"details"`
		}{
			Head:    gameRecord,
			Details: details,
		}
		jsonData, err := json.MarshalIndent(&parseResult, "", "  ")
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(gameRecord.Uuid+".json", jsonData, 0644); err != nil {
			return err
		}
	}

	return nil
}
