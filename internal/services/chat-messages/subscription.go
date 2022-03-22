package chatmessages

import (
	financialbot "github.com/gmaschi/jobsity-go-financial-chat/internal/services/financial-bot"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type Subscription struct {
	conn *connection
	room string
}

// readConnectionToHub reads messages from the socket connection and sends it to the hub
func (s Subscription) readConnectionToHub(username string) {
	c := s.conn
	defer func() {
		hub.unregister <- s
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		m := message{msg, s.room, username}
		hub.broadcast <- m

		if stock, ok := isStockCommand(string(msg)); ok {
			err = financialbot.GetStockData(stock, s.room)
			if err != nil {
				botErrMessage := message{[]byte(err.Error()), s.room, "financial-bot"}
				hub.broadcast <- botErrMessage
			}
		}
	}
}

// writeHubToConnection writes messages from the hub to the connection socket
func (s *Subscription) writeHubToConnection() {
	c := s.conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
