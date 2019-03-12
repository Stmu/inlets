package transport

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func NewWebsocketConn(conn *websocket.Conn, pingWaitDuration time.Duration) *WebsocketConn {
	w := WebsocketConn{
		conn:             conn,
		PingWaitDuration: pingWaitDuration,
	}
	w.setupDeadline()

	return &w
}

type WebsocketConn struct {
	sync.RWMutex
	conn             *websocket.Conn
	PingWaitDuration time.Duration
}

func (w *WebsocketConn) WriteMessage(messageType int, data []byte) error {
	w.Lock()
	defer w.Unlock()
	w.conn.SetWriteDeadline(time.Now().Add(w.PingWaitDuration))
	return w.conn.WriteMessage(messageType, data)
}

func (w *WebsocketConn) setupDeadline() {
	w.conn.SetReadDeadline(time.Now().Add(w.PingWaitDuration))
	w.conn.SetPingHandler(func(string) error {
		log.Printf("Ping invoked. Sending pong.")
		w.Lock()
		w.conn.WriteControl(websocket.PongMessage, []byte(""), time.Now().Add(time.Second))
		w.Unlock()
		return w.conn.SetReadDeadline(time.Now().Add(w.PingWaitDuration))
	})
	w.conn.SetPongHandler(func(string) error {
		newDeadline := time.Now().Add(w.PingWaitDuration + time.Second*1)
		log.Printf("Pong invoked extending deadline to: %s\n", newDeadline.String())

		return w.conn.SetReadDeadline(newDeadline)
	})
}

func (w *WebsocketConn) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *WebsocketConn) ReadMessage() (messageType int, p []byte, err error) {
	return w.conn.ReadMessage()
}

func (w *WebsocketConn) Ping() error {
	w.Lock()
	defer w.Unlock()
	if err := w.conn.WriteControl(websocket.PingMessage, []byte(""), time.Now().Add(time.Second)); err != nil {
		log.Println("Error writing ping")
		return err
	}
	log.Println("Wrote ping")
	return nil
}

// func (w *WebsocketConn) NextReader() (messageType int, r io.Reader, err error) {
// 	return w.conn.NextReader()
// }
