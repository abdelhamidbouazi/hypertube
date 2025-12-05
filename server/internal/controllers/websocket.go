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
	subscribers  map[int][]*websocket.Conn      // movieID -> list of websocket connections
	streamStates map[int]map[string]interface{} // movieID -> last stream state
	mu           sync.RWMutex                   // protects subscribers and streamStates
}

func NewWebSocketController() *WebSocketController {
	return &WebSocketController{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for testing (restrict this in production)
				return true
			},
		},
		subscribers:  make(map[int][]*websocket.Conn),
		streamStates: make(map[int]map[string]interface{}),
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

	wc.mu.RLock()
	if lastState, exists := wc.streamStates[movieID]; exists {
		if stateJSON, err := json.Marshal(lastState); err == nil {
			ws.WriteMessage(websocket.TextMessage, stateJSON)
		}
	}
	wc.mu.RUnlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}

	return nil
}

func (wc *WebSocketController) addSubscriber(movieID int, ws *websocket.Conn) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if wc.subscribers[movieID] == nil {
		wc.subscribers[movieID] = make([]*websocket.Conn, 0)
	}
	wc.subscribers[movieID] = append(wc.subscribers[movieID], ws)
}

func (wc *WebSocketController) removeSubscriber(movieID int, ws *websocket.Conn) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if subscribers, exists := wc.subscribers[movieID]; exists {
		for i, conn := range subscribers {
			if conn == ws {
				wc.subscribers[movieID] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}

		if len(wc.subscribers[movieID]) == 0 {
			delete(wc.subscribers, movieID)
		}
	}
}

func (wc *WebSocketController) UpdateStreamState(movieID int, state map[string]interface{}) {
	wc.mu.Lock()
	wc.streamStates[movieID] = state

	subscribers := wc.subscribers[movieID]
	wc.mu.Unlock()

	if len(subscribers) > 0 {
		stateJSON, err := json.Marshal(state)
		if err != nil {
			return
		}

		wc.mu.RLock()
		for _, ws := range subscribers {
			ws.WriteMessage(websocket.TextMessage, stateJSON)
		}
		wc.mu.RUnlock()
	}
}
