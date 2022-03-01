package room

import "webSocket/user"

var _rooms map[int]Room

func init() {
	_rooms = make(map[int]Room, 0)
}

type Room struct {
	Users []*user.User
}
