package common

import "time"

type LoggingElem struct {
	LType string    // respomse or request or error
	SID   string    // session id
	Head  []byte    // headers
	Body  []byte    // body
	Times time.Time //timestamp
}

const (
	LogException = "Exception"
	LogRequest   = "Request"
	LogResponse  = "Response"
)

const PepHeader = "X-PepFollower-ID"

const (
	Deflocal = "localhost:4589"
	Defurl   = "https://httpbin.org"
	Defsize  = 5
	Defcount = 5
	Defpgurl = "куда-то"
)

const ExecuteLogString = "insert into public.logging(ltype,sid,head,body,t) values($1,$2,$3,$4,$5)"

const (
	ERR_initconn    = "ERROR:Connection  for channel #%d - channel closed"
	OK_initconn     = "OK:Connection set for channel #%d"
	OK_cycleend     = "OK:Channel #%d cycle end"
	Dbg_selchannel  = "Debug:selected channel %d\n"
	Dbg_nochannel   = "Info:not any selected channel"
	Dbg_ERR_execsql = "ERROR:execsql  ssid#%s error>> %s"
	ERR_connection  = "ERROR: connect to db error>> %s"
	Dbg_Info_sidsql = "Info: save to db type %s sid#%s"
)

var DEBUG_LEVEL = 3 /* 0 - no, 1- info, 2- log */

var (
	RGS_ExcepBodyPath = map[string]bool{}
	RSP_ExcepBodyPath = map[string]bool{}
)
