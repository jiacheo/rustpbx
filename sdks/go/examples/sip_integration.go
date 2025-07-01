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

	// Connect to SIP endpoint
	ctx := context.Background()
	conn, err := client.ConnectSIP(ctx, &rustpbx.ConnectionOptions{
		SessionID: "sip-integration-demo",
		Dump:      true,
	})
	if err != nil {
		log.Fatal("Failed to connect to SIP endpoint:", err)
	}
	defer conn.Close()

	log.Println("Connected to RustPBX SIP WebSocket")

	// Track call state
	callActive := false
	currentSession := ""

	// Set up SIP event handlers
	conn.OnEvent(func(event *rustpbx.Event) {
		log.Printf("SIP Event: %s", event.Event)

		switch event.Event {
		case "incoming":
			log.Printf("Incoming SIP call from %s to %s", event.Caller, event.Callee)
			currentSession = event.TrackID

			// SIP-specific auto-accept configuration
			acceptOption := &rustpbx.CallOption{
				Codec: rustpbx.CodecPCMU,
				SIP: &rustpbx.SipOption{
					Username: "pbx-agent",
					Password: "secure-password",
					Realm:    "example.com",
					Headers: map[string]string{
						"X-Call-Type":   "automated",
						"X-Session-ID":  currentSession,
						"User-Agent":    "RustPBX-Go-SDK/1.0",
						"X-Forwarded":   "ai-assistant",
					},
				},
				TTS: &rustpbx.SynthesisOption{
					Provider:   rustpbx.ProviderTencent,
					Speaker:    "101002",
					SampleRate: 16000,
					Volume:     5,
					Emotion:    rustpbx.EmotionNeutral,
				},
				ASR: &rustpbx.TranscriptionOption{
					Provider:   rustpbx.ProviderTencent,
					Language:   "en-US",
					SampleRate: 16000,
				},
				VAD: &rustpbx.VADOption{
					Type:           rustpbx.VADTypeWebRTC,
					Aggressiveness: 3,
				},
				Recorder: &rustpbx.RecorderOption{
					RecorderFile: "/tmp/sip-call-" + time.Now().Format("20060102-150405") + ".wav",
					SampleRate:   16000,
					PTime:        "20ms",
				},
			}

			if err := conn.Accept(acceptOption); err != nil {
				log.Printf("Failed to accept SIP call: %v", err)
			}

		case "answer":
			log.Println("SIP call answered and connected")
			callActive = true

			// Send SIP-specific greeting
			greeting := "Welcome to our SIP-based service. You are connected through Session Initiation Protocol. How may I assist you today?"
			if err := conn.TTSSimple(greeting); err != nil {
				log.Printf("Failed to send SIP greeting: %v", err)
			}

		case "ringing":
			log.Printf("SIP call is ringing (early media: %t)", event.EarlyMedia)

		case "hangup":
			log.Printf("SIP call ended: %s (initiated by %s)", event.Reason, event.Initiator)
			callActive = false

		case "asrFinal":
			if !callActive {
				return
			}
			
			userInput := strings.TrimSpace(event.Text)
			log.Printf("SIP User Input: %s", userInput)

			// Process SIP-specific commands
			response := processSIPCommand(userInput)
			if err := conn.TTSSimple(response); err != nil {
				log.Printf("Failed to send SIP response: %v", err)
			}

		case "asrDelta":
			log.Printf("SIP Partial: %s", event.Text)

		case "speaking":
			log.Printf("SIP Speaking detected on track %s", event.TrackID)

		case "silence":
			log.Printf("SIP Silence on track %s (duration: %dms)", event.TrackID, event.Duration)

		case "dtmf":
			log.Printf("SIP DTMF: %s on track %s", event.Digit, event.TrackID)
			handleSIPDTMF(conn, event.Digit)

		case "error":
			log.Printf("SIP Error from %s: %s (code: %d)", event.Sender, event.Error, event.Code)
		}
	})

	// Set up outbound SIP call configuration
	sipCallOption := &rustpbx.CallOption{
		Caller: "sip:sdk-client@example.com",
		Callee: "sip:service@example.com",
		Codec:  rustpbx.CodecPCMU,
		SIP: &rustpbx.SipOption{
			Username: "sdk-client",
			Password: "client-password",
			Realm:    "example.com",
			Headers: map[string]string{
				"X-Client-Type":    "go-sdk",
				"X-Call-Purpose":   "demonstration",
				"X-Service-Level":  "premium",
				"Contact":          "sip:sdk-client@192.168.1.100:5060",
				"Allow":            "INVITE,ACK,CANCEL,BYE,REFER,OPTIONS,INFO",
				"Supported":        "replaces,timer",
			},
		},
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
			RecorderFile: "/tmp/sip-outbound-" + time.Now().Format("20060102-150405") + ".wav",
			SampleRate:   16000,
			PTime:        "20ms",
		},
		HandshakeTimeout: "30s",
		EnableIPv6:       false,
		Extra: map[string]interface{}{
			"sip_integration": true,
			"protocol_version": "SIP/2.0",
			"transport": "UDP",
		},
	}

	// Initiate SIP call
	log.Println("Initiating SIP call...")
	if err := conn.Invite(sipCallOption); err != nil {
		log.Fatal("Failed to send SIP invite:", err)
	}

	// Demonstrate advanced SIP features after call setup
	time.AfterFunc(10*time.Second, func() {
		if callActive {
			log.Println("Demonstrating SIP call features...")
			
			// Example: Send custom SIP INFO
			conn.SendRawCommand(map[string]interface{}{
				"command": "sip_info",
				"headers": map[string]string{
					"Content-Type": "application/dtmf-relay",
					"Content-Length": "0",
				},
			})
			
			// Example: SIP-specific audio playback
			conn.TTSSimple("This demonstrates SIP protocol integration with advanced telephony features.")
		}
	})

	// Wait for interrupt signal or timeout
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
		log.Println("Received interrupt signal, ending SIP call...")
		if callActive {
			conn.TTSSimple("Thank you for using our SIP-based service. Goodbye!")
			time.Sleep(3 * time.Second)
			conn.Hangup("normal_clearing", "caller")
		}
		time.Sleep(2 * time.Second)
	case <-time.After(180 * time.Second): // 3 minutes
		log.Println("SIP demo timeout, ending call...")
		if callActive {
			conn.TTSSimple("Demo session timeout. Thank you for trying our SIP integration!")
			time.Sleep(3 * time.Second)
			conn.Hangup("normal_clearing", "caller")
		}
	}

	log.Println("SIP integration demo completed")
}

