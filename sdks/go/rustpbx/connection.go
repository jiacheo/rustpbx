package rustpbx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Connection represents a WebSocket connection to RustPBX
type Connection struct {
	conn         *websocket.Conn
	ctx          context.Context
	cancel       context.CancelFunc
	eventHandler EventHandler
	mu           sync.RWMutex
	closed       bool
	done         chan struct{}
}

// NewConnection creates a new WebSocket connection
func NewConnection(ctx context.Context, wsURL string) (*Connection, error) {
	// Create a cancellable context
	connCtx, cancel := context.WithCancel(ctx)

	// Set up WebSocket dialer
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 30 * time.Second

	// Establish WebSocket connection
	conn, _, err := dialer.DialContext(connCtx, wsURL, http.Header{})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to dial WebSocket: %w", err)
	}

	connection := &Connection{
		conn:   conn,
		ctx:    connCtx,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	// Start reading messages in a goroutine
	go connection.readLoop()

	return connection, nil
}

// OnEvent sets the event handler function
func (c *Connection) OnEvent(handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandler = handler
}

// Close closes the WebSocket connection
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	c.cancel()

	// Send close message
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		// If we can't send close message, just close the connection
		c.conn.Close()
		return err
	}

	// Wait for close or timeout
	select {
	case <-c.done:
		return c.conn.Close()
	case <-time.After(5 * time.Second):
		return c.conn.Close()
	}
}

// readLoop continuously reads messages from the WebSocket
func (c *Connection) readLoop() {
	defer close(c.done)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// Set read deadline
			c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			messageType, data, err := c.conn.ReadMessage()
			if err != nil {
				if !c.isClosed() {
					// Connection closed unexpectedly
					c.handleError(fmt.Errorf("WebSocket read error: %w", err))
				}
				return
			}

			if messageType == websocket.TextMessage {
				c.handleMessage(data)
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Connection) handleMessage(data []byte) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		c.handleError(fmt.Errorf("failed to parse event: %w", err))
		return
	}

	c.mu.RLock()
	handler := c.eventHandler
	c.mu.RUnlock()

	if handler != nil {
		handler(&event)
	}
}

// handleError handles connection errors
func (c *Connection) handleError(err error) {
	c.mu.RLock()
	handler := c.eventHandler
	c.mu.RUnlock()

	if handler != nil {
		errorEvent := &Event{
			Event:     "error",
			Timestamp: time.Now().UnixMilli(),
			Error:     err.Error(),
		}
		handler(errorEvent)
	}
}

// isClosed checks if the connection is closed
func (c *Connection) isClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// sendCommand sends a command to the WebSocket
func (c *Connection) sendCommand(command interface{}) error {
	if c.isClosed() {
		return fmt.Errorf("connection is closed")
	}

	data, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("connection is closed")
	}

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err = c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

// Invite sends an invite command to initiate a call
func (c *Connection) Invite(option *CallOption) error {
	cmd := InviteCommand{
		Command: "invite",
		Option:  option,
	}
	return c.sendCommand(cmd)
}

// Accept sends an accept command to accept an incoming call
func (c *Connection) Accept(option *CallOption) error {
	cmd := AcceptCommand{
		Command: "accept",
		Option:  option,
	}
	return c.sendCommand(cmd)
}

// Reject sends a reject command to reject an incoming call
func (c *Connection) Reject(reason string, code int) error {
	cmd := RejectCommand{
		Command: "reject",
		Reason:  reason,
		Code:    code,
	}
	return c.sendCommand(cmd)
}

// Candidate sends ICE candidates for WebRTC negotiation
func (c *Connection) Candidate(candidates []string) error {
	cmd := CandidateCommand{
		Command:    "candidate",
		Candidates: candidates,
	}
	return c.sendCommand(cmd)
}

