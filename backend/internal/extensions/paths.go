package extensions

// WSPath is the WebSocket endpoint for extensions.
//
// Declared once so the handler that reports it to a newly paired extension and
// the route registration cannot drift apart — a mismatch would hand the browser
// a URL that 404s right after a successful pairing.
const WSPath = "/api/extensions/ws"
