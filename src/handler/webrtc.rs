use super::middleware::clientaddr::ClientAddr;
use crate::app::AppState;
use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use reqwest;
use serde::{Deserialize, Serialize};
use std::{env, time::Instant};
use tracing::{error, info};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IceServer {
    urls: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    username: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    credential: Option<String>,
}

// WebRTC SDP structures
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SdpSessionDescription {
    #[serde(rename = "type")]
    pub sdp_type: String,
    pub sdp: String,
}

#[derive(Debug, Deserialize)]
pub struct SdpOfferRequest {
    pub sdp: SdpSessionDescription,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub session_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<serde_json::Value>,
}

#[derive(Debug, Serialize)]
pub struct SdpAnswerResponse {
    pub sdp: SdpSessionDescription,
    pub session_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ice_candidates: Option<Vec<IceCandidate>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IceCandidate {
    pub candidate: String,
    #[serde(rename = "sdpMLineIndex")]
    pub sdp_m_line_index: Option<u16>,
    #[serde(rename = "sdpMid")]
    pub sdp_mid: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
    pub code: u16,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub session_id: Option<String>,
}

/// Handle SDP offer and generate SDP answer
pub(crate) async fn handle_sdp_offer(
    State(state): State<AppState>,
    Json(request): Json<SdpOfferRequest>,
) -> Response {
    info!("Received SDP offer: type={}, session_id={:?}", 
          request.sdp.sdp_type, request.session_id);
    
    // Validate SDP offer
    if request.sdp.sdp_type != "offer" {
        let error = ErrorResponse {
            error: "Invalid SDP type, expected 'offer'".to_string(),
            code: 400,
            session_id: request.session_id,
        };
        return (StatusCode::BAD_REQUEST, Json(error)).into_response();
    }

    if request.sdp.sdp.trim().is_empty() {
        let error = ErrorResponse {
            error: "SDP content cannot be empty".to_string(),
            code: 400,
            session_id: request.session_id,
        };
        return (StatusCode::BAD_REQUEST, Json(error)).into_response();
    }

    // Generate unique session ID if not provided
    let session_id = request.session_id.unwrap_or_else(|| Uuid::new_v4().to_string());
    
    match process_sdp_offer(&state, &request.sdp.sdp, &session_id).await {
        Ok(answer_sdp) => {
            let response = SdpAnswerResponse {
                sdp: SdpSessionDescription {
                    sdp_type: "answer".to_string(),
                    sdp: answer_sdp,
                },
                session_id: session_id.clone(),
                ice_candidates: None, // ICE candidates can be gathered separately
                metadata: request.metadata,
            };
            
            info!("Generated SDP answer for session: {}", session_id);
            (StatusCode::OK, Json(response)).into_response()
        }
        Err(e) => {
            error!("Failed to process SDP offer for session {}: {}", session_id, e);
            let error = ErrorResponse {
                error: format!("Failed to process SDP offer: {}", e),
                code: 500,
                session_id: Some(session_id),
            };
            (StatusCode::INTERNAL_SERVER_ERROR, Json(error)).into_response()
        }
    }
}

/// Process SDP offer and create WebRTC track to generate answer
async fn process_sdp_offer(
    state: &AppState,
    offer_sdp: &str,
    session_id: &str,
) -> Result<String, anyhow::Error> {
    use crate::media::{
        stream::MediaStreamBuilder,
        track::{webrtc::WebrtcTrack, TrackConfig},
    };
    use tokio_util::sync::CancellationToken;
    use std::sync::Arc;
    
    // Create cancellation token for this session
    let cancel_token = CancellationToken::new();
    
    // Create event sender
    let event_sender = crate::event::create_event_sender();
    
    // Create media stream
    let media_stream_builder = MediaStreamBuilder::new(event_sender.clone());
    let media_stream = Arc::new(
        media_stream_builder
            .with_cancel_token(cancel_token.clone())
            .build(),
    );
    
    // Create WebRTC track
    let track_id = format!("webrtc-{}", session_id);
    let mut webrtc_track = WebrtcTrack::new(
        cancel_token.child_token(),
        track_id.clone(),
        TrackConfig::default(),
    );
    
    // Setup WebRTC track with the offer SDP
    let answer = webrtc_track.setup_webrtc_track(offer_sdp.to_string(), None).await?;
    
    // Store the track in media stream for processing
    media_stream.update_track(Box::new(webrtc_track)).await;
    
    // Start media stream in background
    let media_stream_clone = media_stream.clone();
    tokio::spawn(async move {
        if let Err(e) = media_stream_clone.serve().await {
            error!("Media stream error for session {}: {}", session_id, e);
        }
    });
    
    // Store session info in app state for later reference
    // Note: You might want to add a session storage mechanism to AppState
    
    Ok(answer.sdp)
}

/// Handle ICE candidate exchange
#[derive(Debug, Deserialize)]
pub struct IceCandidateRequest {
    pub session_id: String,
    pub candidate: IceCandidate,
}

#[derive(Debug, Serialize)]
pub struct IceCandidateResponse {
    pub session_id: String,
    pub status: String,
}

pub(crate) async fn handle_ice_candidate(
    State(_state): State<AppState>,
    Json(request): Json<IceCandidateRequest>,
) -> Response {
    info!("Received ICE candidate for session: {}", request.session_id);
    
    // In a real implementation, you would:
    // 1. Find the WebRTC peer connection for this session
    // 2. Add the ICE candidate to the peer connection
    // 3. Handle any errors
    
    // For now, we'll just acknowledge receipt
    let response = IceCandidateResponse {
        session_id: request.session_id,
        status: "received".to_string(),
    };
    
    (StatusCode::OK, Json(response)).into_response()
}

/// Close WebRTC session
#[derive(Debug, Deserialize)]
pub struct CloseSessionRequest {
    pub session_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reason: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct CloseSessionResponse {
    pub session_id: String,
    pub status: String,
}

pub(crate) async fn handle_close_session(
    State(_state): State<AppState>,
    Json(request): Json<CloseSessionRequest>,
) -> Response {
    info!("Closing WebRTC session: {} (reason: {:?})", 
          request.session_id, request.reason);
    
    // In a real implementation, you would:
    // 1. Find the session and associated resources
    // 2. Cancel any running tasks
    // 3. Close peer connection
    // 4. Clean up media streams
    
    let response = CloseSessionResponse {
        session_id: request.session_id,
        status: "closed".to_string(),
    };
    
    (StatusCode::OK, Json(response)).into_response()
}

// Keep the existing ICE server function
pub(crate) async fn get_iceservers(
    client_ip: ClientAddr,
    State(state): State<AppState>,
) -> Response {
    let rs_token = env::var("RESTSEND_TOKEN").unwrap_or_default();
    let default_ice_servers = state.config.ice_servers.as_ref();
    if rs_token.is_empty() {
        if let Some(ice_servers) = default_ice_servers {
            return Json(ice_servers).into_response();
        }
        return Json(vec![IceServer {
            urls: vec!["stun:restsend.com:3478".to_string()],
            username: None,
            credential: None,
        }])
        .into_response();
    }

    let start_time = Instant::now();
    let user_id = ""; // TODO: Get user ID from state if needed
    let timeout = std::time::Duration::from_secs(5);
    let url = format!(
        "https://restsend.com/api/iceservers?token={}&user={}&client={}",
        rs_token,
        user_id,
        client_ip.ip().to_string()
    );

    // Create a reqwest client with proper timeout
    let client = match reqwest::Client::builder().timeout(timeout).build() {
        Ok(client) => client,
        Err(e) => {
            error!("voiceserver: failed to create HTTP client: {}", e);
            return Json(default_ice_servers).into_response();
        }
    };

    let response = match client.get(&url).send().await {
        Ok(response) => response,
        Err(e) => {
            error!("voiceserver: alloc ice servers failed: {}", e);
            return Json(default_ice_servers).into_response();
        }
    };

    if !response.status().is_success() {
        error!(
            "voiceserver: ice servers request failed with status: {}",
            response.status()
        );
        return Json(default_ice_servers).into_response();
    }

    // Parse the response JSON
    match response.json::<Vec<IceServer>>().await {
        Ok(ice_servers) => {
            info!(
                "voiceserver: get ice servers - duration: {:?}, count: {}, userId: {}, clientIP: {}",
                start_time.elapsed(),
                ice_servers.len(),
                user_id,
                client_ip
            );
            Json(ice_servers).into_response()
        }
        Err(e) => {
            error!("voiceserver: decode ice servers failed: {}", e);
            Json(default_ice_servers).into_response()
        }
    }
}
