/**
 * WebRTC SDP Client Library for RustPBX
 * 
 * This library provides a simple interface for WebRTC SDP offer/answer exchange
 * with the RustPBX server.
 * 
 * Usage:
 *   const client = new WebRTCSDPClient('http://localhost:8080');
 *   const { sessionId, answerSdp } = await client.sendOffer(offerSdp);
 */

class WebRTCSDPClient {
    /**
     * Create a new WebRTC SDP client
     * @param {string} baseUrl - Base URL of the RustPBX server
     * @param {object} options - Optional configuration
     */
    constructor(baseUrl, options = {}) {
        this.baseUrl = baseUrl.replace(/\/$/, ''); // Remove trailing slash
        this.timeout = options.timeout || 30000; // 30 second default timeout
        this.headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };
    }

    /**
     * Send SDP offer and receive SDP answer
     * @param {string} offerSdp - SDP offer string
     * @param {string} sessionId - Optional session ID
     * @param {object} metadata - Optional metadata
     * @returns {Promise<{sessionId: string, answerSdp: string, iceCanvidates?: Array}>}
     */
    async sendOffer(offerSdp, sessionId = null, metadata = null) {
        const request = {
            sdp: {
                type: 'offer',
                sdp: offerSdp
            },
            ...(sessionId && { session_id: sessionId }),
            ...(metadata && { metadata })
        };

        const response = await this._makeRequest('POST', '/webrtc/offer', request);
        
        return {
            sessionId: response.session_id,
            answerSdp: response.sdp.sdp,
            iceCandidates: response.ice_candidates || []
        };
    }

    /**
     * Send ICE candidate
     * @param {string} sessionId - Session ID
     * @param {object} candidate - ICE candidate object
     * @returns {Promise<{sessionId: string, status: string}>}
     */
    async sendIceCandidate(sessionId, candidate) {
        const request = {
            session_id: sessionId,
            candidate: {
                candidate: candidate.candidate,
                sdpMLineIndex: candidate.sdpMLineIndex,
                sdpMid: candidate.sdpMid
            }
        };

        return await this._makeRequest('POST', '/webrtc/ice-candidate', request);
    }

    /**
     * Close WebRTC session
     * @param {string} sessionId - Session ID
     * @param {string} reason - Optional reason for closing
     * @returns {Promise<{sessionId: string, status: string}>}
     */
    async closeSession(sessionId, reason = null) {
        const request = {
            session_id: sessionId,
            ...(reason && { reason })
        };

        return await this._makeRequest('POST', '/webrtc/close', request);
    }

    /**
     * Get ICE servers
     * @returns {Promise<Array<{urls: Array<string>, username?: string, credential?: string}>>}
     */
    async getIceServers() {
        return await this._makeRequest('GET', '/iceservers');
    }

    /**
     * Simple SDP exchange (convenience method)
     * @param {string} offerSdp - SDP offer string
     * @returns {Promise<{sessionId: string, answerSdp: string}>}
     */
    async exchangeSdp(offerSdp) {
        const result = await this.sendOffer(offerSdp);
        return {
            sessionId: result.sessionId,
            answerSdp: result.answerSdp
        };
    }

    /**
     * Complete session setup with ICE candidates
     * @param {string} offerSdp - SDP offer string
     * @param {Array} iceCandidates - Array of ICE candidates
     * @returns {Promise<{sessionId: string, answerSdp: string}>}
     */
    async setupSession(offerSdp, iceCandidates = []) {
        // Send offer
        const result = await this.sendOffer(offerSdp);
        
        // Send ICE candidates
        for (const candidate of iceCandidates) {
            try {
                await this.sendIceCandidate(result.sessionId, candidate);
            } catch (error) {
                console.warn('Failed to send ICE candidate:', error);
            }
        }
        
        return {
            sessionId: result.sessionId,
            answerSdp: result.answerSdp
        };
    }

    /**
     * Make HTTP request to the server
     * @private
     * @param {string} method - HTTP method
     * @param {string} endpoint - API endpoint
     * @param {object} data - Request data (for POST requests)
     * @returns {Promise<object>}
     */
    async _makeRequest(method, endpoint, data = null) {
        const url = `${this.baseUrl}${endpoint}`;
        
        const options = {
            method,
            headers: this.headers,
            signal: AbortSignal.timeout(this.timeout)
        };

        if (data && (method === 'POST' || method === 'PUT')) {
            options.body = JSON.stringify(data);
        }

        try {
            const response = await fetch(url, options);
            const responseData = await response.json();

            if (!response.ok) {
                const error = new Error(responseData.error || `HTTP ${response.status}`);
                error.code = responseData.code || response.status;
                error.sessionId = responseData.session_id;
                throw error;
            }

            return responseData;
        } catch (error) {
            if (error.name === 'AbortError') {
                throw new Error('Request timeout');
            }
            throw error;
        }
    }
}

