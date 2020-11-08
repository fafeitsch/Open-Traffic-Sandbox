package server

import (
	"github.com/gorilla/websocket"
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
			_ = c.networkConnection.WriteJSON(message)
		}
	}
}

// ClientContainer manages all websocket clients and is responsible for sending updates to the client.
// New client containers should be created with the NewClientContainer method.
type ClientContainer struct {
	clients map[*client]bool
}

// NewClientContainer creates a new ClientContainer.
func NewClientContainer() *ClientContainer {
	return &ClientContainer{clients: make(map[*client]bool)}
}

// ServeHTTP registers new websocket clients to the container. All registered clients will receive updates
// when the BroadcastJson method is called on the client container.
func (w *ClientContainer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	setupCORS(&writer, request)
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		// we do nothing here because the upgrader internally has already notified the client about the error.
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

// BroadcastJson encodes the passed interface as JSON and sends it to all currently registered clients.
func (w *ClientContainer) BroadcastJson(v interface{}) {
	for client, _ := range w.clients {
		client.jsonSendChannel <- v
	}
}

// Close releases all client connections of the current clients.
func (w *ClientContainer) Close() error {
	for client, _ := range w.clients {
		close(client.jsonSendChannel)
	}
	return nil
}
