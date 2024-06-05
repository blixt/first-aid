package chromecontrol

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type session struct {
	conn    *websocket.Conn
	pending map[int]chan rpcResult
	mu      sync.Mutex
}

func (s *session) clearPendingRPCs() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, ch := range s.pending {
		close(ch)
		delete(s.pending, id)
	}
}

func (s *session) handleMessages() {
	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			break
		}

		var result rpcResult
		if err := json.Unmarshal(message, &result); err != nil {
			fmt.Println("Unmarshal error:", err)
			continue
		}

		s.mu.Lock()
		if ch, ok := s.pending[result.ID]; ok {
			select {
			case ch <- result:
			default:
				fmt.Println("Warning: result channel is full, discarding result")
			}
			delete(s.pending, result.ID)
		}
		s.mu.Unlock()
	}

	s.clearPendingRPCs()
}
