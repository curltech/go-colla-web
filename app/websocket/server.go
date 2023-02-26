package websocket

/**
基于iris websocket的服务端，采用原始报文模式，融入iris的http服务器,支持tls
*/

import (
	"errors"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/message"
	gorillaws "github.com/gorilla/websocket"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/websocket"
	"github.com/kataras/neffos/gorilla"
	"net/http"
)

// Almost all features of neffos are disabled because no custom message can pass
// when app expects to accept and send only raw websocket native messages.
// When only allow native messages is a fact?
// When the registered namespace is just one and it's empty
// and contains only one registered event which is the `OnNativeMessage`.
// When `Events{...}` is used instead of `Namespaces{ "namespaceName": Events{...}}`
// then the namespace is empty "".

/**
iris设置websocket的支持
*/

var ConnectionPool = make(map[string]*websocket.Conn, 0)

var ConnectionIndex = make(map[string]map[string]string, 0)

func Set(app *iris.Application) {
	//设置websocket的参数
	upgrader := new(gorillaws.Upgrader)
	upgrader.ReadBufferSize = config.ServerWebsocketParams.ReadBufferSize
	upgrader.WriteBufferSize = config.ServerWebsocketParams.WriteBufferSize
	path := config.ServerWebsocketParams.Path
	upgrader.CheckOrigin = func(r *http.Request) bool {
		if r.Method != "POST" && r.Method != "GET" {
			logger.Sugar.Errorf("method is not GET/POST")
			return false
		}
		if r.URL.Path != path {
			logger.Sugar.Errorf("path error, must be /websocket")
			return false
		}
		return true
	}
	neffosUpgrader := gorilla.Upgrader(*upgrader)
	ws := websocket.New(neffosUpgrader, websocket.Events{
		websocket.OnNativeMessage: OnNativeMessage,
	})

	/**
	需要记录连接*websocket.Conn，用于发送消息
	*/
	ws.OnConnect = func(c *websocket.Conn) error {
		id := c.ID()
		_, ok := ConnectionPool[id]
		if ok {
			logger.Sugar.Errorf("connection: %v exist!", id)

			return errors.New("Exist")
		}
		GetWebsocketConnPool().Connect(c)
		ConnectionPool[id] = c
		logger.Sugar.Infof("[%s] Connected to server!", c.ID())
		/*msg := websocket.Message{
			Body:  []byte(id),
			Event: "onNewConnect", // fire the "onNewConnect" client event.
		}
		c.Server().Broadcast(c, msg)*/
		var msgBody = make(map[string]interface{}, 0)
		msgBody["contentType"] = "ConnectSessionId"
		//msgBody["message"] = id + "," + global.Global.MyselfPeer.PeerId + "," + global.Global.MyselfPeer.PublicKey
		data, err := message.Marshal(msgBody)
		if err != nil {
			logger.Sugar.Errorf("Marshal failure")
		}
		SendRaw(id, data)

		return nil
	}

	ws.OnDisconnect = func(c *websocket.Conn) {
		id := c.ID()
		_, ok := ConnectionPool[id]
		if !ok {
			logger.Sugar.Errorf("connection: %v not exist!", id)

			return
		}
		GetWebsocketConnPool().Disconnect(id)
		delete(ConnectionPool, id)
		logger.Sugar.Infof("[%s] Disconnected from server", id)
		/*message := websocket.Message{
			Body:  []byte(id),
			Event: "onQuitConnect", // fire the "onQuitConnect" client event.
		}
		c.Server().Broadcast(c, message)*/
	}
	ws.OnUpgradeError = func(err error) {
		logger.Sugar.Errorf("Upgrade Error: %v", err)
	}

	// register the server on an endpoint.
	// see the inline javascript code i the websockets.html, this endpoint is used to connect to the server.
	app.Any(path, websocket.Handler(ws))
}

/**
处理接收的消息
*/
func OnNativeMessage(nsConn *websocket.NSConn, msg websocket.Message) error {
	logger.Sugar.Infof("Server got: %s from [%s]", msg.Body, nsConn.Conn.ID())
	var msgBody = make(map[string]interface{}, 0)
	err := message.Unmarshal(msg.Body, &msgBody)
	if err != nil {
		return err
	}
	id := nsConn.Conn.ID()
	//remoteAddr:=nsConn.Conn.Socket().Request().RemoteAddr
	if msgBody["contentType"] == "Heartbeat" {
		_, ok := ConnectionPool[id]
		if !ok {
			ConnectionPool[id] = nsConn.Conn
			logger.Sugar.Infof("heartbeat: %v reset connectionPool!", id)
		}
	} else {
		// HandleChainMessage(msg.Body,remoteAddr) or HandlePCChainMessage
		//response, err := receiver.HandlePCChainMessage(msg.Body)
		//if err != nil {
		//	return err
		//}
		//if response != nil {
		//	/*message := websocket.Message{
		//		Body: response,
		//	}
		//	nsConn.Conn.Write(message)*/
		//}
	}

	return nil
}

func SendRaw(id string, data []byte) {
	conn, ok := ConnectionPool[id]
	if !ok {
		logger.Sugar.Errorf("connection: %v not exist!", id)

		return
	}
	message := websocket.Message{
		Body: data,
	}
	conn.Write(message)
}
