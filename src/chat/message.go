package chat

import "time"

type Message struct {
	ID       string      `json:"id"`
	Received bool        `json:"received"`
	Type     string      `json:"type"`
	Created  time.Time   `json:"created"`
	Sender   *Client     `json:"sender"`
	Receiver *Client     `json:"receiver"`
	Channel  *Channel    `json:"channel"`
	Payload  interface{} `json:"payload"`
}

func (m *Message) getChannelName() string {
	if m.Channel != nil {
		return m.Channel.ID
	}
	if m.Receiver.ID < m.Sender.ID {
		return m.Receiver.ID + "_" + m.Sender.ID
	} else {
		return m.Sender.ID + "_" + m.Receiver.ID
	}
}

type SimplePayload struct {
	Text string `json:"text"`
}

type InitPayload struct {
	PublicChannels []*Channel `json:"public_channels"`
	DirectChannels []*Channel `json:"direct_channels"`
	Clients        []*Client  `json:"clients"`
}
