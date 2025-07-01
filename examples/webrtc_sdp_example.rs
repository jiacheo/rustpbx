use anyhow::Result;
use rustpbx::webrtc_client::{WebRtcClient, IceCandidate};
use serde_json::json;
use tokio::time::{sleep, Duration};
use tracing::{info, error, Level};

// Example SDP offer (simplified for demonstration)
const SAMPLE_SDP_OFFER: &str = r#"v=0
o=- 4611731400430051336 2 IN IP4 127.0.0.1
s=-
t=0 0
a=group:BUNDLE 0
a=extmap-allow-mixed
a=msid-semantic: WMS
m=audio 9 UDP/TLS/RTP/SAVPF 111 103 104 9 0 8 106 105 13 110 112 113 126
c=IN IP4 0.0.0.0
a=rtcp:9 IN IP4 0.0.0.0
a=ice-ufrag:4ZcD
a=ice-pwd:2/1muCWoOi3uHTiCSIWszae17p
a=ice-options:trickle
a=fingerprint:sha-256 75:74:5A:A6:A4:E5:52:F4:A7:67:4C:01:C7:EE:91:3F:21:3D:A2:E3:53:7B:6F:30:86:F2:30:FF:A6:22:D2:04
a=setup:actpass
a=mid:0
a=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level
a=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time
a=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01
a=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid
a=extmap:5 urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id
a=extmap:6 urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id
a=sendrecv
a=msid:- {track-id}
a=rtcp-mux
a=rtpmap:111 opus/48000/2
a=rtcp-fb:111 transport-cc
a=fmtp:111 minptime=10;useinbandfec=1
a=rtpmap:103 ISAC/16000
a=rtpmap:104 ISAC/32000
a=rtpmap:9 G722/8000
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=rtpmap:106 CN/32000
a=rtpmap:105 CN/16000
a=rtpmap:13 CN/8000
a=rtpmap:110 telephone-event/48000
a=rtpmap:112 telephone-event/32000
a=rtpmap:113 telephone-event/16000
a=rtpmap:126 telephone-event/8000
a=ssrc:1001 cname:4BOcZWCeHcBF9vhd
a=ssrc:1001 msid:- {track-id}
"#;

/// Demonstrate basic SDP offer/answer exchange
async fn basic_sdp_exchange() -> Result<()> {
    info!("=== Basic SDP Exchange Example ===");
    
    // Create WebRTC client
    let client = WebRtcClient::new("http://localhost:8080".to_string());
    
    // Send offer and receive answer
    let (session_id, answer_sdp) = client.exchange_sdp(SAMPLE_SDP_OFFER.to_string()).await?;
    
    info!("Session ID: {}", session_id);
    info!("Received SDP answer (first 200 chars):\n{}", 
          &answer_sdp.chars().take(200).collect::<String>());
    
    // Close the session
    client.close_session(session_id, Some("Example completed".to_string())).await?;
    
    Ok(())
}

/// Demonstrate SDP exchange with session tracking
async fn tracked_sdp_exchange() -> Result<()> {
    info!("=== Tracked SDP Exchange Example ===");
    
    let client = WebRtcClient::new("http://localhost:8080".to_string());
    let session_id = "my-custom-session-123".to_string();
    
    // Send offer with specific session ID
    let answer_sdp = client.exchange_sdp_with_session(
        SAMPLE_SDP_OFFER.to_string(),
        session_id.clone()
    ).await?;
    
    info!("Using custom session ID: {}", session_id);
    info!("Answer SDP length: {} bytes", answer_sdp.len());
    
    // Simulate some processing time
    sleep(Duration::from_secs(2)).await;
    
    // Close the session
    client.close_session(session_id, Some("Tracked session completed".to_string())).await?;
    
    Ok(())
}

/// Demonstrate complete session setup with ICE candidates
async fn complete_session_setup() -> Result<()> {
    info!("=== Complete Session Setup Example ===");
    
    let client = WebRtcClient::new("http://localhost:8080".to_string());
    
    // Create sample ICE candidates
    let ice_candidates = vec![
        IceCandidate {
            candidate: "candidate:1 1 UDP 2013266431 192.168.1.100 54400 typ host".to_string(),
            sdp_m_line_index: Some(0),
            sdp_mid: Some("0".to_string()),
        },
        IceCandidate {
            candidate: "candidate:2 1 TCP 1019216383 192.168.1.100 9 typ host tcptype active".to_string(),
            sdp_m_line_index: Some(0),
            sdp_mid: Some("0".to_string()),
        },
    ];
    
    // Setup complete session
    let (session_id, answer_sdp) = client.setup_session(
        SAMPLE_SDP_OFFER.to_string(),
        ice_candidates
    ).await?;
    
    info!("Complete session setup finished");
    info!("Session ID: {}", session_id);
    info!("Answer SDP received: {} bytes", answer_sdp.len());
    
    // Simulate session activity
    sleep(Duration::from_secs(3)).await;
    
    // Close the session
    client.close_session(session_id, Some("Complete session finished".to_string())).await?;
    
    Ok(())
}

