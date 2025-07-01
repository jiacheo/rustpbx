# RustPBX RESTful API Documentation

## Overview

RustPBX is an AI-powered Software-Defined PBX system that provides RESTful APIs for call management, WebRTC communications, SIP handling, and LLM integration. This document covers all HTTP REST endpoints and WebSocket APIs.

## Base URL

```
http://localhost:8080
```

## Authentication

Currently, RustPBX doesn't implement authentication for REST endpoints. For production deployments, consider implementing authentication at the reverse proxy level.

## HTTP Headers

For all API requests, set the appropriate Content-Type header:

```
Content-Type: application/json
```

## Response Format

All API responses are in JSON format. Error responses include error messages and appropriate HTTP status codes.

---

## Call Management APIs

### List Active Calls

Retrieve a list of all currently active calls.

**Endpoint:** `GET /call/lists`

**Response:**
```json
{
  "calls": [
    {
      "id": "uuid-string",
      "call_type": "webrtc" | "sip" | "websocket",
      "created_at": "2023-12-01T10:30:00Z",
      "option": {
        // CallOption object (see CallOption schema below)
      }
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Successfully retrieved call list
- `500 Internal Server Error` - Server error

### Kill/Terminate Call

Forcefully terminate an active call by ID.

**Endpoint:** `POST /call/kill/{id}`

**Parameters:**
- `id` (path parameter): The unique call ID to terminate

**Response:**
```json
true
```

**Status Codes:**
- `200 OK` - Call successfully terminated
- `404 Not Found` - Call ID not found
- `500 Internal Server Error` - Server error

---

## WebSocket Call APIs

### WebSocket Call Connection

Establish a WebSocket connection for real-time call handling.

**Endpoint:** `GET /call`

**Query Parameters:**
- `id` (optional): Custom session ID. If not provided, a UUID will be generated
- `dump` (optional, boolean): Whether to dump events to file. Default: `true`

**Protocol:** WebSocket

**Usage:**
```javascript
const ws = new WebSocket('ws://localhost:8080/call?id=custom-session-id&dump=true');
```

### WebRTC Call Connection

Establish a WebSocket connection specifically for WebRTC calls.

**Endpoint:** `GET /call/webrtc`

**Query Parameters:**
- `id` (optional): Custom session ID
- `dump` (optional, boolean): Whether to dump events to file

**Protocol:** WebSocket

### SIP Call Connection

Establish a WebSocket connection for SIP-based calls.

**Endpoint:** `GET /call/sip`

**Query Parameters:**
- `id` (optional): Custom session ID
- `dump` (optional, boolean): Whether to dump events to file

**Protocol:** WebSocket

---

## WebRTC Support APIs

### Get ICE Servers

Retrieve ICE servers configuration for WebRTC connections.

**Endpoint:** `GET /iceservers`

**Response:**
```json
[
  {
    "urls": ["stun:restsend.com:3478"],
    "username": null,
    "credential": null
  },
  {
    "urls": ["turn:example.com:3478"],
    "username": "user123",
    "credential": "pass123"
  }
]
```

**Status Codes:**
- `200 OK` - Successfully retrieved ICE servers
- `500 Internal Server Error` - Failed to retrieve ICE servers

---

## LLM Proxy APIs

### LLM Proxy Endpoint

Proxy requests to OpenAI-compatible LLM services.

**Endpoint:** `POST /llm/v1/{*path}`

**Description:** This endpoint forwards all requests to the configured LLM service (OpenAI API compatible). The path and query parameters are preserved.

**Headers:**
- `Authorization: Bearer <api-key>` (optional if configured via environment)
- `Content-Type: application/json`

**Environment Variables:**
- `OPENAI_BASE_URL`: Target LLM service URL (default: https://api.openai.com/v1)
- `OPENAI_API_KEY`: API key for authentication

**Examples:**

#### Chat Completions
```bash
curl -X POST http://localhost:8080/llm/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxx" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

#### Models List
```bash
curl http://localhost:8080/llm/v1/models
```

**Response:** The response is forwarded directly from the target LLM service.

