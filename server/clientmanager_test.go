package server

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestWebInterface_BroadcastJson(t *testing.T) {
	webInterface := NewWebInterface()
	server := httptest.NewServer(http.HandlerFunc(webInterface.SocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	client1, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer func() { _ = client1.Close() }()
	client2, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer func() { _ = client2.Close() }()

	var wg sync.WaitGroup
	wg.Add(2)

	messages1 := make([]string, 2)
	messages2 := make([]string, 2)
	go func() {
		_, data, _ := client1.ReadMessage()
		messages1[0] = string(data)
		_, data, _ = client1.ReadMessage()
		messages1[1] = string(data)
		wg.Done()
	}()
	go func() {
		_, data, _ := client2.ReadMessage()
		messages2[0] = string(data)
		_, data, _ = client2.ReadMessage()
		messages2[1] = string(data)
		wg.Done()
	}()

	webInterface.BroadcastJson("hello there")
	webInterface.BroadcastJson(42)
	wg.Wait()
	assert.Equal(t, []string{"\"hello there\"\n", "42\n"}, messages1, "client1 received wrong messages")
	assert.Equal(t, []string{"\"hello there\"\n", "42\n"}, messages2, "client2 received wrong messages")
}
