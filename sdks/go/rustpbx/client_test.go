package rustpbx

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("ws://localhost:8080")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	
	if client.baseURL != "ws://localhost:8080" {
		t.Errorf("Expected baseURL to be 'ws://localhost:8080', got '%s'", client.baseURL)
	}
}

func TestBuildWebSocketURL(t *testing.T) {
	client := NewClient("http://localhost:8080")
	
	tests := []struct {
		endpoint  string
		sessionID string
		dump      bool
		expected  string
	}{
		{"/call", "test-session", true, "ws://localhost:8080/call?dump=true&id=test-session"},
		{"/call/webrtc", "", false, "ws://localhost:8080/call/webrtc?dump=false"},
		{"/call/sip", "sip-session", true, "ws://localhost:8080/call/sip?dump=true&id=sip-session"},
	}
	
	for _, test := range tests {
		result, err := client.buildWebSocketURL(test.endpoint, test.sessionID, test.dump)
		if err != nil {
			t.Errorf("buildWebSocketURL failed: %v", err)
			continue
		}
		
		// Check if all expected parameters are present
		if test.sessionID != "" && !contains(result, "id="+test.sessionID) {
			t.Errorf("Expected URL to contain session ID '%s', got '%s'", test.sessionID, result)
		}
		
		if !contains(result, "dump=") {
			t.Errorf("Expected URL to contain dump parameter, got '%s'", result)
		}
	}
}

func TestCallOptionValidation(t *testing.T) {
	// Test basic CallOption creation
	option := &CallOption{
		Caller: "test@example.com",
		Callee: "agent@example.com",
		Codec:  CodecPCMU,
		TTS: &SynthesisOption{
			Provider:   ProviderTencent,
			Speaker:    "101002",
			SampleRate: 16000,
		},
	}
	
	if option.Caller != "test@example.com" {
		t.Errorf("Expected caller to be 'test@example.com', got '%s'", option.Caller)
	}
	
	if option.Codec != CodecPCMU {
		t.Errorf("Expected codec to be '%s', got '%s'", CodecPCMU, option.Codec)
	}
	
	if option.TTS.Provider != ProviderTencent {
		t.Errorf("Expected TTS provider to be '%s', got '%s'", ProviderTencent, option.TTS.Provider)
	}
}

func TestEventHandling(t *testing.T) {
	// Test event creation and handling
	event := &Event{
		Event:     "incoming",
		TrackID:   "track-123",
		Timestamp: time.Now().UnixMilli(),
		Caller:    "caller@example.com",
		Callee:    "callee@example.com",
	}
	
	if event.Event != "incoming" {
		t.Errorf("Expected event type to be 'incoming', got '%s'", event.Event)
	}
	
	if event.TrackID != "track-123" {
		t.Errorf("Expected track ID to be 'track-123', got '%s'", event.TrackID)
	}
}

func TestConnectionOptions(t *testing.T) {
	options := &ConnectionOptions{
		SessionID: "test-session",
		Dump:      true,
	}
	
	if options.SessionID != "test-session" {
		t.Errorf("Expected session ID to be 'test-session', got '%s'", options.SessionID)
	}
	
	if !options.Dump {
		t.Error("Expected dump to be true")
	}
}

func TestWebSocketError(t *testing.T) {
	err := &WebSocketError{
		Message: "Connection failed",
		Code:    1001,
	}
	
	if err.Error() != "Connection failed" {
		t.Errorf("Expected error message to be 'Connection failed', got '%s'", err.Error())
	}
	
	if err.Code != 1001 {
		t.Errorf("Expected error code to be 1001, got %d", err.Code)
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		(len(s) > len(substr) && s[:len(substr)] == substr) ||
		containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}