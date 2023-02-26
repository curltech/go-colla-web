package websocket

import (
	"github.com/curltech/go-colla-core/logger"
	"github.com/kataras/iris/v12/websocket"
)

type WebsocketConnPool struct {
	pool map[string]*websocket.Conn
}

var websocketConnPool = &WebsocketConnPool{pool: make(map[string]*websocket.Conn)}

func GetWebsocketConnPool() *WebsocketConnPool {
	return websocketConnPool
}

func (this *WebsocketConnPool) Connect(conn *websocket.Conn) {
	if conn != nil {
		logger.Sugar.Infof("remote peer:%v", conn.ID())
		_, ok := this.pool[conn.ID()]
		if !ok {
			this.pool[conn.ID()] = conn
		}
	}
}

func (this *WebsocketConnPool) GetRemoteConn(connectSessionId string) *websocket.Conn {
	p, ok := this.pool[connectSessionId]
	if ok {
		return p
	}

	return nil
}

func (this *WebsocketConnPool) Disconnect(connectSessionId string) {
	_, ok := this.pool[connectSessionId]
	if ok {
		delete(this.pool, connectSessionId)
	}
}
