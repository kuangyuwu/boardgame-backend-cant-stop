package server

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kuangyuwu/boardgame-backend-cant-stop/internal/clog"
)

type WebsocketManager interface {
	Connect(in <-chan []byte, out chan<- []byte, dc <-chan struct{}) error
}

func generateHandlerWebsocket(m WebsocketManager) http.HandlerFunc {
	handlerWebsocket := func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				ok := origin == "http://127.0.0.1:5500"
				if !ok {
					clog.Warnf("CheckOrigin failed: request origin = %s", origin)
				}
				return ok
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			clog.Warnf("webSocket upgrade request failed: %s\n", err)
			return
		}
		defer conn.Close()

		dc := make(chan struct{})
		defer close(dc)
		in, out := make(chan []byte), make(chan []byte)
		defer close(in)

		err = m.Connect(in, out, dc)
		if err != nil {
			clog.Errorf("error connecting to websocket manager: %s\n", err)
			return
		}

		websocketKey := r.Header.Get("Sec-Websocket-Key")
		clog.Infof("a websocket connection established: %s", websocketKey)

		go writeMessage(conn, out, dc)
		readMessage(conn, in)

		clog.Infof("a websocket connection disconnected: %s", websocketKey)
	}
	return handlerWebsocket
}

func readMessage(conn *websocket.Conn, in chan<- []byte) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				clog.Errorf("error reading message: %v", err)
			}
			return
		}
		in <- msg
	}
}

func writeMessage(conn *websocket.Conn, out <-chan []byte, dc <-chan struct{}) {
	for {
		select {
		case msg := <-out:
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				clog.Errorf("error writing message: %v", err)
			}
		case <-dc:
			clog.Info("stop writing message")
			return
		}
	}
}
