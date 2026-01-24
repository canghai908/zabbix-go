package zabbix

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// SenderData represents a single data item to send to Zabbix Server
type SenderData struct {
	Host  string `json:"host"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Clock int64  `json:"clock,omitempty"` // Unix timestamp, 0 means current time
}

// SenderResponse represents the response from Zabbix Server
type SenderResponse struct {
	Response string `json:"response"`
	Info     string `json:"info"`
}

// Sender provides functionality to send data to Zabbix Server using Zabbix Sender Protocol
type Sender struct {
	Server  string        // Zabbix Server address (host:port)
	Port    int           // Zabbix Server port (default: 10051)
	Timeout time.Duration // Connection timeout (default: 5 seconds)
	Logger  *log.Logger   // Logger for debugging
}

// NewSender creates a new Sender instance
func NewSender(server string, port int) *Sender {
	if port == 0 {
		port = 10051 // Default Zabbix Server port
	}
	return &Sender{
		Server:  server,
		Port:    port,
		Timeout: 5 * time.Second,
	}
}

// SetTimeout sets the connection timeout
func (s *Sender) SetTimeout(timeout time.Duration) {
	s.Timeout = timeout
}

func (s *Sender) printf(format string, v ...interface{}) {
	if s.Logger != nil {
		s.Logger.Printf(format, v...)
	}
}

// Send sends a single data item to Zabbix Server
func (s *Sender) Send(data SenderData) (*SenderResponse, error) {
	return s.SendBatch([]SenderData{data})
}

// SendBatch sends multiple data items to Zabbix Server in a single request
func (s *Sender) SendBatch(data []SenderData) (*SenderResponse, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to send")
	}

	// Set clock to current time if not set
	now := time.Now().Unix()
	for i := range data {
		if data[i].Clock == 0 {
			data[i].Clock = now
		}
	}

	// Marshal JSON data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	s.printf("Sending data: %s", string(jsonData))

	// Build ZBXD protocol packet
	// Format: "ZBXD\1" + 8 bytes (data length) + JSON data
	packet := s.BuildPacket(jsonData)

	// Connect to Zabbix Server
	address := net.JoinHostPort(s.Server, fmt.Sprintf("%d", s.Port))
	conn, err := net.DialTimeout("tcp", address, s.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer conn.Close()

	// Set write timeout
	if err := conn.SetWriteDeadline(time.Now().Add(s.Timeout)); err != nil {
		return nil, fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Send packet
	_, err = conn.Write(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to send data: %w", err)
	}

	// Set read timeout
	if err := conn.SetReadDeadline(time.Now().Add(s.Timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read response header
	header := make([]byte, 13) // "ZBXD\1" + 8 bytes length
	_, err = io.ReadFull(conn, header)
	if err != nil {
		return nil, fmt.Errorf("failed to read response header: %w", err)
	}

	// Verify ZBXD marker
	if string(header[0:5]) != "ZBXD\x01" {
		return nil, fmt.Errorf("invalid response header: expected ZBXD\\x01")
	}

	// Read data length
	var dataLen uint64
	if err := binary.Read(bytes.NewReader(header[5:13]), binary.LittleEndian, &dataLen); err != nil {
		return nil, fmt.Errorf("failed to read data length: %w", err)
	}

	if dataLen == 0 {
		return nil, fmt.Errorf("empty response from server")
	}

	// Read response data
	responseData := make([]byte, dataLen)
	_, err = io.ReadFull(conn, responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to read response data: %w", err)
	}

	s.printf("Received response: %s", string(responseData))

	// Parse response
	var response SenderResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// BuildPacket builds a ZBXD protocol packet
// Format: "ZBXD\1" + 8 bytes (little-endian data length) + JSON data
func (s *Sender) BuildPacket(data []byte) []byte {
	header := []byte("ZBXD\x01")
	length := make([]byte, 8)
	binary.LittleEndian.PutUint64(length, uint64(len(data)))

	packet := make([]byte, 0, len(header)+len(length)+len(data))
	packet = append(packet, header...)
	packet = append(packet, length...)
	packet = append(packet, data...)

	return packet
}
