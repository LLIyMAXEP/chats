package chat

import (
	"golang.org/x/net/websocket"
	"io"
	"log"
)

type Client struct {
	ID         string          `json:"id"`
	Online     bool            `json:"online"`
	Connection *websocket.Conn `json:"-"`
	Server     *Hub            `json:"-"`
}

func (c *Client) Listen() {
	c.listenReceive()
}

func (c *Client) listenReceive() {
	for {
		var msg Message
		err := websocket.JSON.Receive(c.Connection, &msg)

		if err == io.EOF {
			log.Println("Client disconnected")
			c.Server.deleteClient(c.ID)
			return
		}

		log.Println("Message received: ", msg)

		switch msg.Type {
		case "directMsg":
			{
				c.Server.sendDirectMsg(msg)
			}
		case "channelMsg":
			{
				c.Server.sendMsgToChannel(msg)
			}
		case "commitMsg":
			{
				c.Server.sendCommitMsg(msg)
			}
		}
	}
}
