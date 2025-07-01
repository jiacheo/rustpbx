use anyhow::Result;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::time::Duration;
use tracing::{error, info};

/// WebRTC SDP Session Description
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SdpSessionDescription {
    #[serde(rename = "type")]
    pub sdp_type: String,
    pub sdp: String,
}

/// WebRTC ICE Candidate
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IceCandidate {
    pub candidate: String,
    #[serde(rename = "sdpMLineIndex")]
    pub sdp_m_line_index: Option<u16>,
    #[serde(rename = "sdpMid")]
    pub sdp_mid: Option<String>,
}

/// SDP Offer Request
#[derive(Debug, Serialize)]
pub struct SdpOfferRequest {
    pub sdp: SdpSessionDescription,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub session_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<Value>,
}

/// SDP Answer Response
#[derive(Debug, Deserialize)]
pub struct SdpAnswerResponse {
    pub sdp: SdpSessionDescription,
    pub session_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ice_candidates: Option<Vec<IceCandidate>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<Value>,
}

/// ICE Candidate Request
#[derive(Debug, Serialize)]
pub struct IceCandidateRequest {
    pub session_id: String,
    pub candidate: IceCandidate,
}

/// ICE Candidate Response
#[derive(Debug, Deserialize)]
pub struct IceCandidateResponse {
    pub session_id: String,
    pub status: String,
}

/// Close Session Request
#[derive(Debug, Serialize)]
pub struct CloseSessionRequest {
    pub session_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reason: Option<String>,
}

/// Close Session Response
#[derive(Debug, Deserialize)]
pub struct CloseSessionResponse {
    pub session_id: String,
    pub status: String,
}

/// Error Response
#[derive(Debug, Deserialize)]
pub struct ErrorResponse {
    pub error: String,
    pub code: u16,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub session_id: Option<String>,
}

/// WebRTC Client for SDP offer/answer exchange
pub struct WebRtcClient {
    client: Client,
    base_url: String,
}

impl WebRtcClient {
    /// Create a new WebRTC client
    pub fn new(base_url: String) -> Self {
        let client = Client::builder()
            .timeout(Duration::from_secs(30))
            .build()
            .expect("Failed to create HTTP client");

        Self { client, base_url }
    }

    /// Create a new WebRTC client with custom HTTP client
    pub fn new_with_client(base_url: String, client: Client) -> Self {
        Self { client, base_url }
    }

    /// Send SDP offer and receive SDP answer
    pub async fn send_offer(
        &self,
        offer_sdp: String,
        session_id: Option<String>,
        metadata: Option<Value>,
    ) -> Result<SdpAnswerResponse> {
        let request = SdpOfferRequest {
            sdp: SdpSessionDescription {
                sdp_type: "offer".to_string(),
                sdp: offer_sdp,
            },
            session_id,
            metadata,
        };

        let url = format!("{}/webrtc/offer", self.base_url);
        info!("Sending SDP offer to: {}", url);

        let response = self
            .client
            .post(&url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            let answer: SdpAnswerResponse = response.json().await?;
            info!("Received SDP answer for session: {}", answer.session_id);
            Ok(answer)
        } else {
            let error: ErrorResponse = response.json().await?;
            Err(anyhow::anyhow!("Server error: {} (code: {})", error.error, error.code))
        }
    }

    /// Send ICE candidate
    pub async fn send_ice_candidate(
        &self,
        session_id: String,
        candidate: IceCandidate,
    ) -> Result<IceCandidateResponse> {
        let request = IceCandidateRequest {
            session_id: session_id.clone(),
            candidate,
        };

        let url = format!("{}/webrtc/ice-candidate", self.base_url);
        info!("Sending ICE candidate for session: {}", session_id);

        let response = self
            .client
            .post(&url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            let result: IceCandidateResponse = response.json().await?;
            Ok(result)
        } else {
            let error: ErrorResponse = response.json().await?;
            Err(anyhow::anyhow!("Server error: {} (code: {})", error.error, error.code))
        }
    }

