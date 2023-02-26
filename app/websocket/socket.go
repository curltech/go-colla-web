package websocket

import (
	"fmt"
	"github.com/curltech/go-colla-core/config"
	socketio "github.com/googollee/go-socket.io"
	"github.com/kataras/iris/v12"
)

/**
整合socket.io和iris，也就是使用iris作为socket.io的服务器实现
*/
func SetSocket(app *iris.Application) {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("connected:", s.ID())
		return nil
	})
	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})
	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})
	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})
	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})
	go server.Serve()
	//defer server.Close()

	path := config.ServerWebsocketParams.Path
	app.HandleMany("GET POST", path+"/{any:path}", iris.FromStd(server))
}
