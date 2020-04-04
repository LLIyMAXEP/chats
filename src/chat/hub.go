package chat

import (
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
)

type Hub struct {
	PublicChannels map[string]*Channel
	DirectChannels map[string]*Channel
	Clients        map[string]*Client
}

func CreateHub() *Hub {
	channels := make(map[string]*Channel)
	channels["broadcast"] = &Channel{
		ID:       "broadcast",
		Clients:  make(map[string]*Client),
		Messages: make([]*Message, 0),
	}
	clients := make(map[string]*Client)
	clients["1"] = &Client{
		ID:         "1",
		Connection: nil,
		Server:     nil,
	}
	clients["2"] = &Client{
		ID:         "2",
		Connection: nil,
		Server:     nil,
	}
	clients["3"] = &Client{
		ID:         "3",
		Connection: nil,
		Server:     nil,
	}
	return &Hub{
		PublicChannels: channels,
		Clients:        clients,
		DirectChannels: make(map[string]*Channel),
	}
}

func (h *Hub) Listen() {
	onConnected := func(ws *websocket.Conn) {
		userId := ws.Request().FormValue("user")
		client := h.addUser(userId, ws)
		for _, channel := range h.PublicChannels {
			channel.Clients[userId] = client
		}
		h.sendInit(client)
		client.Listen()
	}
	http.Handle("/hub", websocket.Handler(onConnected))
}

// send initialize information to connected client
func (h *Hub) sendInit(client *Client) {
	clientsList := make([]*Client, 0, len(h.Clients))
	for _, val := range h.Clients {
		clientsList = append(clientsList, val)
	}

	// to list (slice) for frontend
	directChannelsList := make([]*Channel, 0, len(h.DirectChannels))
	for _, channel := range h.DirectChannels {
		var toUserId string
		for _, toClient := range channel.Clients {
			if toClient.ID != client.ID {
				toUserId = toClient.ID
			}
		}
		if _, ok := channel.Clients[client.ID]; ok {
			copyChannel := &Channel{
				ID:       channel.ID,
				UserID:   toUserId,
				Messages: channel.Messages,
			}
			directChannelsList = append(directChannelsList, copyChannel)
		}
	}

	// to list (slice) for frontend
	publicChannelsList := make([]*Channel, 0, len(h.PublicChannels))
	for _, channel := range h.PublicChannels {
		publicChannelsList = append(publicChannelsList, channel)
	}
	message := Message{
		Type:    "init",
		Created: time.Time{},
		Payload: InitPayload{
			Clients:        clientsList,
			DirectChannels: directChannelsList,
			PublicChannels: publicChannelsList,
		},
	}
	websocket.JSON.Send(client.Connection, message)

	h.sendClientStatus(client)
}

// send client statuses to all connected clients
func (h *Hub) sendClientStatus(client *Client) {
	h.sendAll(Message{
		Type:    "clientStatus",
		Created: time.Now(),
		Payload: client,
	})
}

// send message to all connected clients
func (h *Hub) sendAll(msg Message) {
	for _, client := range h.Clients {
		if client.Connection != nil {
			websocket.JSON.Send(client.Connection, msg)
		} else {
			log.Printf("Client %v not connected", client.ID)
		}
	}
}

// send message direct to user and store it in private channel
func (h *Hub) sendDirectMsg(msg Message) {
	if val, ok := h.Clients[msg.Receiver.ID]; ok {
		channelName := msg.getChannelName()
		if channel, ok := h.DirectChannels[channelName]; ok {
			channel.Messages = append(channel.Messages, &msg)
		} else {
			clients := make(map[string]*Client)
			clients[msg.Sender.ID] = msg.Sender
			clients[msg.Receiver.ID] = msg.Receiver
			messages := make([]*Message, 0, 1)
			messages = append(messages, &msg)
			h.DirectChannels[channelName] = &Channel{
				ID:       channelName,
				Clients:  clients,
				Messages: messages,
			}
		}
		if val.Online {
			websocket.JSON.Send(val.Connection, msg)
		}
	}
}

// send commit msg
func (h *Hub) sendCommitMsg(msg Message) {
	channelName := msg.getChannelName()
	if channel, ok := h.DirectChannels[channelName]; ok {
		for _, message := range channel.Messages {
			if message.ID == msg.ID {
				message.Received = true
				break
			}
		}
	} else if msg.Channel != nil {
		if channel, ok := h.PublicChannels[msg.Channel.ID]; ok {
			for _, message := range channel.Messages {
				if message.ID == msg.ID {
					message.Received = true
					break
				}
			}
		}
	}

	if val, ok := h.Clients[msg.Receiver.ID]; ok && val.Online {
		websocket.JSON.Send(val.Connection, msg)
	}
}

// send msg to channel and store it in this channel
func (h *Hub) sendMsgToChannel(msg Message) {
	if channel, ok := h.PublicChannels[msg.Channel.ID]; ok {
		channel.Messages = append(channel.Messages, &msg)
		for _, client := range channel.Clients {
			if msg.Sender.ID != client.ID {
				websocket.JSON.Send(client.Connection, msg)
			}
		}
	}
}

// add user to map of users of this hub and store its connection
func (h *Hub) addUser(id string, ws *websocket.Conn) *Client {
	if val, ok := h.Clients[id]; ok {
		val.Connection = ws
		val.Server = h
		val.Online = true
		log.Printf("Client %v connected", id)
		return val
	} else {
		client := &Client{
			ID:         id,
			Connection: ws,
			Server:     h,
		}
		log.Printf("Client %v created", id)
		return client
	}
}

// remove user connection from hub
func (h *Hub) deleteClient(id string) {
	if val, ok := h.Clients[id]; ok {
		val.Connection = nil
		val.Online = false
		h.sendClientStatus(val)
		log.Printf("Client %v offline", id)
	} else {
		log.Printf("Client %v not exists", id)
	}
}
