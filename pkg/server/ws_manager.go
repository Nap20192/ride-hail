package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"ride-hail/internal/middleware"
	"ride-hail/pkg/uuid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Manager struct {
	ctx          context.Context
	clients      map[uuid.UUID]*Client
	read         chan RequestWs
	write        chan ResponseWs
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	shutdownOnce sync.Once
	mu           sync.Mutex
}

func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		clients: make(map[uuid.UUID]*Client),
		wg:      sync.WaitGroup{},
		read:    make(chan RequestWs),
		write:   make(chan ResponseWs),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (m *Manager) addClient(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[c.id] = c
}

func (m *Manager) removeClient(c uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, c)
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to upgrade connection"))
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		return
	}
	slog.Info("WebSocket connection established", "client_id", claims.UserID)

	client := NewClient(claims.UserID, conn, m)
	m.addClient(client)
	m.wg.Add(3)
	go func() {
		defer m.wg.Done()
		client.readMessages()
	}()
	go func() {
		defer m.wg.Done()
		client.writeMessages()
	}()
	go func() {
		defer m.wg.Done()
		for message := range client.inbound {
			m.read <- RequestWs{
				Payload:    message,
				ProducerID: client.id,
			}
		}
	}()
}

type ResponseWs struct {
	Payload    []byte
	ConsumerID uuid.UUID
}

type RequestWs struct {
	Payload    []byte
	ProducerID uuid.UUID
}

func (m *Manager) StartWrite(ctx context.Context) {
	m.mu.Lock()
	if ctx == nil {
		ctx = context.Background()
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case <-m.ctx.Done():
				return
			case message, ok := <-m.write:
				if !ok {
					return
				}
				m.mu.Lock()
				client, exists := m.clients[message.ConsumerID]
				m.mu.Unlock()
				if exists {
					select {
					case client.outbound <- message.Payload:
					default:
						slog.Warn("Client outbound channel full", "client_id", message.ConsumerID)
					}
				} else {
					slog.Warn("Client not found for message", "client_id", message.ConsumerID)
				}
			}
		}
	}()
}

func (m *Manager) ReadChannel() <-chan RequestWs {
	return m.read
}

func (m *Manager) WriteChannel() chan<- ResponseWs {
	return m.write
}

func (m *Manager) Shutdown() {
	m.shutdownOnce.Do(func() {
		if m.cancel != nil {
			m.cancel()
		}

		m.mu.Lock()
		for _, client := range m.clients {
			client.close()
		}
		m.mu.Unlock()

		m.wg.Wait()

		close(m.read)
	})
}
