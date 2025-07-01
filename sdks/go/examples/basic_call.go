package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rustpbx/go-sdk/rustpbx"
)

func main() {
	// Create a new RustPBX client
	client := rustpbx.NewClient("ws://localhost:8080")

	// Connect to the WebSocket endpoint
	ctx := context.Background()
	conn, err := client.ConnectCall(ctx, &rustpbx.ConnectionOptions{
		SessionID: "basic-call-example",
		Dump:      true,
	})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	log.Println("Connected to RustPBX WebSocket")

	// Set up event handlers
	conn.OnEvent(func(event *rustpbx.Event) {
		log.Printf("Received event: %s", event.Event)

		switch event.Event {
		case "incoming":
			log.Printf("Incoming call from %s to %s", event.Caller, event.Callee)
			// Automatically accept the call
			acceptOption := &rustpbx.CallOption{
				Codec: rustpbx.CodecPCMU,
				TTS: &rustpbx.SynthesisOption{
					Provider:   rustpbx.ProviderTencent,
					Speaker:    "101002",
					SampleRate: 16000,
				},
			}
			if err := conn.Accept(acceptOption); err != nil {
				log.Printf("Failed to accept call: %v", err)
			}

		case "answer":
			log.Println("Call answered")
			// Send a greeting message
			if err := conn.TTSSimple("Hello! Welcome to RustPBX. How can I help you today?"); err != nil {
				log.Printf("Failed to send TTS: %v", err)
			}

		case "ringing":
			log.Println("Call is ringing")

		case "hangup":
			log.Printf("Call ended: %s (initiated by %s)", event.Reason, event.Initiator)

		case "asrFinal":
			log.Printf("Speech recognized: %s", event.Text)
			// Echo back what was said
			response := "I heard you say: " + event.Text
			if err := conn.TTSSimple(response); err != nil {
				log.Printf("Failed to send TTS response: %v", err)
			}

		case "speaking":
			log.Printf("Speaking detected on track %s", event.TrackID)

		case "silence":
			log.Printf("Silence detected on track %s (duration: %dms)", event.TrackID, event.Duration)

		case "dtmf":
			log.Printf("DTMF digit pressed: %s", event.Digit)

		case "error":
			log.Printf("Error from %s: %s (code: %d)", event.Sender, event.Error, event.Code)
		}
	})

	// Set up basic call option for outgoing calls
	callOption := &rustpbx.CallOption{
		Caller: "sdk-user@example.com",
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
		VAD: &rustpbx.VADOption{
			Type:           rustpbx.VADTypeWebRTC,
			Aggressiveness: 3,
		},
	}

	// Send invite to start a call
	log.Println("Sending call invitation...")
	if err := conn.Invite(callOption); err != nil {
		log.Fatal("Failed to send invite:", err)
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
		log.Println("Received interrupt signal, closing connection...")
		// Send hangup before closing
		conn.HangupSimple()
		time.Sleep(1 * time.Second)
	case <-time.After(60 * time.Second):
		log.Println("Example timeout, closing connection...")
		conn.HangupSimple()
	}

	log.Println("Example completed")
}