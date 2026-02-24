package stream

import (
	"context"
	"io"
	"net/http"

	"github.com/coder/websocket"
)

// Conn represents a WebSocket connection with read/write loops
type Conn struct {
	conn      *websocket.Conn
	readChan  chan []byte
	writeChan chan []byte
	closeChan chan struct{}
}

// NewConn creates a new WebSocket connection wrapper
func NewConn() *Conn {
	return &Conn{
		readChan:  make(chan []byte, 100),
		writeChan: make(chan []byte, 100),
		closeChan: make(chan struct{}),
	}
}

// Accept performs the WebSocket handshake and starts read/write loops
func (c *Conn) Accept(w http.ResponseWriter, r *http.Request) error {
	opts := &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	}

	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		return err
	}

	c.conn = conn

	// Start read loop
	go c.readLoop()

	// Start write loop
	go c.writeLoop()

	return nil
}

// readLoop runs in a goroutine to continuously read messages from the WebSocket
func (c *Conn) readLoop() {
	defer close(c.readChan)

	ctx := context.Background()
	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			if err != io.EOF {
				// Log error if needed
			}
			return
		}

		select {
		case c.readChan <- data:
		case <-c.closeChan:
			return
		}
	}
}

// writeLoop runs in a goroutine to continuously write messages to the WebSocket
func (c *Conn) writeLoop() {
	defer close(c.writeChan)

	ctx := context.Background()
	for {
		select {
		case data, ok := <-c.writeChan:
			if !ok {
				return
			}
			err := c.conn.Write(ctx, websocket.MessageText, data)
			if err != nil {
				return
			}
		case <-c.closeChan:
			return
		}
	}
}

// Read returns a channel for incoming messages
func (c *Conn) Read() <-chan []byte {
	return c.readChan
}

// Write sends a message through the write channel
func (c *Conn) Write(data []byte) error {
	select {
	case c.writeChan <- data:
		return nil
	case <-c.closeChan:
		return io.ErrClosedPipe
	}
}

// Close gracefully closes the WebSocket connection
func (c *Conn) Close() error {
	close(c.closeChan)
	if c.conn != nil {
		return c.conn.Close(websocket.StatusNormalClosure, "")
	}
	return nil
}