**Status Codes:**
- Depends on the target LLM service response
- `500 Internal Server Error` - Proxy error

---

## WebSocket Commands

When connected to any of the call WebSocket endpoints, you can send the following commands:

### Invite Command

Initiate a call with specified options.

```json
{
  "command": "invite",
  "option": {
    // CallOption object (see schema below)
  }
}
```

### Accept Command

Accept an incoming call.

```json
{
  "command": "accept",
  "option": {
    // CallOption object
  }
}
```

### Reject Command

Reject an incoming call.

```json
{
  "command": "reject",
  "reason": "busy",
  "code": 486
}
```

### Candidate Command

Send ICE candidates for WebRTC negotiation.

```json
{
  "command": "candidate",
  "candidates": [
    "candidate:1 1 UDP 2113667327 192.168.1.100 54400 typ host"
  ]
}
```

### TTS (Text-to-Speech) Command

Convert text to speech and play it during the call.

```json
{
  "command": "tts",
  "text": "Hello, how can I help you today?",
  "speaker": "female_voice_01",
  "playId": "play_123",
  "autoHangup": false,
  "streaming": true,
  "endOfStream": false
}
```

### Play Command

Play an audio file from URL.

```json
{
  "command": "play",
  "url": "https://example.com/audio.wav",
  "autoHangup": false
}
```

### Interrupt Command

Interrupt current audio playback.

```json
{
  "command": "interrupt"
}
```

### Pause Command

Pause current audio playback.

```json
{
  "command": "pause"
}
```

### Resume Command

Resume paused audio playback.

```json
{
  "command": "resume"
}
```

### Hangup Command

Terminate the call.

```json
{
  "command": "hangup",
  "reason": "normal_clearing",
  "initiator": "caller"
}
```

### Refer Command

Transfer the call to another destination.

```json
{
  "command": "refer",
  "target": "sip:alice@example.com",
  "options": {
    "bypass": false,
    "timeout": 30,
    "moh": "http://example.com/music.wav",
    "autoHangup": true
  }
}
```

### Mute Command

Mute audio track.

```json
{
  "command": "mute",
  "trackId": "track_123"
}
```

### Unmute Command

Unmute audio track.

```json
{
  "command": "unmute",
  "trackId": "track_123"
}
```

### History Command

Add conversation history entry.

```json
{
  "command": "history",
  "speaker": "user",
  "text": "Previous conversation context"
}
```

---

## WebSocket Events

The WebSocket connection will send various events during the call lifecycle:

### Incoming Event

```json
{
  "event": "incoming",
  "trackId": "track_123",
  "timestamp": 1701423000000,
  "caller": "sip:bob@example.com",
  "callee": "sip:alice@example.com",
  "sdp": "v=0\r\no=- 123456 123456 IN IP4 192.168.1.100\r\n..."
}
```

### Answer Event

```json
{
  "event": "answer",
  "trackId": "track_123",
  "timestamp": 1701423001000,
  "sdp": "v=0\r\no=- 789012 789012 IN IP4 192.168.1.101\r\n..."
}
```

### Ringing Event

```json
{
  "event": "ringing",
  "trackId": "track_123",
  "timestamp": 1701423002000,
  "earlyMedia": false
}
```

### Hangup Event

```json
{
  "event": "hangup",
  "timestamp": 1701423003000,
  "reason": "normal_clearing",
  "initiator": "caller"
}
```

### ASR (Speech Recognition) Events

```json
{
  "event": "asrFinal",
  "trackId": "track_123",
  "timestamp": 1701423004000,
  "index": 1,
  "startTime": 1000,
  "endTime": 3000,
  "text": "Hello, how are you?"
}
```

```json
{
  "event": "asrDelta",
  "trackId": "track_123",
  "timestamp": 1701423005000,
  "index": 2,
  "startTime": 3000,
  "endTime": 4000,
  "text": "I'm doing well"
}
```

### Speaking/Silence Events

```json
{
  "event": "speaking",
  "trackId": "track_123",
  "timestamp": 1701423006000,
  "startTime": 1701423000000
}
```

