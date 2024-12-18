package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	wsClientsPool = NewWSClientsPool()
)

type WSEventType string

const (
	writeWait = 10 * time.Second // Time allowed to write a message to the peer.
	pongWait  = 60 * time.Second // Time allowed to read the next pong message from the peer.
	// pongWait    = time.Hour // Time allowed to read the next pong message from the peer.
	pingMessage = "ping"
	pongMessage = "pong"

	WSMessageEvent WSEventType = "message"
	WSNewRoomEvent WSEventType = "new_room"
)

type WSClientsPool struct {
	connMutex sync.RWMutex
	upgrader  websocket.Upgrader
	clients   map[string][]*WSClientSocket
	nextID    uint64
	closeC    chan struct{}
}

func NewWSClientsPool() *WSClientsPool {
	return &WSClientsPool{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[string][]*WSClientSocket),
		closeC:  make(chan struct{}, 1),
	}
}

func (p *WSClientsPool) generateUniqueID() uint64 { return atomic.AddUint64(&p.nextID, 1) }

// Add adds a new client to the pool, not thread-safe
func (p *WSClientsPool) Add(key string, clientConnection *WSClientSocket) {
	clientConnection.id = p.generateUniqueID()
	if _, exists := p.clients[key]; !exists {
		p.clients[key] = make([]*WSClientSocket, 0)
	}
	p.clients[key] = append(p.clients[key], clientConnection)
}

// Remove removes a client from the pool, not thread-safe
func (p *WSClientsPool) Remove(key string, clientConnectionID uint64) {
	clients, exists := p.clients[key]
	if !exists {
		return
	}
	for i, client := range clients {
		if client.id == clientConnectionID {
			p.clients[key] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	// Remove key if no clients remain
	if len(p.clients[key]) == 0 {
		delete(p.clients, key)
	}
}

func (p *WSClientsPool) readPump(clientConn *WSClientSocket) {
	conn := clientConn.connection

	logger := AppLogger.WithField("id", clientConn.id).WithField("user_id", clientConn.userId)

	conn.SetPingHandler(func(appData string) error {
		clientConn.isWebBrowser = false
		// logger.Debug("received ping message")
		clientConn.pingMessage <- []byte(appData)
		err := conn.SetReadDeadline(time.Now().Add(pongWait))
		return err
	})
	conn.SetReadDeadline(time.Now().Add(pongWait))

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.WithError(err).Error("read failed unexpectedly")
			} else {
				logger.WithError(err).Debugf("handling read error")
			}
			// Notify writePump of error. Force close will be handled there
			clientConn.forceCloseC <- err
			return
		}
		conn.SetReadDeadline(time.Now().Add(pongWait))

		if string(message) == pingMessage || string(message) == pongMessage {
			// logger.Debug("received ping message")
			clientConn.isWebBrowser = true
			clientConn.pingMessage <- message
			continue
		}

		if err := ReceiveWSEvent(clientConn.user, message); err != nil {
			logger.WithError(err).Error("failed to process message")
		}

		logger.Debugf("received %d bytes", len(message))

	}
}

func (p *WSClientsPool) writePump(clientConn *WSClientSocket) {
	conn := clientConn.connection

	logger := AppLogger.WithField("id", clientConn.id).WithField("user_id", clientConn.userId)

	for {
		select {
		case message, ok := <-clientConn.outQueue:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				logger.Error("output queue was closed, forcefully closing")
				return
			}
			data, err := message.JSONMarshal()
			if err != nil {
				logger.WithError(err).Errorf("failed to marshal message")
				p.cleanupConnection(clientConn)
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				logger.WithError(err).Errorf("write failed")
				p.cleanupConnection(clientConn)
				return
			}
			logger.Debugf("written %d bytes", len(data))
		case ping := <-clientConn.pingMessage:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			var msgType int
			if clientConn.isWebBrowser {
				msgType = websocket.TextMessage
			} else {
				msgType = websocket.PongMessage
			}
			err := conn.WriteMessage(msgType, ping)
			if err != nil {
				logger.WithError(err).Error("failed to send pong message")
				p.cleanupConnection(clientConn)
				return
			}
			// logger.Debug("sent pong message")
		case closeErr := <-clientConn.closeC:
			logger.Debugf("closing connection")
			if err := conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(closeErr.Code, closeErr.Text),
				time.Now().Add(writeWait),
			); err != nil {
				logger.WithError(err).Error("failed to send close message")
			}
			p.cleanupConnection(clientConn)
			return
		case forceCloseErr, ok := <-clientConn.forceCloseC:
			if !ok || forceCloseErr != nil {
				logger.WithError(forceCloseErr).Error("handling forced close signal")
				p.cleanupConnection(clientConn)
			}
			return
		}
	}
}

