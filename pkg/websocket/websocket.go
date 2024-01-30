package websocket

import (
	"log"
	"net/http"

	"github.com/yapkah/go-api/pkg/e"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return ws, err
	}
	return ws, nil
}

// func Writer(conn *websocket.Conn) {
// 	for {
// 		fmt.Println("Sending")
// 		messageType, r, err := conn.NextReader()
// 		if err != nil {
// 			fmt.Println("conn.NextReader")
// 			fmt.Println(err)
// 			return
// 		}
// 		w, err := conn.NextWriter(messageType)
// 		if err != nil {
// 			fmt.Println("conn.NextWriter")
// 			fmt.Println(err)
// 			return
// 		}
// 		if _, err := io.Copy(w, r); err != nil {
// 			fmt.Println("io.Copy")
// 			fmt.Println(err)
// 			return
// 		}
// 		if err := w.Close(); err != nil {
// 			fmt.Println("w.Close")
// 			fmt.Println(err)
// 			return
// 		}
// 	}
// }

func ReadWSMsg(conn *websocket.Conn) ([]byte, error) {
	// for {
	// messageType, msg, err := conn.ReadMessage()
	_, msg, err := conn.ReadMessage()

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_get_info_send", Data: err}
	}

	// if err := conn.WriteMessage(messageType, msg); err != nil {
	// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_get_info_send", Data: err}
	// }

	return msg, nil
	// }
}
