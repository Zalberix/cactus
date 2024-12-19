package chatservice

import (
	"fmt"
	"log/slog"

	"github.com/gorilla/websocket"
)

type Service struct {
	upgrader  *websocket.Upgrader
	messages  []*Message
	clients   map[int]*Client
	addCh     chan *Client
	delCh     chan *Client
	sendAllCh chan *Message
	doneCh    chan bool
	errCh     chan error
}

func NewService() *Service {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	messages := []*Message{}
	clients := make(map[int]*Client)
	addCh := make(chan *Client)
	delCh := make(chan *Client)
	sendAllCh := make(chan *Message)
	doneCh := make(chan bool)
	errCh := make(chan error)

	return &Service{
		&upgrader,
		messages,
		clients,
		addCh,
		delCh,
		sendAllCh,
		doneCh,
		errCh,
	}
}

func (s *Service) Add(c *Client) {
	s.addCh <- c
}

func (s *Service) Del(c *Client) {
	s.delCh <- c
}

func (s *Service) SendAll(msg *Message) {
	s.sendAllCh <- msg
}

func (s *Service) Done() {
	s.doneCh <- true
}

func (s *Service) Err(err error) {
	s.errCh <- err
}

func (s *Service) Upgrader() *websocket.Upgrader {
	return s.upgrader
}

func (s *Service) sendPastMessages(c *Client) {
	for _, msg := range s.messages {
		c.Write(msg)
	}
}

func (s *Service) sendAll(msg *Message) {
	for _, c := range s.clients {
		c.Write(msg)
	}
}

func (s *Service) Listen() {

	for {
		select {

		case c := <-s.addCh:
			s.clients[c.id] = c
			s.sendPastMessages(c)

		case c := <-s.delCh:
			delete(s.clients, c.id)

		case msg := <-s.sendAllCh:
			s.messages = append(s.messages, msg)
			s.sendAll(msg)

		case err := <-s.errCh:
			slog.Error(fmt.Sprintf("Error: %s", err.Error()))

		case <-s.doneCh:
			return
		}
	}
}
