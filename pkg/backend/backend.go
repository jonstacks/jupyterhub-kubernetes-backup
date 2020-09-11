package backend

// Backend is the interface for storing a backup to various locations
type Backend interface {
	// Save saves the content at the given path to the backend
	Save(path string) error
}
