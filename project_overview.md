# How RustPBX Works - Project Overview

## 📋 Project Summary

**RustPBX** is a high-performance, AI-powered software-defined PBX (Private Branch Exchange) system implemented in Rust. It serves as a modern telephony platform that combines traditional SIP-based voice communications with cutting-edge AI services for speech processing and intelligent conversation handling.

## 🏗️ Core Architecture

### Main Components

The project is structured around several key modules:

1. **SIP Proxy Server** (`src/proxy/`) - Full SIP protocol implementation
2. **User Agent** (`src/useragent/`) - SIP client functionality for outbound calls  
3. **Media Engine** (`src/media/`) - Audio processing, codecs, and streaming
4. **HTTP/WebSocket API** (`src/handler/`) - REST API and real-time communication
5. **AI Services** (`src/synthesis/`, `src/transcription/`, `src/llm/`) - Speech and language processing
6. **Call Recording** (`src/callrecord/`) - Call storage and management

### Entry Point (`src/bin/rustpbx.rs`)

The application starts with:
1. **Configuration Loading** - Reads TOML config file or uses defaults
2. **Logging Setup** - Configures tracing with file/console output
3. **Application State Building** - Initializes core components
4. **Service Startup** - Launches SIP proxy, HTTP server, and user agent

## 🔧 Core Functionality

### 1. SIP Proxy Server (`src/proxy/`)

The SIP proxy implements a full SIP stack with modular architecture:

- **Server** (`server.rs`) - Main SIP server with UDP/TCP/TLS/WebSocket transport
- **Authentication** (`auth.rs`) - User authentication and registration
- **Registrar** (`registrar.rs`) - SIP user registration handling
- **Call Routing** (`call.rs`) - Call establishment and routing logic
- **ACL** (`acl.rs`) - Access control lists for security
- **Media Proxy** (`mediaproxy/`) - RTP/RTCP media relay with NAT traversal

**Key Features:**
- Multi-transport support (UDP, TCP, TLS, WebSocket)
- Flexible user backends (memory, HTTP, database, plain text)
- Media proxy with NAT detection and traversal
- Call Detail Records (CDR) generation
- Realm-based authentication

### 2. User Agent (`src/useragent/`)

Implements SIP User Agent functionality for:
- **Outbound Calls** - Initiating calls to external systems
- **Registration** - Registering with external SIP servers
- **Invitation Handling** - Processing incoming call invitations
- **Media Negotiation** - SDP negotiation and codec selection

### 3. Media Processing (`src/media/`)

Comprehensive audio processing pipeline:

- **Engine** (`engine.rs`) - Central media processing coordinator
- **Codecs** (`codecs/`) - Audio codec support (PCMU, PCMA, G.722, PCM)
- **Streaming** (`stream.rs`) - Real-time audio streaming
- **Recording** (`recorder.rs`) - Call recording functionality
- **Jitter Buffer** (`jitter.rs`) - Audio packet buffering and smoothing
- **Noise Reduction** (`denoiser.rs`) - Audio denoising using nnnoiseless
- **VAD** (`vad/`) - Voice Activity Detection (WebRTC, Silero)
- **DTMF** (`dtmf.rs`) - Dual-tone multi-frequency detection

### 4. AI Voice Services

#### Speech Recognition (`src/transcription/`)
- Real-time speech-to-text conversion
- Multiple provider support (Tencent Cloud, VoiceAPI)
- Streaming ASR for low-latency processing

#### Speech Synthesis (`src/synthesis/`)
- Text-to-speech conversion
- Emotion and speaker control
- Multiple TTS engines support

#### LLM Integration (`src/llm/`)
- OpenAI-compatible API proxy
- Intelligent conversation handling
- Context-aware responses

### 5. HTTP API & WebSocket (`src/handler/`)

RESTful API and real-time communication:

- **Call Management** (`call.rs`) - Start, control, and monitor calls
- **WebRTC Support** (`webrtc.rs`) - Browser-based communications
- **SIP Handling** (`sip.rs`) - SIP message processing via HTTP
- **LLM Proxy** (`llmproxy.rs`) - AI service integration
- **Middleware** (`middleware/`) - Authentication, CORS, client IP detection

## 🔄 System Workflows

### SIP Call Flow

1. **Registration** - SIP clients register with the proxy server
2. **Authentication** - User credentials validated against configured backend
3. **Call Initiation** - Caller sends INVITE to proxy
4. **Route Resolution** - Proxy determines callee location
5. **Call Establishment** - SIP signaling establishes call
6. **Media Relay** - RTP media flows through media proxy (if enabled)
7. **Call Termination** - BYE message ends the call
8. **CDR Generation** - Call details recorded for billing/analytics

### WebRTC Call Flow

1. **WebSocket Connection** - Browser connects to WebSocket endpoint
2. **SDP Exchange** - Media capabilities negotiated
3. **ICE Negotiation** - NAT traversal using STUN/TURN servers
4. **Media Flow** - Direct peer-to-peer or proxied media
5. **Call Control** - Real-time call management via WebSocket

### AI Voice Processing

1. **Audio Capture** - Voice input from SIP/WebRTC
2. **Preprocessing** - Noise reduction, VAD, format conversion
3. **ASR Processing** - Speech-to-text conversion
4. **LLM Processing** - Natural language understanding and response generation
5. **TTS Synthesis** - Text-to-speech conversion
6. **Audio Output** - Synthesized speech sent to caller

## ⚙️ Configuration System

The system uses TOML configuration files with the following main sections:

### HTTP Server
```toml
http_addr = "0.0.0.0:8080"  # Web interface and API
log_level = "info"          # Logging verbosity
```

### SIP Proxy
```toml
[proxy]
addr = "0.0.0.0"
udp_port = 15060           # SIP UDP port
modules = ["acl", "auth", "registrar", "call"]
```

### User Agent
```toml
[ua]
addr = "0.0.0.0"
udp_port = 13050           # User agent SIP port
rtp_start_port = 12000     # RTP port range
rtp_end_port = 42000
```

### Media Proxy
```toml
[proxy.media_proxy]
mode = "nat_only"          # none, nat_only, all
rtp_start_port = 20000
rtp_end_port = 30000
external_ip = "192.168.1.1"
```

## 🔐 Security Features

- **ACL Rules** - IP-based access control
- **Authentication** - SIP digest authentication
- **TLS Support** - Encrypted SIP signaling
- **Realm-based Security** - Multi-tenant support
- **Rate Limiting** - Configurable concurrency limits

## 📊 Monitoring & Recording

- **Call Detail Records** - Comprehensive call logging
- **Audio Recording** - Configurable call recording
- **Storage Backends** - Local, S3, HTTP webhook support
- **Real-time Metrics** - Active call monitoring
- **Event Logging** - Detailed system event tracking

## 🚀 Deployment Options

### Standalone Deployment
```bash
cargo build --release
./target/release/rustpbx --conf config.toml
```

### Docker Deployment
```bash
docker build -t rustpbx .
docker run -p 8080:8080 -p 5060:5060/udp rustpbx
```

### Key Features for Production

- **High Performance** - Async Rust implementation
- **Scalability** - Configurable concurrency and resource limits
- **Reliability** - Graceful shutdown and error handling
- **Observability** - Comprehensive logging and monitoring
- **Extensibility** - Modular architecture for custom functionality

## 🔧 Development Features

The project includes comprehensive testing and development tools:

- **Unit Tests** - Module-level testing throughout
- **Integration Tests** - End-to-end workflow testing
- **Example Applications** - WebRTC demo and voice processing examples
- **API Documentation** - Complete REST API reference
- **Docker Support** - Containerized deployment

This architecture makes RustPBX suitable for various use cases from simple SIP proxy servers to complex AI-powered voice applications and contact centers.