// TTS sends a text-to-speech command
func (c *Connection) TTS(text, speaker, playID string, options *TTSOptions) error {
	cmd := TTSCommand{
		Command: "tts",
		Text:    text,
		Speaker: speaker,
		PlayID:  playID,
	}

	if options != nil {
		cmd.AutoHangup = options.AutoHangup
		cmd.Streaming = options.Streaming
		cmd.EndOfStream = options.EndOfStream
	}

	return c.sendCommand(cmd)
}

// TTSSimple sends a simple text-to-speech command with default options
func (c *Connection) TTSSimple(text string) error {
	return c.TTS(text, "", "", nil)
}

// Play sends a play command to play audio from URL
func (c *Connection) Play(url string, autoHangup bool) error {
	cmd := PlayCommand{
		Command:    "play",
		URL:        url,
		AutoHangup: autoHangup,
	}
	return c.sendCommand(cmd)
}

// Interrupt sends an interrupt command to stop current audio playback
func (c *Connection) Interrupt() error {
	cmd := Command{Command: "interrupt"}
	return c.sendCommand(cmd)
}

// Pause sends a pause command to pause audio playback
func (c *Connection) Pause() error {
	cmd := Command{Command: "pause"}
	return c.sendCommand(cmd)
}

// Resume sends a resume command to resume audio playback
func (c *Connection) Resume() error {
	cmd := Command{Command: "resume"}
	return c.sendCommand(cmd)
}

// Hangup sends a hangup command to terminate the call
func (c *Connection) Hangup(reason, initiator string) error {
	cmd := HangupCommand{
		Command:   "hangup",
		Reason:    reason,
		Initiator: initiator,
	}
	return c.sendCommand(cmd)
}

// HangupSimple sends a simple hangup command with default values
func (c *Connection) HangupSimple() error {
	return c.Hangup("normal_clearing", "caller")
}

// Refer sends a refer command to transfer the call
func (c *Connection) Refer(target string, options *ReferOption) error {
	cmd := ReferCommand{
		Command: "refer",
		Target:  target,
		Options: options,
	}
	return c.sendCommand(cmd)
}

// Mute sends a mute command to mute an audio track
func (c *Connection) Mute(trackID string) error {
	cmd := MuteCommand{
		Command: "mute",
		TrackID: trackID,
	}
	return c.sendCommand(cmd)
}

// Unmute sends an unmute command to unmute an audio track
func (c *Connection) Unmute(trackID string) error {
	cmd := UnmuteCommand{
		Command: "unmute",
		TrackID: trackID,
	}
	return c.sendCommand(cmd)
}

// History sends a history command to add conversation context
func (c *Connection) History(speaker, text string) error {
	cmd := HistoryCommand{
		Command: "history",
		Speaker: speaker,
		Text:    text,
	}
	return c.sendCommand(cmd)
}

// SendRawCommand sends a raw command as a JSON object
func (c *Connection) SendRawCommand(command map[string]interface{}) error {
	return c.sendCommand(command)
}

// WaitForEvent waits for a specific event type with timeout
func (c *Connection) WaitForEvent(eventType string, timeout time.Duration) (*Event, error) {
	eventChan := make(chan *Event, 1)
	var originalHandler EventHandler

	// Set up temporary event handler
	c.mu.Lock()
	originalHandler = c.eventHandler
	c.eventHandler = func(event *Event) {
		if event.Event == eventType {
			select {
			case eventChan <- event:
			default:
			}
		}
		// Also call original handler if it exists
		if originalHandler != nil {
			originalHandler(event)
		}
	}
	c.mu.Unlock()

	// Restore original handler when done
	defer func() {
		c.mu.Lock()
		c.eventHandler = originalHandler
		c.mu.Unlock()
	}()

	// Wait for event or timeout
	select {
	case event := <-eventChan:
		return event, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for event: %s", eventType)
	case <-c.ctx.Done():
		return nil, fmt.Errorf("connection closed while waiting for event: %s", eventType)
	}
}