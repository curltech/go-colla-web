package websocket

/**
基于iris websocket的客户端，采用原始报文模式
*/
import (
	"context"
	"github.com/curltech/go-colla-core/logger"
	"github.com/kataras/iris/v12/websocket"
	"net/url"
	"sync"
	"time"
)

type WsClient struct {
	conn    *websocket.NSConn
	schema  string
	addr    *string
	path    string
	isAlive bool
	timeout time.Duration
}

// 构造函数
func NewWsClient(schema string, ip string, port string, path string, timeout time.Duration) *WsClient {
	addrString := ip + ":" + port
	var conn *websocket.NSConn
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &WsClient{
		schema:  schema,
		addr:    &addrString,
		path:    path,
		conn:    conn,
		isAlive: false,
		timeout: timeout,
	}
}

// this can be shared with the server.go's.
// `NSConn.Conn` has the `IsClient() bool` method which can be used to
// check if that's is a client or a server-side callback.
func (this *WsClient) dail() {
	var err error
	uri := url.URL{Scheme: this.schema, Host: *this.addr, Path: this.path}
	logger.Sugar.Infof("connecting to %s", uri.String())

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(this.timeout))
	defer cancel()

	// username := "my_username"
	// dialer := websocket.GobwasDialer(websocket.GobwasDialerOptions{Header: websocket.GobwasHeader{"X-Username": []string{username}}})
	dialer := websocket.DefaultGobwasDialer
	clientEvents := websocket.Events{
		websocket.OnNativeMessage: this.OnClientNativeMessage,
	}
	client, err := websocket.Dial(ctx, dialer, uri.String(), clientEvents)
	if err != nil {
		logger.Sugar.Errorf("%v", err)

		return
	}
	defer client.Close()

	this.conn, err = client.Connect(ctx, "")
	if err != nil {
		logger.Sugar.Errorf("%v", err)

		return
	}
	this.isAlive = true
	logger.Sugar.Infof("connecting to %s 链接成功！！！", uri.String())
}

func (this *WsClient) Disconnect() {
	if err := this.conn.Disconnect(nil); err != nil {
		logger.Sugar.Errorf("reply from server: %v", err)
	}
}

func (this *WsClient) Send(msg websocket.Message) {
	this.conn.Conn.Write(msg)
	this.conn.Emit(websocket.OnNativeMessage, []byte("Hello from Go client side!"))
}

/**
处理接收的消息
*/
func (this *WsClient) OnClientNativeMessage(nsConn *websocket.NSConn, msg websocket.Message) error {
	logger.Sugar.Infof("Server got: %s from [%s]", msg.Body, nsConn.Conn.ID())
	nsConn.Conn.Write(msg)

	return nil
}

func (this *WsClient) Connect(ip string, port string, path string, timeout time.Duration) {
	this = NewWsClient("wss", ip, port, path, timeout)
	if this.isAlive == false {
		this.dail()
	}
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
