package services

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketService struct {
	subscribers  sync.Map // map[int][]*websocket.Conn - movieID -> list of websocket connections
	StreamStates sync.Map // map[int]map[string]interface{} - movieID -> last stream state
}

func NewWebSocketService() *WebSocketService {
	return &WebSocketService{}
}

func (wc *WebSocketService) AddSubscriber(movieID int, ws *websocket.Conn) {
	var subscribers []*websocket.Conn
	if val, exists := wc.subscribers.Load(movieID); exists {
		subscribers = val.([]*websocket.Conn)
	} else {
		subscribers = make([]*websocket.Conn, 0)
	}
	subscribers = append(subscribers, ws)
	wc.subscribers.Store(movieID, subscribers)
}

func (wc *WebSocketService) RemoveSubscriber(movieID int, ws *websocket.Conn) {
	if val, exists := wc.subscribers.Load(movieID); exists {
		subscribers := val.([]*websocket.Conn)
		for i, conn := range subscribers {
			if conn == ws {
				subscribers = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}

		if len(subscribers) == 0 {
			wc.subscribers.Delete(movieID)
		} else {
			wc.subscribers.Store(movieID, subscribers)
		}
	}
}

func (wc *WebSocketService) UpdateStreamState(movieID int, state map[string]interface{}) {
	wc.StreamStates.Store(movieID, state)

	var subscribers []*websocket.Conn
	if val, exists := wc.subscribers.Load(movieID); exists {
		subscribers = val.([]*websocket.Conn)
	}

	if len(subscribers) > 0 {
		stateJSON, err := json.Marshal(state)
		if err != nil {
			return
		}

		for _, ws := range subscribers {
			ws.WriteMessage(websocket.TextMessage, stateJSON)
		}
	}
}
