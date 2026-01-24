package zabbix

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Get provides functionality to get data from Zabbix Agent using Zabbix Get Protocol
type Get struct {
	Host    string        // Zabbix Agent address (host:port)
	Port    int           // Zabbix Agent port (default: 10050)
	Timeout time.Duration // Connection timeout (default: 5 seconds)
	Logger  *log.Logger // Logger for debugging
}

// NewGet creates a new Get instance
func NewGet(host string, port int) *Get {
	if port == 0 {
		port = 10050 // Default Zabbix Agent port
	}
	return &Get{
		Host:    host,
		Port:    port,
		Timeout: 5 * time.Second,
	}
}

// SetTimeout sets the connection timeout
func (g *Get) SetTimeout(timeout time.Duration) {
	g.Timeout = timeout
}

func (g *Get) printf(format string, v ...interface{}) {
	if g.Logger != nil {
		g.Logger.Printf(format, v...)
	}
}

// GetValue retrieves a value from Zabbix Agent by key
// Returns the value as a string, or an error if the request fails
func (g *Get) GetValue(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	// Connect to Zabbix Agent
	address := net.JoinHostPort(g.Host, fmt.Sprintf("%d", g.Port))
	conn, err := net.DialTimeout("tcp", address, g.Timeout)
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer conn.Close()

	// Set timeout
	if err := conn.SetDeadline(time.Now().Add(g.Timeout)); err != nil {
		return "", fmt.Errorf("failed to set deadline: %w", err)
	}

	g.printf("Requesting key: %s", key)

	// Send request: "key\n"
	request := key + "\n"
	_, err = conn.Write([]byte(request))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Remove trailing newline
	response = strings.TrimRight(response, "\n\r")

	g.printf("Received value: %s", response)

	// Check for Zabbix error responses
	if strings.HasPrefix(response, "ZBX_NOTSUPPORTED") {
		return "", fmt.Errorf("key not supported: %s", key)
	}
	if strings.HasPrefix(response, "ZBX_ERROR") {
		return "", fmt.Errorf("agent error: %s", response)
	}

	return response, nil
}

// GetValues retrieves multiple values from Zabbix Agent
// Returns a map of key-value pairs, or an error if the request fails
func (g *Get) GetValues(keys []string) (map[string]string, error) {
	result := make(map[string]string)
	var errors []string

	for _, key := range keys {
		value, err := g.GetValue(key)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", key, err))
			continue
		}
		result[key] = value
	}

	if len(errors) > 0 {
		return result, fmt.Errorf("some keys failed: %s", strings.Join(errors, "; "))
	}

	return result, nil
}
