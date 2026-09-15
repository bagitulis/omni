package realtime

import "errors"

// Authenticate is the pluggable JWT → (tenantID, userID, role) resolver.
// Kept as a function type so tests can substitute a fake without importing
// the JWT package. Production wires this to utils.JWTService.ValidateToken.
type Authenticate func(token string) (tenantID, userID, role string, err error)

// ErrInvalidToken is returned by Authenticate implementations when the JWT
// fails validation (signature, expiry, claims). Kept generic so failure
// messages sent back to the client do not disclose which check failed.
var ErrInvalidToken = errors.New("invalid token")
