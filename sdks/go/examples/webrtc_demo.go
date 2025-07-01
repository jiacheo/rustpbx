package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rustpbx/go-sdk/rustpbx"
)

func main() {
	// Create a new RustPBX client
	client := rustpbx.NewClient("ws://localhost:8080")

	// Get ICE servers configuration first
	ctx := context.Background()
	iceServers, err := client.GetICEServers(ctx)
	if err != nil {
		log.Printf("Failed to get ICE servers: %v", err)
	} else {
		log.Printf("Retrieved %d ICE servers", len(iceServers))
		for i, server := range iceServers {
			log.Printf("ICE Server %d: %v", i+1, server.URLs)
		}
	}

	// Connect to WebRTC endpoint
	conn, err := client.ConnectWebRTC(ctx, &rustpbx.ConnectionOptions{
		SessionID: "webrtc-demo",
		Dump:      true,
	})
	if err != nil {
		log.Fatal("Failed to connect to WebRTC endpoint:", err)
	}
	defer conn.Close()

	log.Println("Connected to RustPBX WebRTC WebSocket")

	// Track call state
	callConnected := false

	// Set up event handlers
	conn.OnEvent(func(event *rustpbx.Event) {
		log.Printf("WebRTC Event: %s", event.Event)

		switch event.Event {
		case "incoming":
			log.Printf("Incoming WebRTC call from %s to %s", event.Caller, event.Callee)
			log.Printf("Received SDP offer: %s", event.SDP[:100]+"...") // Show first 100 chars

			// Create SDP answer (in real implementation, you'd generate this properly)
			acceptOption := &rustpbx.CallOption{
				Codec: rustpbx.CodecPCMU,
				Offer: generateSDPAnswer(), // You would generate a proper SDP answer here
				TTS: &rustpbx.SynthesisOption{
					Provider:   rustpbx.ProviderTencent,
					Speaker:    "101002",
					SampleRate: 16000,
				},
				ASR: &rustpbx.TranscriptionOption{
					Provider:   rustpbx.ProviderTencent,
					Language:   "en-US",
					SampleRate: 16000,
				},
			}

			if err := conn.Accept(acceptOption); err != nil {
				log.Printf("Failed to accept WebRTC call: %v", err)
			}

		case "answer":
			log.Println("WebRTC call answered")
			if event.SDP != "" {
				log.Printf("Received SDP answer: %s", event.SDP[:100]+"...")
			}
			callConnected = true

			// Send ICE candidates (example candidates)
			candidates := []string{
				"candidate:1 1 UDP 2113667327 192.168.1.100 54400 typ host",
				"candidate:2 1 UDP 1677729535 203.0.113.100 54400 typ srflx raddr 192.168.1.100 rport 54400",
			}
			if err := conn.Candidate(candidates); err != nil {
				log.Printf("Failed to send ICE candidates: %v", err)
			}

			// Start the conversation
			time.AfterFunc(2*time.Second, func() {
				if err := conn.TTSSimple("Hello! This is a WebRTC call. How can I assist you today?"); err != nil {
					log.Printf("Failed to send initial TTS: %v", err)
				}
			})

		case "ringing":
			log.Printf("WebRTC call is ringing (early media: %t)", event.EarlyMedia)

		case "hangup":
			log.Printf("WebRTC call ended: %s (initiated by %s)", event.Reason, event.Initiator)
			callConnected = false

		case "asrFinal":
			log.Printf("Speech recognized: %s", event.Text)
			if callConnected {
				// Process the speech and respond
				response := processUserInput(event.Text)
				if err := conn.TTSSimple(response); err != nil {
					log.Printf("Failed to send TTS response: %v", err)
				}
			}

		case "asrDelta":
			log.Printf("Partial speech: %s", event.Text)

		case "speaking":
			log.Printf("Speaking activity detected on track %s at %d", event.TrackID, event.StartTime)

		case "silence":
			log.Printf("Silence detected on track %s (duration: %dms)", event.TrackID, event.Duration)

		case "dtmf":
			log.Printf("DTMF digit: %s on track %s", event.Digit, event.TrackID)
			handleDTMF(conn, event.Digit)

		case "error":
			log.Printf("WebRTC Error from %s: %s (code: %d)", event.Sender, event.Error, event.Code)
		}
	})

	// Set up WebRTC call option
	webrtcOption := &rustpbx.CallOption{
		Caller:    "webrtc-client@example.com",
		Callee:    "webrtc-agent@example.com",
		Codec:     rustpbx.CodecPCMU,
		EnableIPv6: false,
		Offer:     generateSDPOffer(), // You would generate a proper SDP offer here
		TTS: &rustpbx.SynthesisOption{
			Provider:   rustpbx.ProviderTencent,
			Speaker:    "101002",
			SampleRate: 16000,
			Volume:     5,
			Speed:      1.0,
			Emotion:    rustpbx.EmotionNeutral,
		},
		ASR: &rustpbx.TranscriptionOption{
			Provider:   rustpbx.ProviderTencent,
			Language:   "en-US",
			SampleRate: 16000,
			BufferSize: 1024,
		},
		VAD: &rustpbx.VADOption{
			Type:           rustpbx.VADTypeWebRTC,
			Aggressiveness: 3,
		},
		Recorder: &rustpbx.RecorderOption{
			RecorderFile: "/tmp/webrtc-call-recording.wav",
			SampleRate:   16000,
			PTime:        "20ms",
		},
	}

	// Initiate WebRTC call
	log.Println("Initiating WebRTC call...")
	if err := conn.Invite(webrtcOption); err != nil {
		log.Fatal("Failed to send WebRTC invite:", err)
	}

	// Wait for interrupt signal or timeout
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
		log.Println("Received interrupt signal, ending WebRTC call...")
		if callConnected {
			conn.Hangup("normal_clearing", "caller")
		}
		time.Sleep(2 * time.Second)
	case <-time.After(120 * time.Second):
		log.Println("Demo timeout, ending WebRTC call...")
		if callConnected {
			conn.Hangup("normal_clearing", "caller")
		}
	}

	log.Println("WebRTC demo completed")
}

