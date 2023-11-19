package pgb

import (
	"database/sql"
	"goproxy/common"
	"log"

	_ "github.com/lib/pq"
)

func Connect(url string) *sql.DB {
	connStr := url
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf(common.ERR_connection, err)
		return nil
	}
	return db
}

func Log(db *sql.DB, elem common.LoggingElem) {

	//db.Ping()
	_, err := db.Exec(common.ExecuteLogString,
		elem.LType, elem.SID, elem.Head, elem.Body, elem.Times)
	if err != nil {
		if (common.DEBUG_LEVEL & 1) == 1 {
			log.Printf(common.Dbg_ERR_execsql, elem.SID, err)
		}
	}
	if (common.DEBUG_LEVEL & 2) == 2 {
		log.Printf(common.Dbg_Info_sidsql, elem.LType, elem.SID)
	}

}
