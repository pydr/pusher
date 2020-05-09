package pusher

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var UpGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
