package notify

// NewRedisFanout is a build-time stub that returns InProcessFanout unless
// a Redis client is wired later. This keeps the package importable in
// environments without a Redis dependency, and lets ops enable Redis by
// swapping the constructor call site.
//
// The full implementation belongs in a follow-up: it subscribes to
// omni:notif:<tenantID> per active tenant, forwards to InProcessFanout for
// same-replica listeners, and publishes to Redis on Publish. Because it
// respects the same interface, no consumer changes.
func NewRedisFanout() Fanout {
	return NewInProcessFanout()
}
