package pusher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWebsocket(t *testing.T) {
	ws, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/api/v1/ws", nil)
	if err != nil {
		t.Fatal("dial failed: ", err)
	}
	defer ws.Close()

	go func() {
		data := new(ReceiveMsg)
		data.Action = Sub
		data.Topic = "market.btc_usdt.depth"
		data.Timestamp = time.Now().Unix()
		jsonData, _ := json.Marshal(data)

		err = ws.WriteMessage(websocket.BinaryMessage, jsonData)
		if err != nil {
			t.Fatal("send msg failed: ", err)
		}
	}()

	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				t.Fatal("read msg failed: ", err)
			}

			fmt.Println("receive msg--->", string(message))

			var data SendMsg
			err = json.Unmarshal(message, &data)
			if err != nil {
				t.Fatal("read msg failed: ", err)
			}

			if data.Action.Action == Ping {
				ac := new(Action)
				ac.Action = Pong
				ac.Status = true
				ac.Message = "success"
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

func subTestHandle(msg *ReceiveMsg) error {
	fmt.Print(msg)
	return nil
}

func unsubHandle(msg *ReceiveMsg) error {
	fmt.Print(msg)
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

func TestNewClient(t *testing.T) {
	http.HandleFunc("/ws", ws)
	http.ListenAndServe(":9090", nil)
}
