package chatservice

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/gorilla/websocket"
)

const channelBufSize = 100

var maxId int = 0

// Chat client.
type Client struct {
	id      int
	ws      *websocket.Conn
	service *Service
	ch      chan *Message
	doneCh  chan bool
}

// Create new chat client.
func NewClient(ws *websocket.Conn, service *Service) *Client {

	if ws == nil {
		slog.Error("ws cannot be nil")
	}

	if service == nil {
		slog.Error("server cannot be nil")
	}

	maxId++
	ch := make(chan *Message, channelBufSize)
	doneCh := make(chan bool)

	return &Client{
		maxId,
		ws,
		service,
		ch,
		doneCh,
	}
}

func (c *Client) Conn() *websocket.Conn {
	return c.ws
}

func (c *Client) Write(msg *Message) {
	select {
	case c.ch <- msg:
	default:
		c.service.Del(c)
		err := fmt.Errorf("client %d is disconnected", c.id)
		c.service.Err(err)
	}
}

func (c *Client) Done() {
	c.doneCh <- true
}

func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

func (c *Client) listenWrite() {
	for {
		select {

		case msg := <-c.ch:
			message, err := json.Marshal(msg)
			if err != nil {
				c.service.Err(err)
				c.doneCh <- true
				return
			}
			err = c.Conn().WriteMessage(websocket.TextMessage, message)
			if err != nil {
				c.service.Err(err)
				c.doneCh <- true
				return
			}

		case <-c.doneCh:
			c.service.Del(c)
			c.doneCh <- true
			return
		}
	}
}

// Listen read request via chanel
func (c *Client) listenRead() {
	for {
		select {

		case <-c.doneCh:
			c.service.Del(c)
			c.doneCh <- true // for listenWrite method
			return

		default:
			var msg Message
			// messageType
			_, message, err := c.Conn().ReadMessage()
			if err == io.EOF {
				c.doneCh <- true
			} else if err != nil {
				c.service.Err(err)
				return
			} else {
				json.Unmarshal(message, &msg)
				c.service.SendAll(&msg)
			}
		}
	}
}
