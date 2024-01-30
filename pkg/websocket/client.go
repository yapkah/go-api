package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// const (
// 	// Time allowed to write a message to the peer.
// 	writeWait = 10 * time.Second

// 	// Time allowed to read the next pong message from the peer.
// 	pongWait = 60 * time.Second

// 	// Send pings to peer with this period. Must be less than pongWait.
// 	pingPeriod = (pongWait * 9) / 10

// 	// Maximum message size allowed from peer.
// 	maxMessageSize = 512
// )

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

// connection is an middleman between the websocket connection and the hub.

var connID = 1

type connection struct {
	id  int
	mux *sync.Mutex
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (s subscription) readPump() {
	c := s.conn
	defer func() {
		h.unregister <- s
		c.ws.Close()
	}()
	// c.ws.SetReadLimit(maxMessageSize)
	// c.ws.SetReadDeadline(time.Now().Add(pongWait))
	// c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		m := message{msg, s.room}
		h.broadcast <- m
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	// c.mux.Lock()
	// c.mux.Unlock()
	// c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

// // writePump pumps messages from the hub to the websocket connection.
// func (s *subscription) writePump() {
// 	c := s.conn
// 	// ticker := time.NewTicker(pingPeriod)
// 	// defer func() {
// 	// 	ticker.Stop()
// 	// 	c.ws.Close()
// 	// }()
// 	for {
// 		select {
// 		case message, ok := <-c.send:
// 			if !ok {
// 				c.write(websocket.CloseMessage, []byte{})
// 				return
// 			}
// 			if err := c.write(websocket.TextMessage, message); err != nil {
// 				return
// 			}
// 			// case <-ticker.C:
// 			// 	if err := c.write(websocket.PingMessage, []byte{}); err != nil {
// 			// 		return
// 			// 	}
// 		}
// 	}
// }

// func Connection(roomId []string, ws *websocket.Conn) {
func Connection(ws *websocket.Conn) {
	var mux sync.Mutex
	c := &connection{id: connID, mux: &mux, send: make(chan []byte, 256), ws: ws}
	connID++
	// for _, v := range roomId {
	// 	s := subscription{c, v}
	// 	h.register <- s
	// 	go s.writePump()
	// 	go s.readPump()
	// }

	s1 := subscription{c, "exchange_price"}
	h.register <- s1
	// go s1.writePump()
	// go s1.readPump()

	// c2 := &connection{send: make(chan []byte, 256), ws: ws}
	s2 := subscription{c, "market_price"}
	h.register <- s2
	// go s2.writePump()
	// go s2.readPump()
	s3 := subscription{c, "available_buy_market_price"}
	h.register <- s3

	s4 := subscription{c, "available_sell_market_price"}
	h.register <- s4

	go c.writePump()
}
func BroadcastData(roomId string, msg []byte) {
	m := message{msg, roomId}
	h.broadcast <- m
}

// serveWs handles websocket requests from the peer.
// func serveWs(w http.ResponseWriter, r *http.Request, roomId string) {
// 	fmt.Print(roomId)
// 	ws, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println(err.Error())
// 		return
// 	}
// 	c := &connection{send: make(chan []byte, 256), ws: ws}
// 	s := subscription{c, roomId}
// 	h.register <- s
// 	go s.writePump()
// 	go s.readPump()
// }
