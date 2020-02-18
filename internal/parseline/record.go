package parseline

// A Record represents a single row in the database
type Record struct {
	Source   string
	SourceID int64
	Username string
	Email    string
	EmailRev string
	Hash     string
	Password string
	Extra    string
}
