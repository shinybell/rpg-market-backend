package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client はWebSocket接続を表す
type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	userID        int64
	transactionID int64
}

// IncomingMessage はクライアントから受信するメッセージ
type IncomingMessage struct {
	Content string `json:"content"`
}

// NewClient は新しいClientを作成
func NewClient(hub *Hub, conn *websocket.Conn, userID, transactionID int64) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		userID:        userID,
		transactionID: transactionID,
	}
}

// readPump はWebSocketからメッセージを読み取る
func (c *Client) readPump(onMessage func(content string)) {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(message, &incoming); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}

		// メッセージ処理コールバック
		onMessage(incoming.Content)
	}
}

// writePump はWebSocketへメッセージを書き込む
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hubがチャネルを閉じた
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// キューに溜まっているメッセージを追加
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Start はクライアントの読み書きを開始
func (c *Client) Start(onMessage func(content string)) {
	// クライアントをHubに登録
	c.hub.register <- c

	go c.writePump()
	go c.readPump(onMessage)
}
