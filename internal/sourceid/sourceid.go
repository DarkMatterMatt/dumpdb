package sourceid

import (
	"database/sql"
	"log"
	"strings"

	lru "github.com/hashicorp/golang-lru"
)

var (
	sourceNameCache, _ = lru.New(1000)
	sourceIDCache, _   = lru.New(1000)
)

// SourceName resolves a source ID to it's source name
func SourceName(id int64, sourceDb *sql.DB, sourceTable string) (string, error) {
	// load from cache
	val, ok := sourceNameCache.Get(id)
	if ok {
		// check that it's an string :(
		if val, ok := val.(string); ok {
			return val, nil
		}
	}

	// fetch from database
	var s string
	err := sourceDb.QueryRow(`
		SELECT name
		FROM `+sourceTable+`
		WHERE id=?
	`, id).Scan(&s)
	if err != nil {
		if err == sql.ErrNoRows {
			return "CouldNotFindSource", nil
		}
		return "", err
	}

	// save in cache
	sourceNameCache.Add(s, id)
	return s, nil
}

// SourceID fetches an integer ID for a string `s` from the sources table
func SourceID(s string, sourceDb *sql.DB, sourceTable string) (int64, error) {
	if strings.HasPrefix(s, "./") {
		s = s[2:]
	}

	// truncate source name to 250 characters (database limitation)
	if len(s) > 250 {
		s = s[:250]
	}

	// load from cache
	val, ok := sourceIDCache.Get(s)
	if ok {
		// check that it's an integer :(
		if val, ok := val.(int64); ok {
			return val, nil
		}
	}

	// upsert from database
	res, err := sourceDb.Exec(`
		INSERT INTO `+sourceTable+` (name, date_added) VALUES(?, UNIX_TIMESTAMP())
		ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)
	`, s)
	if err != nil {
		return 0, nil
	}

	// get id from database
	id, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	// save in cache
	sourceIDCache.Add(s, id)
	return id, nil
}