// generateSDPOffer generates a sample SDP offer (in real implementation, use WebRTC library)
func generateSDPOffer() string {
	return `v=0
o=- 123456789 123456789 IN IP4 192.168.1.100
s=-
c=IN IP4 192.168.1.100
t=0 0
m=audio 54400 RTP/AVP 0
a=rtpmap:0 PCMU/8000
a=sendrecv`
}

// generateSDPAnswer generates a sample SDP answer (in real implementation, use WebRTC library)
func generateSDPAnswer() string {
	return `v=0
o=- 987654321 987654321 IN IP4 192.168.1.101
s=-
c=IN IP4 192.168.1.101
t=0 0
m=audio 54401 RTP/AVP 0
a=rtpmap:0 PCMU/8000
a=sendrecv`
}

// processUserInput processes user speech and generates appropriate responses
func processUserInput(input string) string {
	switch {
	case contains(input, "hello", "hi", "hey"):
		return "Hello! How can I help you today?"
	case contains(input, "time", "what time"):
		return "The current time is " + time.Now().Format("3:04 PM")
	case contains(input, "weather"):
		return "I'm sorry, I don't have access to weather information right now."
	case contains(input, "goodbye", "bye", "end call"):
		return "Goodbye! Have a great day!"
	case contains(input, "test", "testing"):
		return "This is a test response. WebRTC connection is working properly."
	default:
		return "I heard you say: " + input + ". How else can I assist you?"
	}
}

// handleDTMF processes DTMF input
func handleDTMF(conn *rustpbx.Connection, digit string) {
	log.Printf("Processing DTMF digit: %s", digit)
	
	switch digit {
	case "1":
		conn.TTSSimple("You pressed 1. Transferring to support.")
	case "2":
		conn.TTSSimple("You pressed 2. Playing information message.")
		conn.Play("https://example.com/info.wav", false)
	case "9":
		conn.TTSSimple("You pressed 9. Ending call.")
		time.AfterFunc(2*time.Second, func() {
			conn.HangupSimple()
		})
	case "0":
		conn.TTSSimple("You pressed 0. Returning to main menu.")
	default:
		conn.TTSSimple("You pressed " + digit + ". Please try again.")
	}
}

// contains checks if any of the keywords are found in the input string
func contains(input string, keywords ...string) bool {
	input = strings.ToLower(input)
	for _, keyword := range keywords {
		if strings.Contains(input, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}