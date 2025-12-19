package websocket

import (
	"encoding/json"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// Hub はWebSocket接続を管理
type Hub struct {
	// transactionID -> clients のマッピング
	rooms map[int64]map[*Client]bool

	// クライアント登録リクエスト
	register chan *Client

	// クライアント登録解除リクエスト
	unregister chan *Client

	// メッセージブロードキャスト
	broadcast chan *BroadcastMessage
}

// BroadcastMessage はブロードキャストするメッセージ
type BroadcastMessage struct {
	TransactionID int64
	Message       *entity.Message
}

// NewHub は新しいHubを作成
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

// Run はHubを起動
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// ルーム（取引ID）が存在しない場合は作成
			if h.rooms[client.transactionID] == nil {
				h.rooms[client.transactionID] = make(map[*Client]bool)
			}
			h.rooms[client.transactionID][client] = true

		case client := <-h.unregister:
			if clients, ok := h.rooms[client.transactionID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)

					// ルームが空になったら削除
					if len(clients) == 0 {
						delete(h.rooms, client.transactionID)
					}
				}
			}

		case message := <-h.broadcast:
			// 同じ取引IDのクライアントにメッセージをブロードキャスト
			if clients, ok := h.rooms[message.TransactionID]; ok {
				messageBytes, _ := json.Marshal(message.Message)
				for client := range clients {
					select {
					case client.send <- messageBytes:
					default:
						// 送信できない場合はクライアントを閉じる
						close(client.send)
						delete(clients, client)
					}
				}
			}
		}
	}
}

// Broadcast はメッセージをブロードキャスト
func (h *Hub) Broadcast(transactionID int64, message *entity.Message) {
	h.broadcast <- &BroadcastMessage{
		TransactionID: transactionID,
		Message:       message,
	}
}