```json
{
  "event": "silence",
  "trackId": "track_123",
  "timestamp": 1701423007000,
  "startTime": 1701423005000,
  "duration": 2000
}
```

### DTMF Event

```json
{
  "event": "dtmf",
  "trackId": "track_123",
  "timestamp": 1701423008000,
  "digit": "1"
}
```

### Error Event

```json
{
  "event": "error",
  "trackId": "track_123",
  "timestamp": 1701423009000,
  "sender": "tts_engine",
  "error": "TTS service unavailable",
  "code": 503
}
```

---

## Data Schemas

### CallOption

The main configuration object for call setup and behavior.

```json
{
  "denoise": true,
  "offer": "v=0\r\no=- 123456...",
  "callee": "sip:alice@example.com",
  "caller": "sip:bob@example.com",
  "recorder": {
    "recorderFile": "/tmp/recording.wav",
    "samplerate": 16000,
    "ptime": "200ms"
  },
  "vad": {
    "type": "webrtc",
    "aggressiveness": 3
  },
  "asr": {
    "provider": "tencent",
    "model": "16k_zh",
    "language": "zh-CN",
    "appId": "your_app_id",
    "secretId": "your_secret_id",
    "secretKey": "your_secret_key",
    "modelType": "16k_zh",
    "bufferSize": 1024,
    "samplerate": 16000,
    "endpoint": "https://asr.tencentcloudapi.com",
    "extra": {}
  },
  "tts": {
    "samplerate": 16000,
    "provider": "tencent",
    "speed": 1.0,
    "appId": "your_app_id",
    "secretId": "your_secret_id",
    "secretKey": "your_secret_key",
    "volume": 5,
    "speaker": "101002",
    "codec": "pcm",
    "subtitle": false,
    "emotion": "neutral",
    "endpoint": "https://tts.tencentcloudapi.com",
    "extra": {}
  },
  "handshakeTimeout": "30s",
  "enableIpv6": false,
  "sip": {
    "username": "user123",
    "password": "pass123",
    "realm": "example.com",
    "headers": {
      "X-Custom-Header": "value"
    }
  },
  "extra": {
    "custom_field": "custom_value"
  },
  "codec": "pcmu",
  "eou": {
    "type": "tencent",
    "endpoint": "https://eou.example.com",
    "secretKey": "eou_secret",
    "secretId": "eou_id",
    "timeout": 5000
  }
}
```

### RecorderOption

```json
{
  "recorderFile": "/tmp/recordings/session_123.wav",
  "samplerate": 16000,
  "ptime": "200ms"
}
```

### VADOption (Voice Activity Detection)

```json
{
  "type": "webrtc" | "silero" | "ten",
  "aggressiveness": 3
}
```

### TranscriptionOption (ASR)

```json
{
  "provider": "tencent" | "voiceapi",
  "model": "16k_zh",
  "language": "zh-CN",
  "appId": "your_app_id",
  "secretId": "your_secret_id",
  "secretKey": "your_secret_key",
  "modelType": "16k_zh",
  "bufferSize": 1024,
  "samplerate": 16000,
  "endpoint": "https://asr.tencentcloudapi.com",
  "extra": {}
}
```

### SynthesisOption (TTS)

```json
{
  "samplerate": 16000,
  "provider": "tencent" | "voiceapi",
  "speed": 1.0,
  "appId": "your_app_id",
  "secretId": "your_secret_id",
  "secretKey": "your_secret_key",
  "volume": 5,
  "speaker": "101002",
  "codec": "pcm",
  "subtitle": false,
  "emotion": "neutral",
  "endpoint": "https://tts.tencentcloudapi.com",
  "extra": {}
}
```

### SipOption

```json
{
  "username": "user123",
  "password": "pass123",
  "realm": "example.com",
  "headers": {
    "X-Custom-Header": "value"
  }
}
```

### EouOption (End of Utterance)

```json
{
  "type": "tencent",
  "endpoint": "https://eou.example.com",
  "secretKey": "eou_secret",
  "secretId": "eou_id",
  "timeout": 5000
}
```

### ReferOption

```json
{
  "bypass": false,
  "timeout": 30,
  "moh": "http://example.com/music.wav",
  "autoHangup": true
}
```

