package tenhou

import (
	"time"
	"encoding/json"
	"regexp"
)

type MessageReceiver struct {
	originMessageQueue  chan []byte
	orderedMessageQueue chan []byte
}

func NewMessageReceiver() *MessageReceiver {
	const maxQueueSize = 100
	mr := &MessageReceiver{
		originMessageQueue:  make(chan []byte, maxQueueSize),
		orderedMessageQueue: make(chan []byte, maxQueueSize),
	}
	go mr.run()
	return mr
}

var isSelfDraw = regexp.MustCompile("^T[0-9]{1,3}$").MatchString

// TODO: 後續使用 parser 中提供的方法
func (mr *MessageReceiver) isSelfDraw(data []byte) bool {
	d := struct {
		Tag string `json:"tag"`
	}{}
	if err := json.Unmarshal(data, &d); err != nil {
		return false
	}
	return isSelfDraw(d.Tag)
}

func (mr *MessageReceiver) run() {
	for data := range mr.originMessageQueue {
		if !mr.isSelfDraw(data) {
			mr.orderedMessageQueue <- data
			continue
		}

		// 收到了自家摸牌的消息，則等待一段很短的時間
		time.Sleep(75 * time.Millisecond) // 實際間隔在 3~9ms

		// 未收到新數據
		if len(mr.originMessageQueue) == 0 {
			mr.orderedMessageQueue <- data
			continue
		}

		// 在短時間內收到了新數據
		// 因為摸牌後肯定要等待玩家操作，正常情況是不會馬上有新數據的，所以這說明前端亂序發來了數據
		// 把 data 重新塞回去，這樣才是正確的順序
		mr.originMessageQueue <- data
	}
}

func (mr *MessageReceiver) Put(data []byte) {
	mr.originMessageQueue <- data
}

func (mr *MessageReceiver) Get() []byte {
	return <-mr.orderedMessageQueue
}

func (mr *MessageReceiver) IsEmpty() bool {
	return len(mr.originMessageQueue) == 0 && len(mr.orderedMessageQueue) == 0
}
