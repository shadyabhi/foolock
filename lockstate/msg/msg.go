package msg

// Message constants for lock operations
const (
	Acquired          = "acquired"
	Renewed           = "renewed"
	HeldByAnother     = "held by another client"
	GracePeriodActive = "grace period active"
	LockReleased      = "lock released"
	ClientNotHolder   = "client does not hold the lock"
	LockHeld          = "lock held"
	NoLockHeld        = "no lock held"
)
