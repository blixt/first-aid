package chromecontrol

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	server   *http.Server
	mu       sync.Mutex
	cond     *sync.Cond
	rpcID    int
	sessions []*session
}

type rpcCall struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
}

type rpcResult struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

func NewServer() *Server {
	s := &Server{
		server:   &http.Server{Addr: "localhost:49158"},
		sessions: []*session{},
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *Server) Start() error {
	http.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		newSession := &session{
			conn:    conn,
			pending: make(map[int]chan rpcResult),
		}

		s.mu.Lock()
		s.sessions = append(s.sessions, newSession)
		s.cond.Broadcast()
		s.mu.Unlock()

		go func() {
			newSession.handleMessages()
			s.mu.Lock()
			for i, sess := range s.sessions {
				if sess == newSession {
					s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
					break
				}
			}
			s.mu.Unlock()
		}()
	})

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("ListenAndServe error: %v", err))
		}
	}()

	return nil
}

func (s *Server) sendRPC(method string, params any) (json.RawMessage, error) {
	s.mu.Lock()
	for len(s.sessions) == 0 {
		s.cond.Wait()
	}
	s.rpcID++
	id := s.rpcID
	latestSession := s.sessions[len(s.sessions)-1]
	s.mu.Unlock()

	ch := make(chan rpcResult, 1)

	latestSession.mu.Lock()
	latestSession.pending[id] = ch
	latestSession.mu.Unlock()

	call := rpcCall{
		ID:     id,
		Method: method,
		Params: params,
	}

	message, err := json.Marshal(call)
	if err != nil {
		return nil, err
	}

	if err := latestSession.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		return nil, err
	}

	select {
	case result, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("RPC call failed: channel closed")
		}
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		return result.Result, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("RPC call timed out")
	}
}

func (s *Server) Close() error {
	return s.server.Close()
}
