package pusher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func subTestHandle(msg *ReceiveMsg) error {
	fmt.Printf("[Server] got a sub message: %v\n", msg)
	return nil
}

func unsubHandle(msg *ReceiveMsg) error {
	fmt.Printf("[Server] got a unsub message: %v\n", msg)
	return nil
}

func ws(writer http.ResponseWriter, request *http.Request) {
	ws, err := UpGrader.Upgrade(writer, request, nil)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	NewClient(ws).Run(subTestHandle, unsubHandle)
}

func TestMain(m *testing.M) {
	http.HandleFunc("/ws", ws)
	go func() {
		err := http.ListenAndServe(":9090", nil)
		if err == nil {
			fmt.Println("websocket server starting...")
		}
	}()

	time.Sleep(2 * time.Second)
	m.Run()
}

func TestWebsocket(t *testing.T) {
	t.Log("connecting...")
	ws, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:9090/ws", nil)
	if err != nil {
		t.Fatal("dial failed: ", err)
	}
	t.Log("connected")
	defer ws.Close()

	// 订阅
	subMsg := &ReceiveMsg{
		Action:    Sub,
		Topic:     "market.btc_usdt.depth",
		Timestamp: time.Now().Unix(),
	}
	jsonData, _ := json.Marshal(subMsg)
	err = ws.WriteMessage(websocket.BinaryMessage, jsonData)
	if err != nil {
		t.Fatal("send msg failed: ", err)
	}

	unsubMsg := &ReceiveMsg{
		Action:    Unsub,
		Topic:     "market.btc_usdt.depth",
		Timestamp: time.Now().Unix(),
	}
	jsonData, _ = json.Marshal(unsubMsg)
	err = ws.WriteMessage(websocket.BinaryMessage, jsonData)
	if err != nil {
		t.Fatal("send msg failed: ", err)
	}

	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				t.Fatal("read msg failed: ", err)
			}
			fmt.Println("[Client] receive a message: ", string(message))

			var data SendMsg
			err = json.Unmarshal(message, &data)
			if err != nil {
				t.Fatal("read msg failed: ", err)
			}

			// 心跳响应
			if data.Action.Action == Ping {
				ac := &Action{
					Action:  Pong,
					Message: "success",
					Status:  true,
				}

				data := new(SendMsg)
				data.Action = ac
				data.Timestamp = time.Now().Unix()
				jsonData, _ := json.Marshal(data)

				err = ws.WriteMessage(websocket.BinaryMessage, jsonData)
				if err != nil {
					t.Fatal("send msg failed: ", err)
				}
			}
		}
	}()

	select {}
}
