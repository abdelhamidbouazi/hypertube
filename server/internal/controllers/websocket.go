package controllers

import (
	"encoding/json"
	"net/http"
	"server/internal/services"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WebSocketController struct {
	upgrader         websocket.Upgrader
	websocketService *services.WebSocketService
}

func NewWebSocketController(ws *services.WebSocketService) *WebSocketController {
	return &WebSocketController{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for testing (restrict this in production)
				return true
			},
		},
		websocketService: ws,
	}
}

// HandleWebSocket handles WebSocket connections for a specific movie
//
//	@Summary		WebSocket endpoint for movie streaming updates
//	@Description	Establishes a WebSocket connection to receive real-time streaming updates for a specific movie
//	@Tags			WebSocket
//	@Accept			json
//	@Produce		json
//	@Param			movieId	path		int		true	"Movie ID"
//	@Success		101		{string}	string	"Switching Protocols"
//	@Router			/ws/{movieId} [get]
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
	defer wc.websocketService.RemoveSubscriber(movieID, ws)

	wc.websocketService.AddSubscriber(movieID, ws)

	if lastState, exists := wc.websocketService.StreamStates.Load(movieID); exists {
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
