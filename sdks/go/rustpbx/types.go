package rustpbx

import (
	"encoding/json"
	"time"
)

// CallType represents the type of call
type CallType string

const (
	CallTypeWebRTC    CallType = "webrtc"
	CallTypeSIP       CallType = "sip"
	CallTypeWebSocket CallType = "websocket"
)

// Codec represents audio codec types
type Codec string

const (
	CodecPCMU Codec = "pcmu" // G.711 μ-law
	CodecPCMA Codec = "pcma" // G.711 A-law
	CodecG722 Codec = "g722" // G.722 wideband
	CodecPCM  Codec = "pcm"  // Linear PCM
)

// VADType represents Voice Activity Detection types
type VADType string

const (
	VADTypeWebRTC VADType = "webrtc"
	VADTypeSilero VADType = "silero"
	VADTypeTen    VADType = "ten"
)

// Provider represents service providers
type Provider string

const (
	ProviderTencent   Provider = "tencent"
	ProviderVoiceAPI  Provider = "voiceapi"
)

// EOUType represents End of Utterance detection types
type EOUType string

const (
	EOUTypeTencent EOUType = "tencent"
)

// TTSEmotion represents TTS emotion types
type TTSEmotion string

const (
	EmotionNeutral   TTSEmotion = "neutral"
	EmotionSad       TTSEmotion = "sad"
	EmotionHappy     TTSEmotion = "happy"
	EmotionAngry     TTSEmotion = "angry"
	EmotionFear      TTSEmotion = "fear"
	EmotionNews      TTSEmotion = "news"
	EmotionStory     TTSEmotion = "story"
	EmotionRadio     TTSEmotion = "radio"
	EmotionPoetry    TTSEmotion = "poetry"
	EmotionCall      TTSEmotion = "call"
	EmotionSajiao    TTSEmotion = "sajiao"
	EmotionDisgusted TTSEmotion = "disgusted"
	EmotionAmaze     TTSEmotion = "amaze"
	EmotionPeaceful  TTSEmotion = "peaceful"
	EmotionExciting  TTSEmotion = "exciting"
	EmotionAojiao    TTSEmotion = "aojiao"
	EmotionJieshuo   TTSEmotion = "jieshuo"
)

// RecorderOption represents recording configuration
type RecorderOption struct {
	RecorderFile string `json:"recorderFile,omitempty"`
	SampleRate   int    `json:"samplerate,omitempty"`
	PTime        string `json:"ptime,omitempty"`
}

// VADOption represents Voice Activity Detection configuration
type VADOption struct {
	Type           VADType `json:"type,omitempty"`
	Aggressiveness int     `json:"aggressiveness,omitempty"`
}