// Example usage functions
class WebRTCSDPExamples {
    constructor(client) {
        this.client = client;
    }

    /**
     * Basic WebRTC peer connection with SDP exchange
     */
    async basicPeerConnection() {
        console.log('=== Basic Peer Connection Example ===');
        
        // Create RTCPeerConnection
        const peerConnection = new RTCPeerConnection({
            iceServers: await this.client.getIceServers()
        });

        // Add audio track (you can replace this with your own media)
        const stream = await navigator.mediaDevices.getUserMedia({ 
            audio: true, 
            video: false 
        });
        
        for (const track of stream.getTracks()) {
            peerConnection.addTrack(track, stream);
        }

        // Create offer
        const offer = await peerConnection.createOffer();
        await peerConnection.setLocalDescription(offer);

        console.log('Created SDP offer');

        // Send offer to server and get answer
        const { sessionId, answerSdp } = await this.client.exchangeSdp(offer.sdp);
        
        console.log('Received SDP answer, session ID:', sessionId);

        // Set remote description
        await peerConnection.setRemoteDescription({
            type: 'answer',
            sdp: answerSdp
        });

        console.log('WebRTC connection established');

        // Return connection info for cleanup
        return { peerConnection, sessionId, stream };
    }

    /**
     * Advanced WebRTC setup with ICE candidate handling
     */
    async advancedPeerConnection() {
        console.log('=== Advanced Peer Connection Example ===');

        const iceServers = await this.client.getIceServers();
        const peerConnection = new RTCPeerConnection({ iceServers });

        // Store ICE candidates to send later
        const iceCandidates = [];
        
        peerConnection.onicecandidate = (event) => {
            if (event.candidate) {
                iceCandidates.push({
                    candidate: event.candidate.candidate,
                    sdpMLineIndex: event.candidate.sdpMLineIndex,
                    sdpMid: event.candidate.sdpMid
                });
                console.log('ICE candidate generated:', event.candidate.candidate);
            }
        };

        // Add media stream
        const stream = await navigator.mediaDevices.getUserMedia({ 
            audio: true, 
            video: false 
        });
        
        for (const track of stream.getTracks()) {
            peerConnection.addTrack(track, stream);
        }

        // Create and set local description
        const offer = await peerConnection.createOffer();
        await peerConnection.setLocalDescription(offer);

        // Wait for ICE gathering to complete
        await new Promise((resolve) => {
            if (peerConnection.iceGatheringState === 'complete') {
                resolve();
            } else {
                peerConnection.addEventListener('icegatheringstatechange', () => {
                    if (peerConnection.iceGatheringState === 'complete') {
                        resolve();
                    }
                });
            }
        });

        console.log('ICE gathering complete, collected', iceCandidates.length, 'candidates');

        // Setup complete session
        const { sessionId, answerSdp } = await this.client.setupSession(offer.sdp, iceCandidates);

        console.log('Session setup complete, session ID:', sessionId);

        // Set remote description
        await peerConnection.setRemoteDescription({
            type: 'answer',
            sdp: answerSdp
        });

        console.log('Advanced WebRTC connection established');

        return { peerConnection, sessionId, stream };
    }

