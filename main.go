package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
	"webSocket/consts"
	"webSocket/session"
	"webSocket/user"
)

var _SessionPool map[*session.Session]struct{}

var addr = flag.String("addr", ":13111", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { // 不进行检查
		return true
	},
} // use default options

func main() {
	_SessionPool = make(map[*session.Session]struct{}, 0)
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/handleSession", handleSession)
	go CronWork()
	log.Printf("[INFO] system started, listen addr:%v", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func CronWork() {
	for {
		printSystemInfo()
		time.Sleep(60 * time.Second)
	}
}

func printSystemInfo() {
	boardPostMessage := []byte(fmt.Sprintf("欢迎来到ZZMilkTEA的聊天室，当前在线 %d 人", len(_SessionPool)))
	for sess, _ := range _SessionPool {
		// TODO 小心nil指针，先暂时不检查
		err := sess.WebsocketConn.WriteMessage(consts.STRING_MSG, boardPostMessage)
		if err != nil {
			log.Println("write:", err)
			continue
		}
	}
}

func handleSession(w http.ResponseWriter, r *http.Request) {
	// 创建会话
	s := &session.Session{}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// 处理会话断开
	defer func() {
		if err != nil {
			log.Println("write:", err)
			return
		}
		c.Close()
		delete(_SessionPool, s)
	}()

	vars := r.URL.Query()
	nickName, ok := vars["nick_name"]
	if ok {
		s.User = &user.User{NickName: nickName[0]}
	} else {
		log.Print("query:", "nickName err")
		return
	}

	if nickName[0] == "" {
		err = c.WriteMessage(consts.STRING_MSG, []byte("需要填写昵称"))
		if err != nil {
			log.Println("write:", err)
			return
		}
		return
	}

	s.WebsocketConn = c
	_SessionPool[s] = struct{}{}
	err = c.WriteMessage(consts.STRING_MSG, []byte("已连接"))
	if err != nil {
		log.Println("write:", err)
		return
	}

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		boardPostMessage := []byte(s.User.NickName + ":")
		boardPostMessage = append(boardPostMessage, message...)
		for sess, _ := range _SessionPool {
			// TODO 小心nil指针，先暂时不检查
			err = sess.WebsocketConn.WriteMessage(mt, boardPostMessage)
			if err != nil {
				log.Println("write:", err)
				continue
			}
		}
	}
}