/// Demonstrate advanced usage with metadata
async fn advanced_usage_with_metadata() -> Result<()> {
    info!("=== Advanced Usage with Metadata ===");
    
    let client = WebRtcClient::new("http://localhost:8080".to_string());
    
    // Create metadata
    let metadata = json!({
        "user_id": "user123",
        "call_type": "voice",
        "priority": "high",
        "room_id": "conference-room-1",
        "timestamp": chrono::Utc::now().to_rfc3339()
    });
    
    // Send offer with metadata
    let answer = client.send_offer(
        SAMPLE_SDP_OFFER.to_string(),
        Some("metadata-session".to_string()),
        Some(metadata.clone())
    ).await?;
    
    info!("Session with metadata created: {}", answer.session_id);
    if let Some(returned_metadata) = answer.metadata {
        info!("Returned metadata: {}", returned_metadata);
    }
    
    // Send an ICE candidate
    let ice_candidate = IceCandidate {
        candidate: "candidate:1 1 UDP 2013266431 10.0.0.1 54400 typ host".to_string(),
        sdp_m_line_index: Some(0),
        sdp_mid: Some("0".to_string()),
    };
    
    let ice_response = client.send_ice_candidate(answer.session_id.clone(), ice_candidate).await?;
    info!("ICE candidate response: {:?}", ice_response);
    
    // Close the session
    client.close_session(answer.session_id, Some("Advanced example completed".to_string())).await?;
    
    Ok(())
}

/// Demonstrate ICE server retrieval
async fn get_ice_servers_example() -> Result<()> {
    info!("=== ICE Servers Example ===");
    
    let client = WebRtcClient::new("http://localhost:8080".to_string());
    
    // Get ICE servers
    let ice_servers = client.get_ice_servers().await?;
    
    info!("Retrieved {} ICE servers:", ice_servers.len());
    for (i, server) in ice_servers.iter().enumerate() {
        info!("  Server {}: {:?}", i + 1, server.urls);
        if let Some(username) = &server.username {
            info!("    Username: {}", username);
        }
        if server.credential.is_some() {
            info!("    Has credential: yes");
        }
    }
    
    Ok(())
}

/// Demonstrate error handling
async fn error_handling_example() -> Result<()> {
    info!("=== Error Handling Example ===");
    
    let client = WebRtcClient::new("http://localhost:8080".to_string());
    
    // Try to send invalid SDP
    let invalid_sdp = "this is not a valid SDP";
    
    match client.exchange_sdp(invalid_sdp.to_string()).await {
        Ok(_) => info!("Unexpected success with invalid SDP"),
        Err(e) => info!("Expected error with invalid SDP: {}", e),
    }
    
    // Try to send empty SDP
    match client.exchange_sdp("".to_string()).await {
        Ok(_) => info!("Unexpected success with empty SDP"),
        Err(e) => info!("Expected error with empty SDP: {}", e),
    }
    
    Ok(())
}

/// Run multiple concurrent sessions
async fn concurrent_sessions_example() -> Result<()> {
    info!("=== Concurrent Sessions Example ===");
    
    let client = WebRtcClient::new("http://localhost:8080".to_string());
    
    // Create multiple sessions concurrently
    let mut handles = vec![];
    
    for i in 0..3 {
        let client_clone = WebRtcClient::new("http://localhost:8080".to_string());
        let session_id = format!("concurrent-session-{}", i);
        
        let handle = tokio::spawn(async move {
            let result = client_clone.exchange_sdp_with_session(
                SAMPLE_SDP_OFFER.to_string(),
                session_id.clone()
            ).await;
            
            match result {
                Ok(answer) => {
                    info!("Session {} completed, answer length: {}", session_id, answer.len());
                    // Close session
                    if let Err(e) = client_clone.close_session(session_id.clone(), Some("Concurrent session done".to_string())).await {
                        error!("Failed to close session {}: {}", session_id, e);
                    }
                    Ok(())
                }
                Err(e) => {
                    error!("Session {} failed: {}", session_id, e);
                    Err(e)
                }
            }
        });
        
        handles.push(handle);
    }
    
    // Wait for all sessions to complete
    for handle in handles {
        if let Err(e) = handle.await? {
            error!("Concurrent session error: {}", e);
        }
    }
    
    info!("All concurrent sessions completed");
    Ok(())
}

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_max_level(Level::INFO)
        .with_target(false)
        .with_thread_ids(true)
        .with_file(true)
        .with_line_number(true)
        .init();

    info!("Starting WebRTC SDP Examples");
    info!("Make sure RustPBX server is running on http://localhost:8080");
    
    // Wait a moment for user to read the message
    sleep(Duration::from_secs(2)).await;

    // Run examples
    let examples = vec![
        ("Basic SDP Exchange", basic_sdp_exchange()),
        ("Tracked SDP Exchange", tracked_sdp_exchange()),
        ("Complete Session Setup", complete_session_setup()),
        ("Advanced Usage with Metadata", advanced_usage_with_metadata()),
        ("ICE Servers Retrieval", get_ice_servers_example()),
        ("Error Handling", error_handling_example()),
        ("Concurrent Sessions", concurrent_sessions_example()),
    ];

    for (name, example_fn) in examples {
        info!("\n" + "=".repeat(50).as_str());
        info!("Running example: {}", name);
        info!("=".repeat(50));
        
        match example_fn.await {
            Ok(_) => info!("✅ {} completed successfully", name),
            Err(e) => error!("❌ {} failed: {}", name, e),
        }
        
        // Wait between examples
        sleep(Duration::from_secs(1)).await;
    }
    
    info!("\n🎉 All WebRTC SDP examples completed!");
    Ok(())
}