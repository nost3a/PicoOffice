package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/nost3a/PicoOffice/internal/middleware"
	"github.com/nost3a/PicoOffice/internal/service"
	"github.com/nost3a/PicoOffice/internal/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSDocs  /api/ws/docs/:id?token=xxx
func WSDocs(c *gin.Context) {
	tokStr := c.Query("token")
	if tokStr == "" {
		c.JSON(401, gin.H{"error": "missing token"})
		return
	}
	claims, err := middleware.ParseToken(tokStr)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid token"})
		return
	}
	uid := uint(claims["uid"].(float64))
	username := claims["username"].(string)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	cli := &ws.Client{
		Conn:     conn,
		UserID:   uid,
		Username: username,
		RoomID:   "doc-" + c.Param("id"),
		Send:     make(chan []byte, 16),
		LastSeen: time.Now().UnixNano(),
	}
	wsHub.Register <- cli

	go writePump(cli)
	readPump(cli)
}

// writePump sends outbound; every 30s pushes {"type":"ping"},
// check every 10s; close connection after 60s idle
func writePump(cli *ws.Client) {
	defer cli.Conn.Close()
	ping := time.NewTicker(30 * time.Second)
	watch := time.NewTicker(10 * time.Second)
	defer ping.Stop()
	defer watch.Stop()
	for {
		select {
		case msg, ok := <-cli.Send:
			if !ok {
				return
			}
			if err := cli.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ping.C:
			p, _ := json.Marshal(map[string]string{"type": "ping"})
			if err := cli.Conn.WriteMessage(websocket.TextMessage, p); err != nil {
				return
			}
		case <-watch.C:
			if time.Now().UnixNano()-atomic.LoadInt64(&cli.LastSeen) > int64(60*time.Second) {
				return
			}
		}
	}
}

// readPump broadcasts to room on message; refreshes activity
func readPump(cli *ws.Client) {
	defer func() {
		wsHub.Unregister <- cli
		cli.Conn.Close()
	}()
	for {
		_, raw, err := cli.Conn.ReadMessage()
		if err != nil {
			break
		}
		atomic.StoreInt64(&cli.LastSeen, time.Now().UnixNano())
		var body map[string]interface{}
		if err := json.Unmarshal(raw, &body); err != nil {
			continue
		}

		// OT: type=steps goes through step ingest/broadcast path
		if body["type"] == "steps" {
			handleStepsMsg(cli, raw)
			continue
		}

		// pong/cursor/op all go through the same broadcast path
		body["user"] = cli.Username
		body["user_id"] = cli.UserID
		out, _ := json.Marshal(body)
		wsHub.Broadcast(cli.RoomID, out, cli)
	}
}

// stepsIn OT steps payload from client
type stepsIn struct {
	Type          string            `json:"type"`
	ClientVersion int               `json:"client_version"`
	Steps         []json.RawMessage `json:"steps"`
}

// handleStepsMsg: ingest step -> broadcast to room; nack on stale
func handleStepsMsg(cli *ws.Client, raw []byte) {
	var in stepsIn
	if err := json.Unmarshal(raw, &in); err != nil || in.Type != "steps" {
		return
	}
	// RoomID = "doc-{id}"
	docID, _ := strconv.Atoi(strings.TrimPrefix(cli.RoomID, "doc-"))
	if docID <= 0 {
		return
	}
	res, err := service.ReceiveSteps(db, uint(docID), in.ClientVersion, in.Steps, cli.UserID)
	if err != nil {
		// abnormal (client ahead etc.), return error to sender
		bad, _ := json.Marshal(map[string]interface{}{
			"type":  "steps_error",
			"error": err.Error(),
		})
		select {
		case cli.Send <- bad:
		default:
		}
		return
	}

	if !res.Accepted {
		// client stale: return steps in range for local transform, no broadcast
		conflict, _ := json.Marshal(map[string]interface{}{
			"type":         "steps_conflict",
			"server_error": true,
			"version":      res.Version,
			"steps":        res.Steps,
		})
		select {
		case cli.Send <- conflict:
		default:
		}
		return
	}

	// step accepted: broadcast to room (sender already applied locally)
	broadcast, _ := json.Marshal(map[string]interface{}{
		"type":    "steps",
		"version": res.Version,
		"steps":   in.Steps,
		"user":    cli.Username,
		"user_id": cli.UserID,
	})
	wsHub.Broadcast(cli.RoomID, broadcast, cli)
	// ack to sender with new version
	ack, _ := json.Marshal(map[string]interface{}{
		"type":    "steps_ack",
		"version": res.Version,
	})
	select {
	case cli.Send <- ack:
	default:
	}
}