    /**
     * Example with metadata and error handling
     */
    async metadataExample() {
        console.log('=== Metadata Example ===');

        try {
            const peerConnection = new RTCPeerConnection({
                iceServers: await this.client.getIceServers()
            });

            // Add data channel
            const dataChannel = peerConnection.createDataChannel('chat', {
                ordered: true
            });

            // Create offer
            const offer = await peerConnection.createOffer();
            await peerConnection.setLocalDescription(offer);

            // Send offer with metadata
            const metadata = {
                userId: 'user123',
                callType: 'data',
                timestamp: new Date().toISOString(),
                features: ['chat', 'file-transfer']
            };

            const result = await this.client.sendOffer(
                offer.sdp,
                'metadata-session-' + Date.now(),
                metadata
            );

            console.log('Session created with metadata:', result.sessionId);

            // Set remote description
            await peerConnection.setRemoteDescription({
                type: 'answer',
                sdp: result.answerSdp
            });

            return { peerConnection, sessionId: result.sessionId, dataChannel };

        } catch (error) {
            console.error('Metadata example failed:', error.message);
            if (error.code) {
                console.error('Error code:', error.code);
            }
            throw error;
        }
    }

    /**
     * Concurrent sessions example
     */
    async concurrentSessions() {
        console.log('=== Concurrent Sessions Example ===');

        const sessions = [];
        const promises = [];

        for (let i = 0; i < 3; i++) {
            const promise = this.createSession(`concurrent-${i}`);
            promises.push(promise);
        }

        const results = await Promise.allSettled(promises);
        
        for (let i = 0; i < results.length; i++) {
            if (results[i].status === 'fulfilled') {
                sessions.push(results[i].value);
                console.log(`Session ${i} created successfully:`, results[i].value.sessionId);
            } else {
                console.error(`Session ${i} failed:`, results[i].reason.message);
            }
        }

        return sessions;
    }

    /**
     * Helper function to create a session
     * @private
     */
    async createSession(prefix) {
        const peerConnection = new RTCPeerConnection({
            iceServers: await this.client.getIceServers()
        });

        // Add a dummy audio track
        const audioContext = new AudioContext();
        const oscillator = audioContext.createOscillator();
        const destination = audioContext.createMediaStreamDestination();
        oscillator.connect(destination);
        oscillator.start();

        const track = destination.stream.getAudioTracks()[0];
        peerConnection.addTrack(track, destination.stream);

        const offer = await peerConnection.createOffer();
        await peerConnection.setLocalDescription(offer);

        const { sessionId, answerSdp } = await this.client.exchangeSdp(offer.sdp);

        await peerConnection.setRemoteDescription({
            type: 'answer',
            sdp: answerSdp
        });

        return { peerConnection, sessionId, audioContext };
    }

    /**
     * Cleanup resources
     */
    async cleanup(connections) {
        for (const conn of connections) {
            if (conn.peerConnection) {
                conn.peerConnection.close();
            }
            if (conn.stream) {
                conn.stream.getTracks().forEach(track => track.stop());
            }
            if (conn.audioContext) {
                await conn.audioContext.close();
            }
            if (conn.sessionId) {
                try {
                    await this.client.closeSession(conn.sessionId, 'Example cleanup');
                } catch (error) {
                    console.warn('Failed to close session:', error.message);
                }
            }
        }
    }
}

// Export for use in Node.js or browsers
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { WebRTCSDPClient, WebRTCSDPExamples };
} else if (typeof window !== 'undefined') {
    window.WebRTCSDPClient = WebRTCSDPClient;
    window.WebRTCSDPExamples = WebRTCSDPExamples;
}

// Example usage (for browser console or Node.js)
async function runExamples() {
    console.log('Starting WebRTC SDP Client Examples');
    
    const client = new WebRTCSDPClient('http://localhost:8080');
    const examples = new WebRTCSDPExamples(client);
    
    const connections = [];
    
    try {
        // Run basic example
        const basic = await examples.basicPeerConnection();
        connections.push(basic);
        
        // Wait a bit
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        // Run advanced example
        const advanced = await examples.advancedPeerConnection();
        connections.push(advanced);
        
        // Wait a bit
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        // Run metadata example
        const metadata = await examples.metadataExample();
        connections.push(metadata);
        
        // Wait a bit
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        // Run concurrent sessions
        const concurrent = await examples.concurrentSessions();
        connections.push(...concurrent);
        
        console.log('All examples completed successfully!');
        
        // Wait before cleanup
        await new Promise(resolve => setTimeout(resolve, 5000));
        
    } catch (error) {
        console.error('Example failed:', error);
    } finally {
        // Cleanup
        await examples.cleanup(connections);
        console.log('Cleanup completed');
    }
}

// Uncomment to run examples automatically
// runExamples().catch(console.error);