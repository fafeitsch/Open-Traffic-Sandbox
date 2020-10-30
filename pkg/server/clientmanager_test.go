package server

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestWebInterface_BroadcastJson(t *testing.T) {
	webInterface := NewClientContainer()
	server := httptest.NewServer(webInterface)
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
	messages2 := make([]string, 3)
	go func() {
		_, data, _ := client1.ReadMessage()
		messages1[0] = string(data)
		_, data, _ = client1.ReadMessage()
		messages1[1] = string(data)
		_ = client1.Close()
		wg.Done()
	}()
	go func() {
		_, data, _ := client2.ReadMessage()
		messages2[0] = string(data)
		_, data, _ = client2.ReadMessage()
		messages2[1] = string(data)
		_, data, _ = client2.ReadMessage()
		messages2[2] = string(data)
		wg.Done()
	}()

	webInterface.BroadcastJson("hello there")
	webInterface.BroadcastJson(42)
	webInterface.BroadcastJson("only one")
	webInterface.Close()
	wg.Wait()
	assert.Equal(t, []string{"\"hello there\"\n", "42\n"}, messages1, "client1 received wrong messages")
	assert.Equal(t, []string{"\"hello there\"\n", "42\n", "\"only one\"\n"}, messages2, "client2 received wrong messages")
}

func TestWebInterface_SocketHandler(t *testing.T) {
	webInterface := NewClientContainer()
	request := httptest.NewRequest("POST", "http://127.0.0.1:8080/sockets", strings.NewReader("websocket intialization"))
	recorder := httptest.NewRecorder()
	webInterface.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusBadRequest, recorder.Code, "response code is wrong")
	message, err := ioutil.ReadAll(recorder.Body)
	require.NoError(t, err)
	assert.Equal(t, "Bad Request\n", string(message), "error message not correct")
}
