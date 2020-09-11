package server

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type client struct {
	networkConnection *websocket.Conn
	jsonSendChannel   chan interface{}
	onUnregister      func(*client)
}

func (c *client) activateOutgoingMessages() {
	defer func() {
		_ = c.networkConnection.Close()
		c.onUnregister(c)
	}()
	for {
		select {
		case message, ok := <-c.jsonSendChannel:
			if !ok {
				_ = c.networkConnection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.networkConnection.WriteJSON(message)
			if err != nil {
				return
			}
		}
	}
}

type WebInterface struct {
	clients map[*client]bool
}

func NewWebInterface() WebInterface {
	return WebInterface{clients: make(map[*client]bool)}
}

func (w *WebInterface) SocketHandler(writer http.ResponseWriter, request *http.Request) {
	setupCORS(&writer, request)
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	var client = &client{
		networkConnection: conn,
		jsonSendChannel:   make(chan interface{}, 256),
		onUnregister: func(unregisteredClient *client) {
			delete(w.clients, unregisteredClient)
		},
	}
	w.clients[client] = true

	go client.activateOutgoingMessages()
}

func setupCORS(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func (w *WebInterface) BroadcastJson(v interface{}) {
	for client, _ := range w.clients {
		client.jsonSendChannel <- v
	}
}
