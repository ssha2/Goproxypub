package logger

/*логирование в нескольких каналах*/

import (
	"database/sql"
	"goproxy/common"
	"goproxy/pgb"
	"log"
)

var (
	counts   int                       = 1
	channels []chan common.LoggingElem = nil
)

/***********************сайзинг каналов *******************************/
func Logginginit(countsize int, size int) {
	counts = countsize
	// на всякий
	if counts <= 0 {
		counts = 1
	}
	if size <= 0 {
		size = 1
	}
	channels = []chan common.LoggingElem{make(chan common.LoggingElem, size)}
	for i := 2; i <= counts; i++ {
		channels = append(channels, make(chan common.LoggingElem, size))
	}
}

/***********************выбор каналов *******************************/

func Loggingsend(element common.LoggingElem) {

	go func(elem common.LoggingElem) {
		selected := 0
		for i := 0; i < counts; i++ {
			b := loggingselector(elem, channels[i])
			if b {
				selected = i + 1
				break
			}
		}
		if selected > 0 {
			if (common.DEBUG_LEVEL & 2) == 2 {
				log.Printf(common.Dbg_selchannel, selected)
			}

		} else {
			if (common.DEBUG_LEVEL & 1) == 1 {
				log.Println(common.Dbg_nochannel)
			}
		}
	}(element)
}

func loggingselector(elem common.LoggingElem, cnl chan common.LoggingElem) bool {
	// non blocking send
	select {
	case cnl <- elem:
		return true
	default:
		return false
	}
}

/***********************из канала в DB*******************************/

func Loggingrun(pgurl string) {

	for i := 0; i < counts; i++ {

		go func(n int, chn chan common.LoggingElem, purl string) {
			var db *sql.DB = pgb.Connect(purl)
			if db == nil {
				close(chn)
				log.Printf(common.ERR_initconn, n)
				return
			} else {
				if (common.DEBUG_LEVEL & 1) == 1 {
					log.Printf(common.OK_initconn, n)
				}
			}
			defer db.Close()
			for {
				elem, runned := <-chn
				pgb.Log(db, elem)
				if !runned {
					if (common.DEBUG_LEVEL & 1) == 1 {
						log.Printf(common.OK_cycleend, n)
					}
					break
				}
			}
		}(i+1, channels[i], pgurl)

	}

}