    /// Close WebRTC session
    pub async fn close_session(
        &self,
        session_id: String,
        reason: Option<String>,
    ) -> Result<CloseSessionResponse> {
        let request = CloseSessionRequest {
            session_id: session_id.clone(),
            reason,
        };

        let url = format!("{}/webrtc/close", self.base_url);
        info!("Closing session: {}", session_id);

        let response = self
            .client
            .post(&url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            let result: CloseSessionResponse = response.json().await?;
            Ok(result)
        } else {
            let error: ErrorResponse = response.json().await?;
            Err(anyhow::anyhow!("Server error: {} (code: {})", error.error, error.code))
        }
    }

    /// Get ICE servers
    pub async fn get_ice_servers(&self) -> Result<Vec<IceServer>> {
        let url = format!("{}/iceservers", self.base_url);
        info!("Getting ICE servers from: {}", url);

        let response = self.client.get(&url).send().await?;

        if response.status().is_success() {
            let ice_servers: Vec<IceServer> = response.json().await?;
            Ok(ice_servers)
        } else {
            Err(anyhow::anyhow!("Failed to get ICE servers: HTTP {}", response.status()))
        }
    }
}

/// ICE Server configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IceServer {
    pub urls: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub username: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub credential: Option<String>,
}

/// Convenience functions for common WebRTC operations
impl WebRtcClient {
    /// Simple offer/answer exchange
    pub async fn exchange_sdp(
        &self,
        offer_sdp: String,
    ) -> Result<(String, String)> {
        let answer_response = self.send_offer(offer_sdp, None, None).await?;
        Ok((answer_response.session_id, answer_response.sdp.sdp))
    }

    /// Offer/answer exchange with session tracking
    pub async fn exchange_sdp_with_session(
        &self,
        offer_sdp: String,
        session_id: String,
    ) -> Result<String> {
        let answer_response = self.send_offer(offer_sdp, Some(session_id), None).await?;
        Ok(answer_response.sdp.sdp)
    }

    /// Complete WebRTC session setup
    pub async fn setup_session(
        &self,
        offer_sdp: String,
        ice_candidates: Vec<IceCandidate>,
    ) -> Result<(String, String)> {
        // Send offer and get answer
        let answer_response = self.send_offer(offer_sdp, None, None).await?;
        let session_id = answer_response.session_id.clone();
        
        // Send ICE candidates
        for candidate in ice_candidates {
            if let Err(e) = self.send_ice_candidate(session_id.clone(), candidate).await {
                error!("Failed to send ICE candidate: {}", e);
            }
        }
        
        Ok((session_id, answer_response.sdp.sdp))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use mockall::predicate::*;
    use serde_json::json;

    #[tokio::test]
    async fn test_webrtc_client_creation() {
        let client = WebRtcClient::new("http://localhost:8080".to_string());
        assert_eq!(client.base_url, "http://localhost:8080");
    }

    #[tokio::test]
    async fn test_sdp_serialization() {
        let offer_request = SdpOfferRequest {
            sdp: SdpSessionDescription {
                sdp_type: "offer".to_string(),
                sdp: "v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.1\r\n".to_string(),
            },
            session_id: Some("test-session".to_string()),
            metadata: Some(json!({"test": "data"})),
        };

        let json_str = serde_json::to_string(&offer_request).unwrap();
        assert!(json_str.contains("offer"));
        assert!(json_str.contains("test-session"));
    }

    #[tokio::test]
    async fn test_ice_candidate_serialization() {
        let candidate = IceCandidate {
            candidate: "candidate:1 1 UDP 2013266431 192.168.1.1 54400 typ host".to_string(),
            sdp_m_line_index: Some(0),
            sdp_mid: Some("0".to_string()),
        };

        let json_str = serde_json::to_string(&candidate).unwrap();
        assert!(json_str.contains("candidate:1"));
        assert!(json_str.contains("sdpMLineIndex"));
    }
}