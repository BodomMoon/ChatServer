/*
Package main provide an webSocket over ssl chat service .
*/
package main

import (
	client "ChatServer/client"

	logger "ChatServer/logger"
	manager "ChatServer/manager"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	TAG := "[main] "
	logger.InfoLog(TAG, "server start...")
	logger.Init()
	manager.InitManager()

	http.HandleFunc("/", webSocketHandle)

	err := http.ListenAndServeTLS(":9090", "cert.pem", "key.pem", nil)

	logger.UnInit()
	fmt.Println(err)
}

func webSocketHandle(res http.ResponseWriter, req *http.Request) {
	TAG := "[webSocketHandle] "
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if err != nil {
		logger.ErrorLog(TAG, "websocket create fail:", err)
		http.NotFound(res, req)
		return
	}
	client.InitClient(conn)
}
