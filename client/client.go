package client

import (
	"ChatServer/config"
	logger "ChatServer/logger"
	manager "ChatServer/manager"
	"ChatServer/protobuf"
	"ChatServer/status"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

// Client : 每一個連進來的client
type Client struct {
	name            string
	conn            *websocket.Conn
	messageHandlers map[string]func(*protobuf.Header) int32
}

// InitClient 是一個對外創建 Client 的接口
func InitClient(conn *websocket.Conn) {
	client := &Client{}
	go client.init(conn)
}

func (client *Client) init(conn *websocket.Conn) {
	client.name = "Little chicken"
	client.conn = conn

	//client.conn.SetWriteDeadline(time.Now().Add(config.SocketTimeout * time.Second))

	client.messageHandlers = map[string]func(*protobuf.Header) int32{
		"Header_HeartBeat":     client.onHeartBeat,
		"Header_SetNameReq":    client.onSetNameReq,
		"Header_MessageReq":    client.onMessageReq,
		"Header_MessageNotify": client.onMessageNotify}

	outChan := make(chan *protobuf.Header, 10)
	if manager.DoLogin(outChan) != status.OK {
		logger.ErrorLog("client login fail")
		return
	}

	go client.worker(outChan)
}

func (client *Client) worker(outChan chan *protobuf.Header) {
	defer func() {
		client.conn.Close()
		manager.DoLogout(outChan)
	}()
	TAG := "[Clinet.worker] "

	inChan := make(chan *protobuf.Header)

	go inputWorker(inChan, client.conn)

	for {
		select {
		case req, ok := <-inChan:
			if ok == false {
				logger.InfoLog(TAG, "Client", client.name, "Disconnect:")
				return
			}
			client.parseMessage(req)
		case res, ok := <-outChan:
			if ok == false {
				logger.InfoLog(TAG, "Client", client.name, "left chat room:")
				return
			}
			resPacket, _ := proto.Marshal(res)
			client.conn.WriteMessage(websocket.TextMessage, resPacket)
		}
	}
}

func inputWorker(inChan chan<- *protobuf.Header, conn *websocket.Conn) {
	defer func() {
		close(inChan)
	}()
	TAG := "[Clinet.inputWorker] "

	for {
		conn.SetReadDeadline(time.Now().Add(config.SocketTimeout * time.Second))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			logger.InfoLog(TAG, "A client disconnect:", err)
			break
		}
		var req protobuf.Header
		proto.Unmarshal(msg, &req)
		inChan <- &req
	}
}

func (client *Client) parseMessage(req *protobuf.Header) int32 {
	TAG := "[Clinet.parseMessage] "
	containType := reflect.ValueOf(req.Contain).Elem().Type().Name()

	if _, exist := client.messageHandlers[containType]; !exist {
		logger.InfoLog(TAG, client.name, "send", containType, "is an unknown packet")
		res := setResponse(status.InvalidReq, req.Uuid)
		resPacket, _ := proto.Marshal(res)
		client.conn.WriteMessage(websocket.TextMessage, resPacket)
		return status.InvalidReq
	}

	logger.InfoLog(TAG, client.name, "send:", containType)
	res := client.messageHandlers[containType](req)
	logger.InfoLog(TAG, client.name, containType, "handle result", res)

	return res
}

func setResponse(code int32, uuid string) *protobuf.Header {
	return &protobuf.Header{
		Code: code,
		Uuid: uuid,
	}
}

func (client *Client) onHeartBeat(req *protobuf.Header) int32 {
	res := setResponse(status.OK, req.Uuid)
	resPacket, _ := proto.Marshal(res)
	client.conn.WriteMessage(websocket.TextMessage, resPacket)
	return status.OK
}

func (client *Client) onSetNameReq(req *protobuf.Header) int32 {
	client.name = req.GetSetNameReq().Username
	res := setResponse(status.OK, req.Uuid)
	resPacket, _ := proto.Marshal(res)
	client.conn.WriteMessage(websocket.TextMessage, resPacket)

	return status.OK
}

func (client *Client) onMessageReq(req *protobuf.Header) int32 {
	res := setResponse(status.OK, req.Uuid)
	resPacket, _ := proto.Marshal(res)
	client.conn.WriteMessage(websocket.TextMessage, resPacket)

	contain := req.GetMessageReq()

	message := &protobuf.Header{
		Contain: &protobuf.Header_MessageNotify{
			MessageNotify: &protobuf.MessageNotify{
				Username: client.name,
				Message:  contain.Message,
			},
		},
	}

	if manager.DoBroadcast(message) != status.OK {
		logger.ErrorLog("client send message fail")
		return status.TimeOut
	}

	return status.OK
}

func (client *Client) onMessageNotify(req *protobuf.Header) int32 {
	// This shouldn't happen
	return status.InvalidReq
}
