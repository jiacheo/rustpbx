package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rustpbx/go-sdk/rustpbx"
)

// ConversationHistory tracks the conversation for context
type ConversationHistory struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletion represents OpenAI-compatible chat completion request
type ChatCompletion struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatResponse represents the response from chat completion
type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func main() {
	// Create a new RustPBX client
	client := rustpbx.NewClient("ws://localhost:8080")

	// Connect to the call endpoint
	ctx := context.Background()
	conn, err := client.ConnectCall(ctx, &rustpbx.ConnectionOptions{
		SessionID: "ai-assistant-demo",
		Dump:      true,
	})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	log.Println("Connected to RustPBX AI Voice Assistant")

	// Initialize conversation history
	conversation := &ConversationHistory{
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful AI voice assistant. Keep responses concise and conversational, suitable for voice interaction. Be friendly and helpful.",
			},
		},
	}

	// Track call state
	callActive := false

	// Set up comprehensive event handlers
	conn.OnEvent(func(event *rustpbx.Event) {
		log.Printf("AI Assistant Event: %s", event.Event)

		switch event.Event {
		case "incoming":
			log.Printf("Incoming call from %s to %s", event.Caller, event.Callee)

			// Auto-accept with full AI configuration
			acceptOption := &rustpbx.CallOption{
				Codec: rustpbx.CodecPCMU,
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
					RecorderFile: "/tmp/ai-assistant-" + time.Now().Format("20060102-150405") + ".wav",
					SampleRate:   16000,
					PTime:        "20ms",
				},
			}

			if err := conn.Accept(acceptOption); err != nil {
				log.Printf("Failed to accept call: %v", err)
			}

		case "answer":
			log.Println("AI Assistant call connected")
			callActive = true

			// Send welcome message
			welcomeMsg := "Hello! I'm your AI voice assistant. How can I help you today?"
			if err := conn.TTSSimple(welcomeMsg); err != nil {
				log.Printf("Failed to send welcome message: %v", err)
			}

			// Add assistant's welcome to conversation history
			conn.History("assistant", welcomeMsg)
			conversation.Messages = append(conversation.Messages, Message{
				Role:    "assistant",
				Content: welcomeMsg,
			})

		case "ringing":
			log.Println("AI Assistant call is ringing")

		case "hangup":
			log.Printf("AI Assistant call ended: %s (initiated by %s)", event.Reason, event.Initiator)
			callActive = false
			
			// Save conversation summary
			saveFinalSummary(conversation)

		case "asrFinal":
			if !callActive {
				return
			}
			
			userInput := strings.TrimSpace(event.Text)
			log.Printf("User said: %s", userInput)

			// Add user input to conversation history
			conn.History("user", userInput)
			conversation.Messages = append(conversation.Messages, Message{
				Role:    "user",
				Content: userInput,
			})

			// Check for special commands
			if handleSpecialCommands(conn, userInput) {
				return
			}

			// Get AI response
			aiResponse, err := getAIResponse(client, conversation)
			if err != nil {
				log.Printf("Failed to get AI response: %v", err)
				fallbackResponse := "I'm sorry, I'm having trouble processing that right now. Could you please repeat?"
				conn.TTSSimple(fallbackResponse)
				return
			}

			// Send AI response via TTS
			if err := conn.TTS(aiResponse, "101002", "", &rustpbx.TTSOptions{
				Streaming:   true,
				AutoHangup:  false,
				EndOfStream: false,
			}); err != nil {
				log.Printf("Failed to send AI response via TTS: %v", err)
			}

			// Add AI response to conversation history
			conn.History("assistant", aiResponse)
			conversation.Messages = append(conversation.Messages, Message{
				Role:    "assistant",
				Content: aiResponse,
			})

		case "asrDelta":
			// Log partial transcription for debugging
			log.Printf("Partial speech: %s", event.Text)

		case "speaking":
			log.Printf("Speech activity detected on track %s", event.TrackID)

		case "silence":
			log.Printf("Silence detected on track %s (duration: %dms)", event.TrackID, event.Duration)
			
			// If silence is too long, prompt user
			if event.Duration > 10000 && callActive { // 10 seconds
				conn.TTSSimple("Are you still there? I'm here to help if you need anything.")
			}

		case "dtmf":
			log.Printf("DTMF digit: %s", event.Digit)
			handleDTMFCommands(conn, event.Digit, conversation)

		case "error":
			log.Printf("AI Assistant Error from %s: %s (code: %d)", event.Sender, event.Error, event.Code)
			if callActive {
				conn.TTSSimple("I encountered an error. Let me try to help you differently.")
			}
		}
	})

	// Set up AI assistant call configuration
	aiAssistantOption := &rustpbx.CallOption{
		Caller: "ai-assistant@example.com",
		Callee: "user@example.com",
		Codec:  rustpbx.CodecPCMU,
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
			RecorderFile: "/tmp/ai-assistant-outbound-" + time.Now().Format("20060102-150405") + ".wav",
			SampleRate:   16000,
			PTime:        "20ms",
		},
		Extra: map[string]interface{}{
			"ai_assistant": true,
			"version":      "1.0",
		},
	}

	// Start AI assistant
	log.Println("Starting AI Voice Assistant...")
	if err := conn.Invite(aiAssistantOption); err != nil {
		log.Fatal("Failed to start AI assistant:", err)
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
		log.Println("Received interrupt signal, shutting down AI assistant...")
		if callActive {
			conn.TTSSimple("Thank you for using the AI assistant. Goodbye!")
			time.Sleep(3 * time.Second)
			conn.HangupSimple()
		}
	case <-time.After(300 * time.Second): // 5 minutes
		log.Println("AI Assistant session timeout")
		if callActive {
			conn.TTSSimple("Our session time is up. Thank you for using the AI assistant. Goodbye!")
			time.Sleep(3 * time.Second)
			conn.HangupSimple()
		}
	}

	log.Println("AI Voice Assistant demo completed")
}