// processSIPCommand processes SIP-specific user commands
func processSIPCommand(input string) string {
	input = strings.ToLower(input)

	switch {
	case strings.Contains(input, "sip") && strings.Contains(input, "status"):
		return "SIP connection is active. Protocol version SIP/2.0, transport UDP, codec PCMU."

	case strings.Contains(input, "transfer") || strings.Contains(input, "forward"):
		return "I can transfer your call using SIP REFER method. Please specify the destination."

	case strings.Contains(input, "hold") || strings.Contains(input, "wait"):
		return "Placing you on hold using SIP re-INVITE. Please wait while I assist you."

	case strings.Contains(input, "conference") || strings.Contains(input, "join"):
		return "I can set up a SIP conference call. This would use SIP multicast capabilities."

	case strings.Contains(input, "quality") || strings.Contains(input, "audio"):
		return "Audio quality is optimized for SIP. Current codec is PCMU at 16kHz sample rate."

	case strings.Contains(input, "record") || strings.Contains(input, "save"):
		return "This call is being recorded in WAV format as configured in the SIP session."

	case strings.Contains(input, "dtmf") || strings.Contains(input, "digit"):
		return "You can send DTMF tones using your phone keypad. They will be processed via SIP INFO messages."

	case strings.Contains(input, "network") || strings.Contains(input, "connection"):
		return "SIP signaling is active. RTP media streams are established for audio transmission."

	case strings.Contains(input, "help") || strings.Contains(input, "what can"):
		return "I can help with SIP call features including transfer, hold, conference, and DTMF processing. What would you like to do?"

	case strings.Contains(input, "bye") || strings.Contains(input, "goodbye"):
		return "Thank you for using our SIP service. Initiating SIP BYE to end the session."

	default:
		return "I understand: " + input + ". This is processed through our SIP-based voice system. How else can I help?"
	}
}

// handleSIPDTMF processes DTMF tones in SIP context
func handleSIPDTMF(conn *rustpbx.Connection, digit string) {
	log.Printf("Processing SIP DTMF: %s", digit)

	switch digit {
	case "1":
		conn.TTSSimple("DTMF 1 received via SIP INFO. Connecting to customer service.")
		// Implement SIP transfer logic
		
	case "2":
		conn.TTSSimple("DTMF 2 received. Activating SIP call recording.")
		
	case "3":
		conn.TTSSimple("DTMF 3 received. Joining SIP conference bridge.")
		
	case "4":
		conn.TTSSimple("DTMF 4 received. Placing call on SIP hold with music.")
		conn.Play("https://example.com/sip-hold-music.wav", false)
		
	case "5":
		conn.TTSSimple("DTMF 5 received. Resuming SIP call from hold.")
		conn.Resume()
		
	case "0":
		conn.TTSSimple("DTMF 0 received. Returning to SIP main menu.")
		
	case "*":
		conn.TTSSimple("Star key received. Accessing SIP advanced features.")
		
	case "#":
		conn.TTSSimple("Pound key received. Confirming SIP operation.")
		
	default:
		conn.TTSSimple("DTMF " + digit + " received via SIP signaling. Please try a different option.")
	}
}