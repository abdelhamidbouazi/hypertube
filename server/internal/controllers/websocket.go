package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WebSocketController struct {
	upgrader     websocket.Upgrader
	subscribers  sync.Map // map[int][]*websocket.Conn - movieID -> list of websocket connections
	streamStates sync.Map // map[int]map[string]interface{} - movieID -> last stream state
}

func NewWebSocketController() *WebSocketController {
	return &WebSocketController{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for testing (restrict this in production)
				return true
			},
		},
	}
}

// HandleWebSocket handles WebSocket connections for a specific movie
// @Summary WebSocket endpoint for movie streaming updates
// @Description Establishes a WebSocket connection to receive real-time streaming updates for a specific movie
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param movieId path int true "Movie ID"
// @Success 101 {string} string "Switching Protocols"
// @Router /ws/{movieId} [get]
func (wc *WebSocketController) HandleWebSocket(c echo.Context) error {
	movieIDStr := c.Param("movieId")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid movie ID")
	}

	ws, err := wc.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()
	defer wc.removeSubscriber(movieID, ws)

	wc.addSubscriber(movieID, ws)

	if lastState, exists := wc.streamStates.Load(movieID); exists {
		if stateJSON, err := json.Marshal(lastState); err == nil {
			ws.WriteMessage(websocket.TextMessage, stateJSON)
		}
	}

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}

	return nil
}

func (wc *WebSocketController) addSubscriber(movieID int, ws *websocket.Conn) {
	var subscribers []*websocket.Conn
	if val, exists := wc.subscribers.Load(movieID); exists {
		subscribers = val.([]*websocket.Conn)
	} else {
		subscribers = make([]*websocket.Conn, 0)
	}
	subscribers = append(subscribers, ws)
	wc.subscribers.Store(movieID, subscribers)
}

func (wc *WebSocketController) removeSubscriber(movieID int, ws *websocket.Conn) {
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

func (wc *WebSocketController) UpdateStreamState(movieID int, state map[string]interface{}) {
	wc.streamStates.Store(movieID, state)

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
