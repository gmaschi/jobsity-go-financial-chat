package chatmessages

import (
	"encoding/json"
	"fmt"
	financialbot "github.com/gmaschi/jobsity-go-financial-chat/internal/services/financial-bot"
	financialconsumer "github.com/gmaschi/jobsity-go-financial-chat/internal/services/financial-consumer"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
)

type (
	// Hub manages the set of active connections for each room
	Hub struct {
		rooms      map[string]map[*connection]bool
		broadcast  chan message
		register   chan Subscription
		unregister chan Subscription
	}

	// message defines the basic message structure
	message struct {
		data     []byte
		room     string
		username string
	}
)

var hub = Hub{
	rooms:      make(map[string]map[*connection]bool),
	broadcast:  make(chan message),
	register:   make(chan Subscription),
	unregister: make(chan Subscription),
}

func init() {
	go hub.Start()
}

// ServeWs serves the websocket connection for a given user at a given room
func ServeWs(w http.ResponseWriter, r *http.Request, roomId, username string) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	conn := &connection{ws: ws, send: make(chan []byte, 256)}
	sub := Subscription{conn, roomId}
	hub.register <- sub

	go sub.writeHubToConnection()
	go sub.readConnectionToHub(username)
}

// Start starts the hub
func (h *Hub) Start() {
	financialConsumer := financialconsumer.New()
	defer financialConsumer.Conn.Close()
	defer financialConsumer.Ch.Close()

	var formattedMessage string
	for {
		select {
		case s := <-h.register:
			connections := h.rooms[s.room]
			if connections == nil {
				connections = make(map[*connection]bool)
				h.rooms[s.room] = connections
			}

			h.rooms[s.room][s.conn] = true

		case s := <-h.unregister:
			connections := h.rooms[s.room]
			if connections != nil {
				if _, ok := connections[s.conn]; ok {
					delete(connections, s.conn)
					close(s.conn.send)
					if len(connections) == 0 {
						delete(h.rooms, s.room)
					}
				}
			}

		case m := <-h.broadcast:
			formattedMessage = fmt.Sprintf("%s: %s", m.username, m.data)
			deliverMessagesToConnections(h, formattedMessage, m.room)

		case m, ok := <-financialConsumer.Messages:
			if ok {
				var payload financialbot.Payload
				err := json.Unmarshal(m.Body, &payload)
				if err != nil {
					formattedMessage = fmt.Sprintf("financial-bot: could not retrieve stock data")
				} else {
					formattedMessage = payload.Message
				}

				deliverMessagesToConnections(h, formattedMessage, payload.Room)
			}
		}
	}
}

func isStockCommand(message string) (string, bool) {
	message = strings.TrimSpace(message)
	stockCommandPrefix := "/stock="
	if !strings.HasPrefix(message, stockCommandPrefix) {
		return "", false
	}

	stock := strings.TrimPrefix(message, stockCommandPrefix)

	return stock, true
}

func deliverMessagesToConnections(h *Hub, message, room string) {
	connections := h.rooms[room]
	for c := range connections {
		select {
		case c.send <- []byte(message):
		default:
			close(c.send)
			delete(connections, c)
			if len(connections) == 0 {
				delete(h.rooms, room)
			}
		}
	}
}
