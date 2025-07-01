# RustPBX Go SDK Examples

This directory contains comprehensive examples demonstrating the usage of the RustPBX Go WebSocket SDK.

## Prerequisites

1. Ensure you have Go 1.21+ installed
2. Run `go mod tidy` to install dependencies
3. Have a RustPBX server running (default: `localhost:8080`)
4. Configure environment variables for TTS/ASR services if needed:
   ```bash
   export TENCENT_SECRET_ID="your_secret_id"
   export TENCENT_SECRET_KEY="your_secret_key"
   export TENCENT_APPID="your_app_id"
   export OPENAI_API_KEY="your_openai_api_key"  # for AI assistant example
   ```

## Available Examples

### 1. Basic Call (`basic_call.go`)

A simple example demonstrating basic call functionality:

```bash
go run examples/basic_call.go
```

**Features demonstrated:**
- WebSocket connection to RustPBX
- Call invitation and acceptance
- Event handling for call lifecycle
- Text-to-Speech (TTS) integration
- Automatic Speech Recognition (ASR)
- DTMF processing

**Key concepts:**
- Setting up event handlers
- Configuring call options
- Managing call state
- Basic error handling

### 2. WebRTC Demo (`webrtc_demo.go`)

Advanced WebRTC functionality with SDP exchange:

```bash
go run examples/webrtc_demo.go
```

**Features demonstrated:**
- WebRTC endpoint connection
- ICE servers configuration
- SDP offer/answer exchange
- ICE candidate handling
- Real-time media processing
- Advanced event processing

**Key concepts:**
- WebRTC signaling flow
- Media negotiation
- ICE connectivity
- RTP media handling

### 3. AI Voice Assistant (`ai_voice_assistant.go`)

Full-featured AI voice assistant with LLM integration:

```bash
go run examples/ai_voice_assistant.go
```

**Features demonstrated:**
- AI-powered conversation
- LLM proxy integration
- Conversation history tracking
- Advanced voice commands
- Dynamic response generation
- Context-aware interactions

**Key concepts:**
- OpenAI-compatible API usage
- Conversation management
- Advanced TTS with emotions
- Intelligent response processing
- Session management

### 4. SIP Integration (`sip_integration.go`)

SIP protocol specific features and telephony integration:

```bash
go run examples/sip_integration.go
```

**Features demonstrated:**
- SIP endpoint connection
- SIP headers and authentication
- Call transfer (REFER method)
- SIP-specific DTMF handling
- Hold/resume functionality
- Conference capabilities

**Key concepts:**
- SIP protocol handling
- Custom SIP headers
- Call transfer mechanisms
- Advanced telephony features

## Building and Running

### Individual Examples

```bash
# Build and run basic call example
go build -o basic_call examples/basic_call.go
./basic_call

# Build and run WebRTC demo
go build -o webrtc_demo examples/webrtc_demo.go
./webrtc_demo

# Build and run AI assistant
go build -o ai_assistant examples/ai_voice_assistant.go
./ai_assistant

# Build and run SIP integration
go build -o sip_integration examples/sip_integration.go
./sip_integration
```

### Using the Makefile

```bash
# Build all examples
make build-examples

# Run specific example
make run-basic
make run-webrtc
make run-ai
make run-sip

# Clean built binaries
make clean
```

## Common Usage Patterns

### 1. Setting Up a Basic Connection

```go
client := rustpbx.NewClient("ws://localhost:8080")
conn, err := client.ConnectCall(ctx, &rustpbx.ConnectionOptions{
    SessionID: "unique-session-id",
    Dump:      true,
})
if err != nil {
    log.Fatal("Connection failed:", err)
}
defer conn.Close()
```

### 2. Handling Events

```go
conn.OnEvent(func(event *rustpbx.Event) {
    switch event.Event {
    case "incoming":
        // Handle incoming call
    case "answer":
        // Call was answered
    case "hangup":
        // Call ended
    case "asrFinal":
        // Speech recognition result
    case "error":
        // Handle errors
    }
})
```