func (p *WSClientsPool) cleanupConnection(clientConn *WSClientSocket) {
	logger := AppLogger.WithField("id", clientConn.id).WithField("user_id", clientConn.userId)
	clientConn.connection.Close()
	p.connMutex.Lock()
	close(clientConn.outQueue)
	close(clientConn.closeC)
	p.Remove(clientConn.userId, clientConn.id)
	p.connMutex.Unlock()
	logger.Debugln("connection closed")
}

// close closes all connections in the pool and cleans up
func (p *WSClientsPool) close() {
	// for _, clientConnections := range p.clients {
	// 	for _, conn := range clientConnections {
	// 		p.cleanupConnection(conn)
	// 	}
	// }
	AppLogger.Info("all connections closed")
}

type WSClientSocket struct {
	connection   *websocket.Conn
	id           uint64 // unique id for the connection, assigned by the pool on creation
	user         *User
	userId       string // user id, can not be used as a key
	outQueue     chan WSClientEventMessage
	pingMessage  chan []byte
	isWebBrowser bool
	closeC       chan websocket.CloseError
	forceCloseC  chan error
}

type WSClientEventMessage struct {
	Type      WSEventType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	DataModel interface{}     `json:"-"` // this will be marshalled to json
	Data      json.RawMessage `json:"data"`
}

func (m *WSClientEventMessage) hasData() bool { return m.DataModel != nil && len(m.Data) > 0 }

func (m *WSClientEventMessage) JSONMarshal() ([]byte, error) {
	if m.hasData() {
		return json.Marshal(m)
	}
	if err := m.populateData(); err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func (m *WSClientEventMessage) populateData() (err error) {
	m.Data, err = json.Marshal(m.DataModel)
	return
}

func HandleChatWS(w http.ResponseWriter, r *http.Request) {
	conn, err := wsClientsPool.upgrader.Upgrade(w, r, nil)
	if err != nil {
		AppLogger.WithError(err).
			WithField("remote_addr", r.RemoteAddr).
			WithField("url_path", r.URL.Path).
			Error("failed to upgrade ws connection")
		return
	}

	user, ok := r.Context().Value("user").(*User)
	if !ok {
		AppLogger.Error("user not found in chat ws context")
		conn.Close()
		return
	}

	client := &WSClientSocket{
		connection:  conn,
		user:        user,
		userId:      user.ID.String(),
		outQueue:    make(chan WSClientEventMessage, 1),
		closeC:      make(chan websocket.CloseError, 1),
		pingMessage: make(chan []byte, 1),
		forceCloseC: make(chan error, 1),
	}

	wsClientsPool.connMutex.Lock()
	wsClientsPool.Add(user.ID.String(), client)
	wsClientsPool.connMutex.Unlock()

	go wsClientsPool.readPump(client)
	go wsClientsPool.writePump(client)

	AppLogger.WithField("id", client.id).WithField("user_id", user.ID).Info("new ws connection")
}

func BroadcastWSMassage(usersIDs []string, transformer func(userId string) WSClientEventMessage) {
	for _, userID := range usersIDs {
		func() {
			wsClientsPool.connMutex.RLock()
			defer wsClientsPool.connMutex.RUnlock()

			clients, exists := wsClientsPool.clients[userID]
			if !exists {
				AppLogger.WithField("user_id", userID).Warn("no clients found for user")
				return
			}
			for _, client := range clients {
				logger := AppLogger.WithField("id", client.id).WithField("user_id", client.userId)
				select {
				case client.outQueue <- transformer(userID):
				default:
					logger.Error("failed to send message to client")
				}
			}
		}()
	}
}

func BroadcastMessageToWSClients(usersIDs []string, message WSClientEventMessage) {
	for _, userID := range usersIDs {
		if err := SendMessageToWSClient(userID, message); err != nil {
			AppLogger.WithError(err).Errorf("failed to send message to user %s", userID)
		}
	}
}

func SendMessageToWSClient(userID string, message WSClientEventMessage) error {
	wsClientsPool.connMutex.RLock()
	defer wsClientsPool.connMutex.RUnlock()
	clients, exists := wsClientsPool.clients[userID]
	if !exists {
		return fmt.Errorf("[websocket] no clients found for user %s", userID)
	}
	for _, client := range clients {
		logger := AppLogger.WithField("id", client.id).WithField("user_id", client.userId)
		select {
		case client.outQueue <- message:
			logger.Debug("message sent to client")
		default:
			logger.Error("failed to send message to client")
		}
	}
	return nil
}
