package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var wsConnMap = make(map[string]*websocket.Conn)
var kkk = ""

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("长链接失败", err)
		return
	}

	if conn != nil {
		kkk = r.Header.Get("Sec-Websocket-Key")
		wsConnMap[kkk] = conn
	}

	go func() {
		for {
			time.Sleep(time.Second * 1)
			_conn := wsConnMap[kkk]
			if _conn != nil {
				_conn.WriteMessage(1, []byte("pong"))
			}
		}
	}()

	for {
		t, msg, readMessageErr := conn.ReadMessage()
		if readMessageErr != nil {
			break
		}
		writeMessageErr := conn.WriteMessage(t, msg)
		if writeMessageErr != nil {
			return
		}
	}
}

func Ws(c *gin.Context) {
	wsHandler(c.Writer, c.Request)
}
