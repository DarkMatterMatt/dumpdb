package parseline

// A Record represents a single row in the database
type Record struct {
	Source   string
	SourceID uint32
	Username string
	Email    string
	EmailRev string
	Hash     string
	Password string
}
