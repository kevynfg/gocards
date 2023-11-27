package game

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type client struct {

	// socket is the web socket for this client.
	socket *websocket.Conn

	// receive is a channel to receive messages from other clients.
	receive chan []byte

	// room is the room this client is chatting in.
	room *Room
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		fmt.Printf("message received from %v, message: %v\n", string(c.socket.RemoteAddr().String()), string(msg))
		c.room.Forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
		fmt.Printf("message sent from %v, message: %v\n", string(c.socket.RemoteAddr().String()), string(msg))
	}
}