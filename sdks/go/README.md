# RustPBX Go WebSocket SDK

A Go SDK for interacting with RustPBX WebSocket APIs, providing real-time call management, WebRTC communications, and AI-powered voice processing.

## Features

- **WebSocket Call Management**: Connect to WebSocket endpoints for real-time call handling
- **WebRTC Support**: Full WebRTC call functionality with SDP exchange
- **SIP Integration**: SIP protocol support through WebSocket interface
- **AI-Powered Features**: Text-to-Speech (TTS), Automatic Speech Recognition (ASR)
- **Call Control**: Comprehensive call control commands (invite, accept, reject, hangup, etc.)
- **Media Management**: Audio playback, recording, muting, and interruption
- **Event Handling**: Real-time event processing for call lifecycle management

## Installation

```bash
go mod init your-project
go get github.com/rustpbx/go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/rustpbx/go-sdk/rustpbx"
)

func main() {
    // Create a new WebSocket client
    client := rustpbx.NewClient("ws://localhost:8080")
    
    // Connect to the WebSocket endpoint
    ctx := context.Background()
    conn, err := client.ConnectCall(ctx, &rustpbx.ConnectionOptions{
        SessionID: "my-session-123",
        Dump:      true,
    })
    if err != nil {
        log.Fatal("Failed to connect:", err)
    }
    defer conn.Close()
    
    // Set up event handlers
    conn.OnEvent(func(event *rustpbx.Event) {
        log.Printf("Received event: %s", event.Event)
    })
    
    // Send invite command
    err = conn.Invite(&rustpbx.CallOption{
        Caller: "user@example.com",
        Callee: "agent@example.com",
        Codec:  "pcmu",
        TTS: &rustpbx.SynthesisOption{
            Provider:   "tencent",
            Speaker:    "101002",
            SampleRate: 16000,
        },
    })
    if err != nil {
        log.Fatal("Failed to send invite:", err)
    }
    
    // Keep connection alive
    time.Sleep(30 * time.Second)
}
```

## API Reference

### Client Creation

```go
client := rustpbx.NewClient("ws://localhost:8080")
```

### Connection Types

#### WebSocket Call Connection
```go
conn, err := client.ConnectCall(ctx, options)
```

#### WebRTC Call Connection
```go
conn, err := client.ConnectWebRTC(ctx, options)
```

#### SIP Call Connection
```go
conn, err := client.ConnectSIP(ctx, options)
```

### Commands

#### Call Management
- `Invite(option *CallOption)` - Initiate a call
- `Accept(option *CallOption)` - Accept an incoming call
- `Reject(reason string, code int)` - Reject an incoming call
- `Hangup(reason, initiator string)` - Terminate the call

#### Media Control
- `TTS(text, speaker, playID string, options *TTSOptions)` - Text-to-speech
- `Play(url string, autoHangup bool)` - Play audio from URL
- `Interrupt()` - Interrupt current audio
- `Pause()` - Pause audio playback
- `Resume()` - Resume audio playback

#### Call Control
- `Mute(trackID string)` - Mute audio track
- `Unmute(trackID string)` - Unmute audio track
- `Refer(target string, options *ReferOption)` - Transfer call
- `Candidate(candidates []string)` - Send ICE candidates

### Events

The SDK provides comprehensive event handling for:

- `incoming` - Incoming call notification
- `answer` - Call answered
- `ringing` - Call ringing
- `hangup` - Call terminated
- `asrFinal` - Final speech recognition result
- `asrDelta` - Partial speech recognition result
- `speaking` - Speaker activity detected
- `silence` - Silence detected
- `dtmf` - DTMF tone received
- `error` - Error occurred

### Configuration Options

#### CallOption
Complete call configuration with support for:
- Audio codecs (pcmu, pcma, g722, pcm)
- ASR providers (tencent, voiceapi)
- TTS providers (tencent, voiceapi)
- Voice Activity Detection (VAD)
- Noise suppression and recording

#### ConnectionOptions
- `SessionID` - Custom session identifier
- `Dump` - Enable event dumping to file

## Examples

See the `examples/` directory for comprehensive usage examples:

- **Basic Call**: Simple call setup and management
- **WebRTC Demo**: WebRTC call with SDP exchange
- **AI Voice Assistant**: Full AI-powered voice interaction
- **SIP Integration**: SIP protocol usage
- **Media Playback**: Audio playback and control

## Error Handling

The SDK provides structured error handling with specific error types:

```go
if err != nil {
    if wsErr, ok := err.(*rustpbx.WebSocketError); ok {
        log.Printf("WebSocket error: %s (code: %d)", wsErr.Message, wsErr.Code)
    } else {
        log.Printf("General error: %v", err)
    }
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.