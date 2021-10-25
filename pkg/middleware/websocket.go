package middleware

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var WsUpgrader = websocket.Upgrader{
	// allow origin request
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
