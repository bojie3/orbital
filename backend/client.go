package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 409,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3000"
	},
}

// upgrader is used to upgrade HTTP server connection to WebSocket

func newClient(conn *websocket.Conn, wsServer *WsServer, name string) *Client {
	return &Client{
		conn:     conn,
		wsServer: wsServer,
        send:     make(chan []byte),
		rooms:    make(map[*Room]bool),
		Name:     name,
	}
}

//receives http and return http request
func ServeWs(wsServer *WsServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

		if err != nil {
			fmt.Println(err)
			return
		}
		defer conn.Close()
		client := newClient(conn, wsServer, name)
        go client.readPump()
		go client.writePump()
		
		wsServer.register <- client
		fmt.Println("New Client joined the hub!")

	}
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

func (client *Client) disconnect() {
	client.wsServer.unregister <- client

	for room := range client.rooms {
		room.unregister <- client
	}

	close(client.send)
	client.conn.Close()
	fmt.Println("disconnected")

}

func (client *Client) getName() string {
	return client.Name
}

func (client *Client) findClientByName(name string) *Client {
	var foundClient *Client
	for client := range client.wsServer.clients {
		if client.Name == name{
			foundClient = client
			break
		}
	}
	return foundClient
}

func (client *Client) handleNewMessage(jsonMessage []byte) {
	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		fmt.Printf("Error on unmarshal JSON message %s", err)
	}
	message.Sender = client
	switch message.Action {
	case SendMessageAction:
		roomName := message.Target
		if room := client.wsServer.findRoomByName(roomName); room != nil {
			room.broadcast <- &message
		}

	case JoinRoomAction:
		client.handleJoinRoomMessage(message)

	case LeaveRoomAction:
		client.handleLeaveRoomMessage(message)

	case SendPrivateMessage:
		client.handlePrivateMessage(message)
	}

	

}

func (client *Client) handlePrivateMessage(message Message) {
	targetClientName := message.Target
	if targetClient := client.wsServer.findClientByName(targetClientName); targetClient != nil{
		targetClient.send <- message.encode()
	}
}

func (client *Client) handleJoinRoomMessage(message Message) {
	roomName := message.Message

	room := client.wsServer.findRoomByName(roomName)
	if room == nil {
		room = client.wsServer.createRoom(roomName)
	}

	client.rooms[room] = true

	room.register <- client

}

func (client *Client) handleLeaveRoomMessage(message Message) {
	room := client.wsServer.findRoomByName(message.Message)
	if _, ok := client.rooms[room]; ok {
		delete(client.rooms, room)
	}

	room.unregister <- client

}

func (client *Client) readPump() {
	defer func() {
        client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("unexpected close error: %v", err)
			}
			break
		}

		client.handleNewMessage(jsonMessage)
	}
}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
