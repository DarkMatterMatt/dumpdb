package getsourceid

import (
	"database/sql"
	"log"
	"strings"

	lru "github.com/hashicorp/golang-lru"
)

var sourceCache *lru.Cache

// GetSourceID fetches an integer ID for a string `source` from the sources table.
func GetSourceID(source string, sourceDb *sql.DB, sourceTable string) uint32 {
	if strings.HasPrefix(source, "./") {
		source = source[2:]
	}

	// truncate source name to 250 characters (database limitation)
	if len(source) > 250 {
		source = source[:250]
	}

	// load from cache
	val, ok := sourceCache.Get(source)
	if ok {
		// check that it's an integer :(
		if val, ok := val.(uint32); ok {
			return val
		}
	}

	// upsert from database
	res, err := sourceDb.Exec(`
		INSERT INTO `+sourceTable+` (name, last_updated) VALUES(?, UNIX_TIMESTAMP())
		ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)
	`, source)
	if err != nil {
		log.Fatal(err)
	}

	// get id from database
	id, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	// save in cache
	sourceCache.Add(source, uint32(id))
	return uint32(id)
}
