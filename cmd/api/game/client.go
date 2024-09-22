package game

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type CreateGame struct {
	GameDuration  int    `json:"gameMinutesDuration"`
	RoundDuration int    `json:"roundSecondsDuration"`
	Topic         string `json:"topic"`
	MaxPlayers    int    `json:"maxPlayers"`
	RoomAdmin     string `json:"roomAdmin,omitempty"`
}

type UpdateRoom struct {
	GameDuration  int    `json:"gameDuration,omitempty"`
	RoundDuration int    `json:"roundDuration,omitempty"`
	Topic         string `json:"topic,omitempty"`
	MaxPlayers    int    `json:"maxPlayers,omitempty"`
	RoomAdmin     string `json:"roomAdmin,omitempty"`
}

type client struct {

	// socket is the web socket for this client.
	socket *websocket.Conn

	// receive is a channel to receive messages from other clients.
	receive chan []byte

	// room is the room this client is chatting in.
	room *Room
}

func (c *client) read() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from panic", r)
		}
	}()

	defer c.socket.Close()
	for {
		msgType, msg, err := c.socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read: unexpected close error: %v", err)
			} else {
				log.Printf("read: failed to read message from client: %v", err)
			}
			return
		}
		fmt.Printf("message received from client %v, message: %v\n, messageType: %v\n", string(c.socket.RemoteAddr().String()), string(msg), msgType)

		var message Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			fmt.Println("error trying to Unmarshal: ", err)
			return
		}
		switch message.Type {
		case "create-room":
			fmt.Println("create-room", message.Data)
			c.sendPNGFile()
		case "choose-topic":
			fmt.Println("topic chosen")
			msg = []byte("topic-chosen")
		case "connection":
			fmt.Println("connection")
			msg = []byte("confirmation")
		case "ping":
			fmt.Println("ping")
			msg = []byte("pong")
		}
		c.room.Forward <- msg
	}
}

func (c *client) write() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered in write: ", r)
		}
	}()
	defer c.socket.Close()
	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("write: failed to write message to client", err)
			return
		}
		fmt.Printf("message sent from %v, message: %v\n", string(c.socket.RemoteAddr().String()), string(msg))
	}
}
