package session

import (
	"github.com/gorilla/websocket"
	"webSocket/user"
)

type Session struct {
	User          *user.User
	WebsocketConn *websocket.Conn
}
