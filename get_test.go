package zabbix_test

import (
	"testing"
	"time"

	. "github.com/canghai908/zabbix-go"
)

func TestGetNewGet(t *testing.T) {
	get := NewGet("localhost", 10050)
	if get.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", get.Host)
	}
	if get.Port != 10050 {
		t.Errorf("Expected port 10050, got %d", get.Port)
	}
	if get.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", get.Timeout)
	}
}

func TestGetNewGetDefaultPort(t *testing.T) {
	get := NewGet("localhost", 0)
	if get.Port != 10050 {
		t.Errorf("Expected default port 10050, got %d", get.Port)
	}
}

func TestGetSetTimeout(t *testing.T) {
	get := NewGet("localhost", 10050)
	newTimeout := 10 * time.Second
	get.SetTimeout(newTimeout)
	if get.Timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, get.Timeout)
	}
}

func TestGetGetValueEmptyKey(t *testing.T) {
	get := NewGet("localhost", 10050)
	_, err := get.GetValue("")
	if err == nil {
		t.Error("Expected error for empty key, got nil")
	}
}

// Note: Integration tests require a running Zabbix Agent
// Uncomment and set TEST_ZABBIX_AGENT environment variable to run
/*
func TestGetGetValueIntegration(t *testing.T) {
	agent := os.Getenv("TEST_ZABBIX_AGENT")
	if agent == "" {
		t.Skip("Set TEST_ZABBIX_AGENT environment variable to run integration test")
	}

	get := NewGet(agent, 10050)
	get.SetTimeout(10 * time.Second)

	// Try to get a common key
	value, err := get.GetValue("system.uptime")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if value == "" {
		t.Error("Expected non-empty value, got empty string")
	}

	t.Logf("Got value: %s", value)
}

func TestGetGetValuesIntegration(t *testing.T) {
	agent := os.Getenv("TEST_ZABBIX_AGENT")
	if agent == "" {
		t.Skip("Set TEST_ZABBIX_AGENT environment variable to run integration test")
	}

	get := NewGet(agent, 10050)
	get.SetTimeout(10 * time.Second)

	keys := []string{"system.uptime", "system.hostname"}
	values, err := get.GetValues(keys)
	if err != nil {
		t.Fatalf("Failed to get values: %v", err)
	}

	if len(values) != len(keys) {
		t.Errorf("Expected %d values, got %d", len(keys), len(values))
	}

	for _, key := range keys {
		if _, ok := values[key]; !ok {
			t.Errorf("Missing value for key: %s", key)
		}
	}
}
*/
