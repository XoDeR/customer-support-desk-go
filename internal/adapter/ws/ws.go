package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const Channel = "support.events"

type Hub struct {
	mu       sync.RWMutex
	clients  map[*websocket.Conn]entity.Role
	upgrader websocket.Upgrader
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]entity.Role), upgrader: websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, role entity.Role) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.mu.Lock()
	h.clients[conn] = role
	h.mu.Unlock()
	defer func() { h.mu.Lock(); delete(h.clients, conn); h.mu.Unlock(); _ = conn.Close() }()
	for {
		if _, _, err = conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Hub) Broadcast(event entity.DomainEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn, role := range h.clients {
		if role == entity.RoleCustomer && event.Type == "comment.created" {
			if p, ok := event.Payload.(map[string]any); ok && p["visibility"] == "internal" {
				continue
			}
		}
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			_ = conn.Close()
		}
	}
}

type Publisher struct{ client *redis.Client }

func NewPublisher(client *redis.Client) *Publisher { return &Publisher{client} }
func (p *Publisher) Publish(ctx context.Context, event entity.DomainEvent) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, Channel, b).Err()
}
func (p *Publisher) EnqueueEscalation(ctx context.Context, event entity.DomainEvent) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.client.LPush(ctx, "support.escalations", b).Err()
}
func Subscribe(ctx context.Context, client *redis.Client, hub *Hub) {
	pubsub := client.Subscribe(ctx, Channel)
	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var event entity.DomainEvent
				if json.Unmarshal([]byte(msg.Payload), &event) == nil {
					hub.Broadcast(event)
				}
			}
		}
	}()
}
