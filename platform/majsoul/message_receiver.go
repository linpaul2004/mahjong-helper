package majsoul

import (
	"github.com/golang/protobuf/proto"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/api"
	"fmt"
	"os"
	"reflect"
	"encoding/binary"
	"strings"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/proto/lq"
)

// 若 NotifyMessage 不為空，這該消息為通知，RequestMessage 和 ResponseMessage 字段為空
// 否則該消息為請求響應，NotifyMessage 字段為空
type Message struct {
	Name            string        `json:"name"`
	RequestMessage  proto.Message `json:"request_message,omitempty"`
	ResponseMessage proto.Message `json:"response_message,omitempty"`
	NotifyMessage   proto.Message `json:"notify_message,omitempty"`
}

type MessageReceiver struct {
	originMessageQueue  chan []byte   // 包含所有 WebSocket 發出的消息和收到的消息
	orderedMessageQueue chan *Message // 整理後的 WebSocket 收到的消息（包含請求響應和通知）

	indexToMessageMap map[uint16]*Message
}

func NewMessageReceiver() *MessageReceiver {
	const maxQueueSize = 100
	mr := &MessageReceiver{
		originMessageQueue:  make(chan []byte, maxQueueSize),
		orderedMessageQueue: make(chan *Message, maxQueueSize),
		indexToMessageMap:   map[uint16]*Message{},
	}
	go mr.run()
	return mr
}

func (mr *MessageReceiver) run() {
	for data := range mr.originMessageQueue {
		messageType := data[0]
		switch messageType {
		case api.MessageTypeNotify:
			notifyName, data, err := api.UnwrapData(data[1:])
			if err != nil {
				fmt.Fprintln(os.Stderr, "MessageReceiver.run.api.UnwrapData.NOTIFY", err)
				continue
			}
			notifyName = notifyName[1:] // 移除開頭的 .

			mt := proto.MessageType(notifyName)
			if mt == nil {
				fmt.Fprintf(os.Stderr, "MessageReceiver.run 未找到 %s，請檢查！\n", notifyName)
				continue
			}
			messagePtr := reflect.New(mt.Elem())
			if err := proto.Unmarshal(data, messagePtr.Interface().(proto.Message)); err != nil {
				fmt.Fprintln(os.Stderr, "MessageReceiver.run.proto.Unmarshal.NOTIFY", notifyName, err)
				continue
			}

			mr.orderedMessageQueue <- &Message{
				Name:          notifyName,
				NotifyMessage: messagePtr.Interface().(proto.Message),
			}
		case api.MessageTypeRequest:
			messageIndex := binary.LittleEndian.Uint16(data[1:3])

			rawMethodName, data, err := api.UnwrapData(data[3:])
			if err != nil {
				fmt.Fprintln(os.Stderr, "MessageReceiver.run.api.UnwrapData.REQUEST", err)
				continue
			}
			rawMethodName = rawMethodName[1:] // 移除開頭的 .

			// 通過 rawMethodName 找到請求類型和請求響應類型
			splits := strings.Split(rawMethodName, ".")
			clientName, methodName := splits[1], splits[2]
			methodType := lq.FindMethod(clientName, methodName)
			reqType := methodType.In(1)
			respType := methodType.Out(0)

			messagePtr := reflect.New(reqType.Elem())
			if err := proto.Unmarshal(data, messagePtr.Interface().(proto.Message)); err != nil {
				fmt.Fprintln(os.Stderr, "MessageReceiver.run.proto.Unmarshal.REQUEST", rawMethodName, err)
				continue
			}
			reqMessage := messagePtr.Interface().(proto.Message)

			messagePtr = reflect.New(respType.Elem())
			respMessage := messagePtr.Interface().(proto.Message)

			mr.indexToMessageMap[messageIndex] = &Message{
				Name:            rawMethodName,
				RequestMessage:  reqMessage,
				ResponseMessage: respMessage,
			}
		case api.MessageTypeResponse:
			// 似乎是有序返回的……
			messageIndex := binary.LittleEndian.Uint16(data[1:3])
			message, ok := mr.indexToMessageMap[messageIndex]
			if !ok {
				// 用戶在啟動助手前就啟動了雀魂
				continue
			}
			delete(mr.indexToMessageMap, messageIndex)
			if err := api.UnwrapMessage(data[3:], message.ResponseMessage); err != nil {
				fmt.Fprintln(os.Stderr, "MessageReceiver.run.proto.Unmarshal.RESPONSE", message.Name, err)
				continue
			}
			mr.orderedMessageQueue <- message
		default:
			panic(fmt.Sprintln("[MessageReceiver] 數據有誤", messageType))
		}
	}
}

func (mr *MessageReceiver) Put(data []byte) {
	mr.originMessageQueue <- data
}

func (mr *MessageReceiver) Get() *Message {
	return <-mr.orderedMessageQueue
}
