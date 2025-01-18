package main

import (
	"log"
	"net"
	"sync"
	"time"
)

// Connection represents an active client connection
type Connection struct {
	ID        string
	Addr      string
	Connected time.Time
}

// ConnectionTracker keeps track of all active connections
type ConnectionTracker struct {
	connections map[string]Connection
	mu          sync.RWMutex
}

// NewConnectionTracker creates a new connection tracker
func NewConnectionTracker() *ConnectionTracker {
	return &ConnectionTracker{
		connections: make(map[string]Connection),
	}
}

// Add adds a new connection to the tracker
func (ct *ConnectionTracker) Add(conn net.Conn) Connection {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	c := Connection{
		ID:        conn.RemoteAddr().String(),
		Addr:      conn.RemoteAddr().String(),
		Connected: time.Now(),
	}

	ct.connections[c.ID] = c
	return c
}

// Remove removes a connection from the tracker
func (ct *ConnectionTracker) Remove(id string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	delete(ct.connections, id)
}

// GetActiveConnections returns all active connections
func (ct *ConnectionTracker) GetActiveConnections() []Connection {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	active := make([]Connection, 0, len(ct.connections))
	for _, conn := range ct.connections {
		active = append(active, conn)
	}
	return active
}

func main() {
	// Initialize the connection tracker
	tracker := NewConnectionTracker()

	// Create a custom logger with timestamp
	logger := log.New(log.Writer(), "", log.Ldate|log.Ltime)

	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		logger.Fatalf("Error listening on port 42069: %s", err)
	}
	defer ln.Close()

	logger.Printf("Server started, listening on port 42069")

	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Printf("Error accepting connection: %s", err)
			continue
		}

		// Add connection to tracker and log it
		c := tracker.Add(conn)
		logger.Printf("New connection from %s (total active: %d)",
			c.Addr,
			len(tracker.GetActiveConnections()),
		)

		// Handle connection in a goroutine
		go handleConnection(conn, c.ID, tracker, logger)
	}
}

func handleConnection(conn net.Conn, id string, tracker *ConnectionTracker, logger *log.Logger) {
	defer func() {
		conn.Close()
		tracker.Remove(id)
		logger.Printf("Connection closed from %s (total active: %d)",
			conn.RemoteAddr(),
			len(tracker.GetActiveConnections()),
		)
	}()

	// Get connection details
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	logger.Printf("Connection details for %s:", id)
	logger.Printf("  - IP Address: %s", remoteAddr.IP)
	logger.Printf("  - Port: %d", remoteAddr.Port)
	logger.Printf("  - Network Type: %s", remoteAddr.Network())

	// Handle the connection (example implementation)
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err.Error() != "EOF" {
				logger.Printf("Error reading from connection %s: %s", id, err)
			}
			return
		}

		// Log received data
		logger.Printf("Received %d bytes from %s: %s", n, id, string(buffer[:n]))
		logger.Printf("Received %d bytes from %s: %s", n, id, buffer[:n])

		// Echo the data back (example response)
		_, err = conn.Write(buffer[:n])
		if err != nil {
			logger.Printf("Error writing to connection %s: %s", id, err)
			return
		}
	}
}
