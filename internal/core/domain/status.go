package domain

// DiffStatus represents the schema comparison status between live pod and BSR
type DiffStatus string

const (
	// StatusSync indicates the schemas are in sync
	StatusSync DiffStatus = "SYNC"
	// StatusMismatch indicates schema drift has been detected
	StatusMismatch DiffStatus = "MISMATCH"
	// StatusUnknown indicates the status could not be determined
	StatusUnknown DiffStatus = "UNKNOWN"
)
