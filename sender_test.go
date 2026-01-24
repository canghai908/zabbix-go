package zabbix_test

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/canghai908/zabbix-go"
)

func TestSenderBuildPacket(t *testing.T) {
	sender := NewSender("localhost", 10051)
	data := []byte(`{"test":"data"}`)
	packet := sender.BuildPacket(data)

	// Check ZBXD header
	if string(packet[0:5]) != "ZBXD\x01" {
		t.Errorf("Expected ZBXD\\x01 header, got %q", packet[0:5])
	}

	// Check data length (8 bytes, little-endian)
	expectedLen := uint64(len(data))
	actualLen := uint64(packet[5]) | uint64(packet[6])<<8 | uint64(packet[7])<<16 | uint64(packet[8])<<24 |
		uint64(packet[9])<<32 | uint64(packet[10])<<40 | uint64(packet[11])<<48 | uint64(packet[12])<<56

	if actualLen != expectedLen {
		t.Errorf("Expected data length %d, got %d", expectedLen, actualLen)
	}

	// Check data
	if string(packet[13:]) != string(data) {
		t.Errorf("Expected data %q, got %q", string(data), string(packet[13:]))
	}
}

func TestSenderDataMarshal(t *testing.T) {
	data := SenderData{
		Host:  "test-host",
		Key:   "test.key",
		Value: "123",
		Clock: time.Now().Unix(),
	}

	jsonData, err := json.Marshal([]SenderData{data})
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	var unmarshaled []SenderData
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	if len(unmarshaled) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(unmarshaled))
	}

	if unmarshaled[0].Host != data.Host {
		t.Errorf("Expected host %q, got %q", data.Host, unmarshaled[0].Host)
	}
	if unmarshaled[0].Key != data.Key {
		t.Errorf("Expected key %q, got %q", data.Key, unmarshaled[0].Key)
	}
	if unmarshaled[0].Value != data.Value {
		t.Errorf("Expected value %q, got %q", data.Value, unmarshaled[0].Value)
	}
}

func TestSenderNewSender(t *testing.T) {
	sender := NewSender("localhost", 10051)
	if sender.Server != "localhost" {
		t.Errorf("Expected server localhost, got %s", sender.Server)
	}
	if sender.Port != 10051 {
		t.Errorf("Expected port 10051, got %d", sender.Port)
	}
	if sender.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", sender.Timeout)
	}
}

func TestSenderNewSenderDefaultPort(t *testing.T) {
	sender := NewSender("localhost", 0)
	if sender.Port != 10051 {
		t.Errorf("Expected default port 10051, got %d", sender.Port)
	}
}

func TestSenderSetTimeout(t *testing.T) {
	sender := NewSender("localhost", 10051)
	newTimeout := 10 * time.Second
	sender.SetTimeout(newTimeout)
	if sender.Timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, sender.Timeout)
	}
}

func TestSenderSendBatchEmpty(t *testing.T) {
	sender := NewSender("localhost", 10051)
	_, err := sender.SendBatch([]SenderData{})
	if err == nil {
		t.Error("Expected error for empty batch, got nil")
	}
}

// Note: Integration tests require a running Zabbix Server
// Uncomment and set TEST_ZABBIX_SERVER environment variable to run
/*
func TestSenderSendIntegration(t *testing.T) {
	server := os.Getenv("TEST_ZABBIX_SERVER")
	if server == "" {
		t.Skip("Set TEST_ZABBIX_SERVER environment variable to run integration test")
	}

	sender := NewSender(server, 10051)
	sender.SetTimeout(10 * time.Second)

	data := SenderData{
		Host:  "test-host",
		Key:   "test.key",
		Value: "123",
	}

	response, err := sender.Send(data)
	if err != nil {
		t.Fatalf("Failed to send data: %v", err)
	}

	if response.Response != "success" {
		t.Errorf("Expected success response, got %q", response.Response)
	}
}
*/