// TranscriptionOption represents ASR configuration
type TranscriptionOption struct {
	Provider   Provider          `json:"provider,omitempty"`
	Model      string            `json:"model,omitempty"`
	Language   string            `json:"language,omitempty"`
	AppID      string            `json:"appId,omitempty"`
	SecretID   string            `json:"secretId,omitempty"`
	SecretKey  string            `json:"secretKey,omitempty"`
	ModelType  string            `json:"modelType,omitempty"`
	BufferSize int               `json:"bufferSize,omitempty"`
	SampleRate int               `json:"samplerate,omitempty"`
	Endpoint   string            `json:"endpoint,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// SynthesisOption represents TTS configuration
type SynthesisOption struct {
	SampleRate int                    `json:"samplerate,omitempty"`
	Provider   Provider               `json:"provider,omitempty"`
	Speed      float64                `json:"speed,omitempty"`
	AppID      string                 `json:"appId,omitempty"`
	SecretID   string                 `json:"secretId,omitempty"`
	SecretKey  string                 `json:"secretKey,omitempty"`
	Volume     int                    `json:"volume,omitempty"`
	Speaker    string                 `json:"speaker,omitempty"`
	Codec      string                 `json:"codec,omitempty"`
	Subtitle   bool                   `json:"subtitle,omitempty"`
	Emotion    TTSEmotion             `json:"emotion,omitempty"`
	Endpoint   string                 `json:"endpoint,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// SipOption represents SIP configuration
type SipOption struct {
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
	Realm    string            `json:"realm,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}

// EouOption represents End of Utterance configuration
type EouOption struct {
	Type      EOUType `json:"type,omitempty"`
	Endpoint  string  `json:"endpoint,omitempty"`
	SecretKey string  `json:"secretKey,omitempty"`
	SecretID  string  `json:"secretId,omitempty"`
	Timeout   int     `json:"timeout,omitempty"`
}

// ReferOption represents call transfer configuration
type ReferOption struct {
	Bypass     bool   `json:"bypass,omitempty"`
	Timeout    int    `json:"timeout,omitempty"`
	MOH        string `json:"moh,omitempty"`
	AutoHangup bool   `json:"autoHangup,omitempty"`
}

// CallOption represents the main call configuration
type CallOption struct {
	Denoise          bool                     `json:"denoise,omitempty"`
	Offer            string                   `json:"offer,omitempty"`
	Callee           string                   `json:"callee,omitempty"`
	Caller           string                   `json:"caller,omitempty"`
	Recorder         *RecorderOption          `json:"recorder,omitempty"`
	VAD              *VADOption               `json:"vad,omitempty"`
	ASR              *TranscriptionOption     `json:"asr,omitempty"`
	TTS              *SynthesisOption         `json:"tts,omitempty"`
	HandshakeTimeout string                   `json:"handshakeTimeout,omitempty"`
	EnableIPv6       bool                     `json:"enableIpv6,omitempty"`
	SIP              *SipOption               `json:"sip,omitempty"`
	Extra            map[string]interface{}   `json:"extra,omitempty"`
	Codec            Codec                    `json:"codec,omitempty"`
	EOU              *EouOption               `json:"eou,omitempty"`
}

// TTSOptions represents TTS command options
type TTSOptions struct {
	Speaker       string `json:"speaker,omitempty"`
	PlayID        string `json:"playId,omitempty"`
	AutoHangup    bool   `json:"autoHangup,omitempty"`
	Streaming     bool   `json:"streaming,omitempty"`
	EndOfStream   bool   `json:"endOfStream,omitempty"`
}

// Command represents WebSocket commands
type Command struct {
	Command string `json:"command"`
}

// InviteCommand represents invite command
type InviteCommand struct {
	Command string      `json:"command"`
	Option  *CallOption `json:"option"`
}

// AcceptCommand represents accept command
type AcceptCommand struct {
	Command string      `json:"command"`
	Option  *CallOption `json:"option"`
}

// RejectCommand represents reject command
type RejectCommand struct {
	Command string `json:"command"`
	Reason  string `json:"reason"`
	Code    int    `json:"code"`
}

// CandidateCommand represents candidate command
type CandidateCommand struct {
	Command    string   `json:"command"`
	Candidates []string `json:"candidates"`
}

// TTSCommand represents TTS command
type TTSCommand struct {
	Command     string `json:"command"`
	Text        string `json:"text"`
	Speaker     string `json:"speaker,omitempty"`
	PlayID      string `json:"playId,omitempty"`
	AutoHangup  bool   `json:"autoHangup,omitempty"`
	Streaming   bool   `json:"streaming,omitempty"`
	EndOfStream bool   `json:"endOfStream,omitempty"`
}

// PlayCommand represents play command
type PlayCommand struct {
	Command    string `json:"command"`
	URL        string `json:"url"`
	AutoHangup bool   `json:"autoHangup,omitempty"`
}

// HangupCommand represents hangup command
type HangupCommand struct {
	Command   string `json:"command"`
	Reason    string `json:"reason,omitempty"`
	Initiator string `json:"initiator,omitempty"`
}

// ReferCommand represents refer command
type ReferCommand struct {
	Command string       `json:"command"`
	Target  string       `json:"target"`
	Options *ReferOption `json:"options,omitempty"`
}

// MuteCommand represents mute command
type MuteCommand struct {
	Command string `json:"command"`
	TrackID string `json:"trackId"`
}

// UnmuteCommand represents unmute command
type UnmuteCommand struct {
	Command string `json:"command"`
	TrackID string `json:"trackId"`
}

// HistoryCommand represents history command
type HistoryCommand struct {
	Command string `json:"command"`
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
}

// Event represents WebSocket events
type Event struct {
	Event     string          `json:"event"`
	TrackID   string          `json:"trackId,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Caller    string          `json:"caller,omitempty"`
	Callee    string          `json:"callee,omitempty"`
	SDP       string          `json:"sdp,omitempty"`
	EarlyMedia bool           `json:"earlyMedia,omitempty"`
	Reason    string          `json:"reason,omitempty"`
	Initiator string          `json:"initiator,omitempty"`
	Index     int             `json:"index,omitempty"`
	StartTime int64           `json:"startTime,omitempty"`
	EndTime   int64           `json:"endTime,omitempty"`
	Text      string          `json:"text,omitempty"`
	Duration  int64           `json:"duration,omitempty"`
	Digit     string          `json:"digit,omitempty"`
	Sender    string          `json:"sender,omitempty"`
	Error     string          `json:"error,omitempty"`
	Code      int             `json:"code,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// Call represents an active call
type Call struct {
	ID        string      `json:"id"`
	CallType  CallType    `json:"call_type"`
	CreatedAt time.Time   `json:"created_at"`
	Option    *CallOption `json:"option"`
}

// CallListResponse represents the response from /call/lists
type CallListResponse struct {
	Calls []Call `json:"calls"`
}

// ICEServer represents ICE server configuration
type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   *string  `json:"username"`
	Credential *string  `json:"credential"`
}

// ConnectionOptions represents WebSocket connection options
type ConnectionOptions struct {
	SessionID string
	Dump      bool
}

// EventHandler represents an event handler function
type EventHandler func(event *Event)

// WebSocketError represents WebSocket specific errors
type WebSocketError struct {
	Message string
	Code    int
}

func (e *WebSocketError) Error() string {
	return e.Message
}