### 3. Configuring Call Options

```go
option := &rustpbx.CallOption{
    Caller: "user@example.com",
    Callee: "agent@example.com",
    Codec:  rustpbx.CodecPCMU,
    TTS: &rustpbx.SynthesisOption{
        Provider:   rustpbx.ProviderTencent,
        Speaker:    "101002",
        SampleRate: 16000,
        Emotion:    rustpbx.EmotionNeutral,
    },
    ASR: &rustpbx.TranscriptionOption{
        Provider:   rustpbx.ProviderTencent,
        Language:   "en-US",
        SampleRate: 16000,
    },
}
```

### 4. Sending Commands

```go
// Invite (start call)
conn.Invite(callOption)

// Accept incoming call
conn.Accept(callOption)

// Send text-to-speech
conn.TTSSimple("Hello, how can I help you?")

// Advanced TTS with options
conn.TTS("Welcome!", "101002", "play-1", &rustpbx.TTSOptions{
    Streaming:   true,
    AutoHangup:  false,
})

// Hangup call
conn.HangupSimple()

// Transfer call
conn.Refer("sip:agent@example.com", &rustpbx.ReferOption{
    Timeout: 30,
    AutoHangup: true,
})
```

## Error Handling

### Connection Errors

```go
conn, err := client.ConnectCall(ctx, options)
if err != nil {
    if wsErr, ok := err.(*rustpbx.WebSocketError); ok {
        log.Printf("WebSocket error: %s (code: %d)", wsErr.Message, wsErr.Code)
    } else {
        log.Printf("Connection error: %v", err)
    }
    return
}
```

### Command Errors

```go
if err := conn.TTSSimple("Hello"); err != nil {
    log.Printf("TTS failed: %v", err)
    // Fallback or retry logic
}
```

### Event-based Errors

```go
conn.OnEvent(func(event *rustpbx.Event) {
    if event.Event == "error" {
        log.Printf("Server error from %s: %s (code: %d)", 
            event.Sender, event.Error, event.Code)
        // Handle specific error types
    }
})
```

## Best Practices

### 1. Resource Management

- Always call `defer conn.Close()` after establishing connection
- Use context for timeout management
- Properly handle connection cleanup on signals

### 2. Event Handling

- Set up event handlers before sending commands
- Handle all relevant event types for your use case
- Log events for debugging and monitoring

### 3. Error Recovery

- Implement retry logic for transient failures
- Provide fallback responses for ASR/TTS failures
- Gracefully handle connection drops

### 4. Configuration

- Use environment variables for sensitive data
- Validate configuration before use
- Provide sensible defaults

### 5. Testing

- Test with different call scenarios
- Verify error handling paths
- Use the examples as integration tests

## Troubleshooting

### Common Issues

1. **Connection Failed**
   - Verify RustPBX server is running
   - Check network connectivity
   - Validate WebSocket URL

2. **TTS/ASR Not Working**
   - Verify API credentials in environment
   - Check provider configuration
   - Ensure proper language settings

3. **Audio Issues**
   - Verify codec compatibility
   - Check sample rate settings
   - Ensure proper media configuration

4. **Event Not Received**
   - Verify event handler is set before sending commands
   - Check for connection drops
   - Review server logs

### Debug Mode

Enable verbose logging by setting the `RUSTPBX_DEBUG` environment variable:

```bash
export RUSTPBX_DEBUG=true
go run examples/basic_call.go
```

## Next Steps

1. Review the API documentation in the main README
2. Explore the source code in the `rustpbx/` package
3. Run the unit tests: `go test ./rustpbx -v`
4. Create your own custom examples based on these templates
5. Integrate the SDK into your own applications

## Support

- Check the main repository documentation
- Review the API specification in `docs/API.md`
- Submit issues for bugs or feature requests
- Contribute improvements via pull requests