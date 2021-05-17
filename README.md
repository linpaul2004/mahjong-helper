# 日本麻將助手

## 主要功能

- 對戰時自動分析手牌，綜合進張、打點等，給出推薦捨牌
- 有人立直或多副露時，顯示各張牌的危險度
- 記錄他家手切摸切
- 助手帶你看牌譜，推薦每一打的進攻和防守選擇
- 支持四人麻將和三人麻將

### 支持平臺

- 雀魂網頁版（[國際中文服](https://game.maj-soul.com/1/) | [日服](https://game.mahjongsoul.com) | [國際服](https://mahjongsoul.game.yo-star.com)）
- 天鳳（[Web](https://tenhou.net/3/) | [4K](https://tenhou.net/4/)）


## 導航

- [安裝](#安裝)
- [使用說明](#使用說明)
- [示例](#示例)
  * [牌效率](#牌效率)
  * [鳴牌判斷](#鳴牌判斷)
  * [手摸切與安牌顯示](#手摸切與安牌顯示)
- [牌譜與觀戰](#牌譜與觀戰)
- [其他功能說明](#其他功能說明)
- [參與討論](#參與討論)


## 安裝

分下面幾步：

1. 前往 [releases](https://github.com/EndlessCheng/mahjong-helper/releases/latest) 頁面下載助手。解壓到本地後打開。

2. 雀魂需要瀏覽器允許本地證書，在瀏覽器地址欄中輸入 `chrome://flags/#allow-insecure-localhost`，然後點擊高亮那一項的「啟用」按鈕（[若沒有該項見此](https://github.com/EndlessCheng/mahjong-helper/issues/108)）。該功能僅限基於 Chrome 內核開發的瀏覽器。
   
   （不同瀏覽器/版本的描述可能不一樣，如果打開的頁面是英文的話，高亮的就是 `Allow invalid certificates for resources loaded from localhost`，把它的 Disabled 改成 Enabled）
   
   設置完成後**重啟瀏覽器**。

3. 安裝瀏覽器擴展 Header Editor，用於修改 code.js 文件，發送雀魂遊戲中的玩家操作信息至本地運行的助手。
   
   若能翻牆請前往 [谷歌商城](https://chrome.google.com/webstore/detail/header-editor/eningockdidmgiojffjmkdblpjocbhgh?hl=zh) 下載該擴展。或者 [從 CRX 安裝該擴展](https://www.chromefor.com/header-editor_v4-0-7/)（若無法安裝試試 360 瀏覽器）。
   
   安裝好擴展後點進該擴展的`管理`界面，點擊`導入和導出`，在下載規則中填入 `https://endlesscheng.gitee.io/public/mahjong-helper.json`，點擊右側的下載按鈕，然後點擊下方的`保存`。

4. 如果您的瀏覽器之前打開過雀魂網頁，需要清除緩存：打開雀魂網頁，按下 F12，右鍵地址欄左側的刷新按鈕，選擇「清空緩存並進行硬刷新」。這一操作只需要首次使用時做一次。如遇問題，請參考 [#104](https://github.com/EndlessCheng/mahjong-helper/issues/104)。

#### 安裝完成！在使用助手前，請先閱讀本頁面下方的[示例](#示例)，以了解助手輸出信息的含義。

### 從源碼安裝此助手

您也可以選擇從源碼安裝此助手：

`go get -u -v github.com/EndlessCheng/mahjong-helper/...`

完成後程序生成於 `$GOPATH/bin/` 目錄下。


## 使用說明

按照上述流程安裝完成後，啟動助手，選擇平臺即可。

需要先啟動本助手，再打開網頁。

### 注意事項

終端有個小 bug，在使用中若鼠標點擊到了空白處，可能會導緻終端卡住，此時按下回車鍵就可以恢複正常。


## 示例

### 牌效率

每個切牌選擇會顯示進張數、向聽前進後的進張數、可以做的役種等等信息。

**助手會綜合每個切牌選擇的速度、打點、和率，速度越快，打點和率越高的越靠前。**

每個切牌選擇以如下格式顯示：

```
進張數[改良後的進張數加權均值] 切哪張牌 => 向聽前進後的進張數的加權均值 [手牌速度] [期望打點] [役種] [是否振聽] [進張牌]
```

例如：

![](img/example01.png)


補充說明：

- 無改良時，不顯示改良進張數
- 鳴牌時會顯示用手上的哪些牌去吃/碰，詳見後文
- 防守時，切牌的文字顔色會因這張牌的安全程度而不同，詳見後文
- 門清聽牌時，會顯示立直的期望點數（考慮自摸、一發和裏寶）；若默聽有役則會額外顯示默聽的榮和點數
- 存在高低目的場合會顯示加權和率的平均點數
- 役種只對較為特殊的進行提示，如三色、一通、七對等。雀魂亂鬥之間會有額外的古役提醒
- 若鳴牌且無役會提示 `[無役]`
- 聽牌或一向聽時根據自家捨牌情況提示振聽
- m-萬子 p-餅子 s-索子 z-字牌，順序為東南西北白發中

進張數顔色說明：

- 紅色：優秀
- 黃色：普通
- 藍色：較差

來看看下面這幾道何切題吧。

**1\. 完全一向聽**

![](img/example-speed01c.png)

標準的完全一向聽形狀，切 8s。

![](img/example-speed1.png)

**2\. 三個複合形的一向聽**（選自《麻雀 傑作「何切る」300選》Q106）

![](img/example-Q106.png)

這種情況要切哪一張牌呢？

單看進張，切 7s 是進張最廣的，但是從更長遠的角度來看，切 7s 後會有愚型聽牌的可能。

一般來說，犧牲一點進張去換取更高的好型聽牌率，或者更高的打點是可以接受的。

如下圖所示，這裏展示了本助手對進張數、好型聽牌和打點的綜合判斷。相比 7s，切 4m 雖然進張數少了四枚，但是能好型聽牌，綜合和牌率比 7s 要高，同時還有平和的可能，可以說在保證了速度的同時又兼顧了打點，是最平衡的一打。所以切 4m 這個選項排在第一位。

![](img/example-speed2.png)

**3\. 向聽倒退**

![](img/example03b.png)

這裏巡目尚早，相比切 8m，切 1m 雖然向聽倒退但是進張面廣且有斷幺一役，速度是高於 8m 的。

如下圖，助手額外給出了向聽倒退的建議。（根據進張的不同，可能會形成七對，也可能會形成平和等）

![](img/example03a.png)

### 鳴牌判斷

下圖是一個鳴了紅中之後，聽坎 5s 的例子，寶牌為 6m。

上家打出 6m 寶牌之後考慮是否鳴牌：

這裏就可以考慮用 57m 吃，打出 9m，提升打點的同時又能維持聽牌。此外，若巡目尚早可以拆掉 46s 追求混一色。

![](img/example_naki2.png)

### 手摸切與安牌顯示

下圖展示了某局中三家的手摸切情況（寶牌為紅中和 6s，自家手牌此時為 345678m 569p 45667s）：

- 白色為手切，暗灰色為摸切
- 鳴牌後打出的那張牌會用灰底白字顯示，供讀牌分析用
- 副露玩家的手切中張牌(3-7)會有不同顔色的高亮，用來輔助判斷其聽牌率
- 玩家立直或聽牌率較高時會額外顯示對該玩家的安牌，用 | 分隔，左側為現物，右側按照危險度由低到高排序（No Chance 和 One Chance 的安牌作為補充參考顯示在後面，簡寫為 NC 和 OC）
- 下圖上家親家暗槓 2m 後 4p 立直，對家 8s 追立，下家一副露但是手切了很多中張牌，聽牌率較高
- 多家立直/副露時會顯示綜合危險度
- 危險度綜合考慮了巡目、副露數、他家打點估計（包含親家與否、副露中的寶牌數等）
- `[n無筋]` 指該玩家的無筋危險牌的剩餘種類數。剩餘種類數越少，這些無筋牌就越危險。剩餘種類數為零錶示該玩家是愚型聽牌或振聽（註：把 1p4p 這種算作一對筋牌，對於四人麻將來說一共有 3\*6=18 對筋牌，三人麻將則為 2\*6=12 對筋牌）

![](img/example5c1.png)

危險度顔色：

- 白色：現物
- 藍色：<5% 
- 黃色：5~10%
- 紅色：10~15%
- 深紅：>15%

補充說明：

- 危險度排序是基於巡目、筋牌、No Chance、早外、寶牌、聽牌率等數據的綜合考慮結果，對於 One Chance 和其他特殊情況並沒有考慮，請玩家自行斟酌
- 某些情況下的 No Chance 安牌，本助手是會將其視作現物的（比如 3m 為壁，剩下的 2m 在牌河和自己手裏時，2m 是不會放銃的）


## 牌譜與觀戰

目前助手支持解析雀魂的牌譜（含分享）和觀戰下的手牌，切換視角也可以解析其他玩家的手牌。


## 其他功能說明

分析何切題時對一副手牌進行分析，可以輸入如下命令（mahjong-helper 指的是程序名稱，可以修改成自定義的名稱）：

- 分析何切
    
    東南西北白發中分別用 1-7z 錶示，紅 5 萬用 0m 錶示，紅 5 餅用 0p 錶示，紅 5 索用 0s 錶示
    
    `mahjong-helper 34068m 5678p 23567s`
    
    在 `#` 後面添加副露的牌，暗槓用大寫錶示
    
    `mahjong-helper 234688m 34s # 6666P 234p`
    
- 分析鳴牌
    
    在 `+` 後面添加要鳴的牌，支持用 0 錶示的紅寶牌
    
    `mahjong-helper 33567789m 46s + 6m`
    
    `mahjong-helper 33567789m 46s + 0s`
    
    `mahjong-helper 24688m 34s # 6666P 234p + 3m`

- 用交互模式分析手牌
    
    `mahjong-helper -i 34568m 5678p 23567s`
    
    輸入的切牌、摸牌用簡寫形式，如 `6m`
    
    [配套小工具](https://github.com/EndlessCheng/mahjong-helper-gui)

- 指出寶牌是哪些（-d 參數，不能有空格）
    
    比如下面的寶牌是 3p 8p 3m 3m
    
    `mahjong-helper -d=38p33m 34568m 5678p 23567s`

- 額外顯示打點估計（-s 參數，支持一向聽和兩向聽）
    
    `mahjong-helper -d=38p33m -s 34568m 5678p 23567s`
    
    特別說明，也可以直接用 `mahjong-helper -s` 啟動助手，可以顯示更多的信息（適合高分辨率的屏幕）

- 幫助信息（-h 參數）

    `mahjong-helper -h`


## 如何獲取 WebSocket 收發的消息

1. 打開開發者工具，找到相關 JS 文件，保存到本地。
2. 搜索 `WebSocket`, `socket`，找到 `message`, `onmessage`, `send` 等函數。
3. 修改代碼，使用 `XMLHttpRequest` 將收發的消息發送到（在 localhost 開啟的）mahjong-helper 服務器，服務器收到消息後會自動進行相關分析。（這一步也可以用油猴腳本來完成）
4. 上傳 JS 代碼到一個可以公網訪問的地方，最簡單的方法是傳至 GitHub Pages，即個人的 github.io 項目。拿到該 JS 文件地址。
5. 安裝瀏覽器擴展 Header Editor，重定向原 JS 文件地址到上一步中拿到的地址。
6. 允許本地證書通過瀏覽器，在瀏覽器（僅限 Chrome 內核）中輸入
    
    ```
    chrome://flags/#allow-insecure-localhost
    ```
    
    然後把高亮那一項的 Disabled 改成 Enabled（不同瀏覽器/版本的描述可能不一樣，如果是中文的話點擊「啟用」按鈕）。

7. 重啟瀏覽器。

下面說明天鳳和雀魂的代碼註入點。

### 天鳳 (tenhou)

1. 搜索 `WebSocket`，找到下方 `message` 對應的函數，該函數中的 `a.data` 就是 WebSocket 收到的 JSON 數據。
2. 在該函數開頭（或末尾）添加如下代碼：

    ```javascript
    var req = new XMLHttpRequest();
    req.open("POST", "http://localhost:12121/");
    req.send(a.data);
    ```

### 雀魂 (majsoul)

雀魂收發的消息是 protobuf，接收的消息一部分為含有類型的通知消息，另一部分為不含有類型的請求響應消息，
對於後者需要獲取雀魂發送的消息以獲得響應消息類型。

也就是說需要將雀魂發送和接收的消息都發給 mahjong-helper 服務器。

類似天鳳，搜索 `WebSocket` 找到下方的 `_socket.onmessage` 和 `_socket.send`，添加代碼。

服務器收到消息後，可以基於 [liqi.json](https://github.com/EndlessCheng/mahjong-helper/blob/master/platform/majsoul/proto/lq/liqi.json) 文件解析雀魂的 protobuf 數據。

[record.go](https://github.com/EndlessCheng/mahjong-helper/blob/master/platform/majsoul/record.go) 展示了使用 WebSocket 登錄和下載牌譜的例子。

考慮到還有觀看牌譜這種獲取前端 UI 事件的情況，還需修改額外的代碼。在網頁控制臺輸入 `GameMgr.inRelease = 0`，開啟調試模式，通過雀魂已有的日誌可以看到相關代碼在哪。具體修改了哪些內容可以對比雀魂的 code.js 和我修改後的 [code-zh.js](https://endlesscheng.gitee.io/public/js/majsoul/code-zh.js)。


## 參與討論

吐槽本項目、日麻技術、麻將算法交流，歡迎加入 QQ 群 [375865038](https://jq.qq.com/?_wv=1027&k=5FyZOgH)


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