---

## Configuration Options

### TTS Emotions (Tencent Cloud)

Available emotions for TTS synthesis:
- `neutral` - Default neutral voice
- `sad` - Sad emotion
- `happy` - Happy emotion
- `angry` - Angry emotion
- `fear` - Fearful emotion
- `news` - News broadcasting style
- `story` - Storytelling style
- `radio` - Radio broadcasting style
- `poetry` - Poetry reading style
- `call` - Phone call style
- `sajiao` - Coquettish style
- `disgusted` - Disgusted emotion
- `amaze` - Amazed emotion
- `peaceful` - Peaceful emotion
- `exciting` - Exciting emotion
- `aojiao` - Proud emotion
- `jieshuo` - Commentary style

### Audio Codecs

Supported audio codecs:
- `pcmu` - G.711 μ-law
- `pcma` - G.711 A-law
- `g722` - G.722 wideband
- `pcm` - Linear PCM

### Call Types

- `webrtc` - WebRTC-based calls
- `sip` - SIP protocol calls
- `websocket` - Direct WebSocket audio streaming

---

## Error Handling

### HTTP Error Responses

```json
{
  "error": "Error message description",
  "code": "ERROR_CODE",
  "details": "Additional error details"
}
```

### Common HTTP Status Codes

- `200 OK` - Request successful
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service temporarily unavailable

### WebSocket Error Events

Errors during WebSocket communication are sent as error events:

```json
{
  "event": "error",
  "trackId": "track_123",
  "timestamp": 1701423009000,
  "sender": "component_name",
  "error": "Error description",
  "code": 500
}
```

---

## Usage Examples

### Basic WebRTC Call

```javascript
// Connect to WebRTC endpoint
const ws = new WebSocket('ws://localhost:8080/call/webrtc');

// Send invite with basic options
ws.send(JSON.stringify({
  command: 'invite',
  option: {
    caller: 'user@example.com',
    callee: 'agent@example.com',
    codec: 'pcmu',
    tts: {
      provider: 'tencent',
      speaker: '101002',
      samplerate: 16000
    },
    asr: {
      provider: 'tencent',
      language: 'zh-CN',
      samplerate: 16000
    }
  }
}));

// Listen for events
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received event:', data);
};
```

### TTS Playback

```javascript
// Play text-to-speech
ws.send(JSON.stringify({
  command: 'tts',
  text: 'Welcome to our service. How can I help you today?',
  speaker: '101002',
  autoHangup: false
}));
```

### Call Management

```bash
# List active calls
curl http://localhost:8080/call/lists

# Kill a specific call
curl -X POST http://localhost:8080/call/kill/session-uuid-123
```

---

## Integration Notes

### Environment Variables

Key environment variables for configuration:

- `OPENAI_BASE_URL` - LLM service endpoint
- `OPENAI_API_KEY` - LLM service API key
- `TENCENT_APPID` - Tencent Cloud App ID
- `TENCENT_SECRET_ID` - Tencent Cloud Secret ID
- `TENCENT_SECRET_KEY` - Tencent Cloud Secret Key
- `VOICEAPI_ENDPOINT` - VoiceAPI service endpoint
- `VOICEAPI_SPEAKER_ID` - Default speaker ID for VoiceAPI
- `RESTSEND_TOKEN` - Token for external ICE servers

### File Paths

- Call recordings: `/tmp/recorders/{session_id}.wav`
- Event dumps: `/tmp/recorders/{session_id}.events.jsonl`
- Media cache: `/tmp/mediacache/`

### Call Flow

1. **WebSocket Connection**: Client connects to appropriate endpoint
2. **Call Invitation**: Send `invite` command with `CallOption`
3. **SDP Exchange**: Handle `incoming`, `answer` events
4. **Media Processing**: Real-time audio with TTS/ASR
5. **Call Management**: Use commands for control and playback
6. **Call Termination**: Send `hangup` command or handle `hangup` event

This API documentation covers all RESTful endpoints and WebSocket commands available in RustPBX. For additional details on specific configurations or advanced usage, refer to the configuration examples and source code documentation.