// handleSpecialCommands processes special voice commands
func handleSpecialCommands(conn *rustpbx.Connection, input string) bool {
	input = strings.ToLower(input)

	switch {
	case strings.Contains(input, "goodbye") || strings.Contains(input, "bye") || strings.Contains(input, "end call"):
		conn.TTSSimple("Thank you for using the AI assistant. Have a wonderful day! Goodbye!")
		time.AfterFunc(3*time.Second, func() {
			conn.HangupSimple()
		})
		return true

	case strings.Contains(input, "transfer") || strings.Contains(input, "human agent"):
		conn.TTSSimple("I'll transfer you to a human agent now.")
		// In a real scenario, implement call transfer logic
		conn.Refer("sip:agent@example.com", &rustpbx.ReferOption{
			Timeout:    30,
			AutoHangup: true,
		})
		return true

	case strings.Contains(input, "mute") || strings.Contains(input, "quiet"):
		conn.TTSSimple("I'll be quiet now. Say 'unmute' when you want me to respond again.")
		// Implementation would involve setting a mute flag
		return true

	case strings.Contains(input, "repeat") || strings.Contains(input, "say that again"):
		conn.TTSSimple("I'm sorry, let me repeat my last response.")
		// Implementation would involve repeating the last response
		return true
	}

	return false
}

// handleDTMFCommands processes DTMF commands for the AI assistant
func handleDTMFCommands(conn *rustpbx.Connection, digit string, conversation *ConversationHistory) {
	switch digit {
	case "1":
		conn.TTSSimple("Switching to customer service mode.")
		// Add system message to change behavior
		conversation.Messages = append(conversation.Messages, Message{
			Role:    "system",
			Content: "You are now in customer service mode. Be extra helpful and professional.",
		})

	case "2":
		conn.TTSSimple("Switching to technical support mode.")
		conversation.Messages = append(conversation.Messages, Message{
			Role:    "system",
			Content: "You are now in technical support mode. Focus on troubleshooting and technical solutions.",
		})

	case "3":
		conn.TTSSimple("Playing hold music while I process your request.")
		conn.Play("https://example.com/hold-music.wav", false)

	case "9":
		conn.TTSSimple("Ending our conversation. Thank you!")
		time.AfterFunc(2*time.Second, func() {
			conn.HangupSimple()
		})

	case "0":
		conn.TTSSimple("Returning to main assistant mode.")
		// Reset to original system prompt
		for i, msg := range conversation.Messages {
			if msg.Role == "system" && i == 0 {
				conversation.Messages[0] = Message{
					Role:    "system",
					Content: "You are a helpful AI voice assistant. Keep responses concise and conversational, suitable for voice interaction. Be friendly and helpful.",
				}
				break
			}
		}

	default:
		conn.TTSSimple(fmt.Sprintf("You pressed %s. Press 1 for customer service, 2 for technical support, or 9 to end the call.", digit))
	}
}

// getAIResponse calls the LLM to get an AI response
func getAIResponse(client *rustpbx.Client, conversation *ConversationHistory) (string, error) {
	// Prepare chat completion request
	request := ChatCompletion{
		Model:    "gpt-3.5-turbo",
		Messages: conversation.Messages,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request to LLM proxy
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := client.ProxyLLMRequest(ctx, "chat/completions", "POST", bytes.NewReader(requestBody), headers)
	if err != nil {
		return "", fmt.Errorf("failed to call LLM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResponse ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		return "", fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(chatResponse.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return chatResponse.Choices[0].Message.Content, nil
}

// saveFinalSummary saves a summary of the conversation
func saveFinalSummary(conversation *ConversationHistory) {
	log.Println("Conversation Summary:")
	log.Printf("Total messages: %d", len(conversation.Messages))
	
	userMessages := 0
	assistantMessages := 0
	
	for _, msg := range conversation.Messages {
		switch msg.Role {
		case "user":
			userMessages++
		case "assistant":
			assistantMessages++
		}
	}
	
	log.Printf("User messages: %d", userMessages)
	log.Printf("Assistant messages: %d", assistantMessages)
	log.Println("Conversation ended successfully")
}