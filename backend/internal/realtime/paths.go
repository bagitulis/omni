package realtime

// RealtimePath is the single mount point for the dashboard WebSocket.
// Kept as a const so nginx config, main.go, and the frontend client can share
// it via search.
const RealtimePath = "/api/realtime/ws"
