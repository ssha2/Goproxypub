package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/google/uuid"
)

const (
	deflocal  = "localhost:4589"
	defurl    = "https://httpbin.org"
	pepHeader = "X-PepFollower-ID"
	//Для теста
	//curl -X POST "http://localhost:4589/post" -H "accept: application/json" -d '{"test:test"}'
	// .\goproxy.exe -l "localhost:4589" -t "https://httpbin.org"
	deftopicname = "loggingproxy"
	defbrokerard = "localhost:9092"
	defsize      = 5
	defcount     = 5
)

const (
	loggingException = "except"
	loggingRequest   = "reqs"
	loggingResponse  = "resp"
)

/***********************writer для proxy + логирование response*******************************/

type CustomResponseWriter struct {
	http.ResponseWriter
	u_id string
}

func (rr *CustomResponseWriter) WriteHeader(statusCode int) {
	rr.Header().Add(pepHeader, rr.u_id)
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *CustomResponseWriter) Write(b []byte) (int, error) {

	headBytes := []byte("")
	for name, values := range rr.Header() {
		for _, value := range values {
			headBytes = append(headBytes, "\n"...)
			headBytes = append(headBytes, name...)
			headBytes = append(headBytes, ":"...)
			headBytes = append(headBytes, value...)
		}
	}

	loggingsend(loggingElem{loggingResponse, rr.u_id, headBytes, b, time.Now()})

	return rr.ResponseWriter.Write(b)
}

/***********************Логирование request *******************************/

func generateUniqueID() string {
	id := uuid.New()
	return id.String()
}

func logRequest(r *http.Request, u_id *string) {

	headBytes := []byte("url:")
	headBytes = append(headBytes, r.URL.Path...)
	headBytes = append(headBytes, "\nquery:"...)
	headBytes = append(headBytes, r.URL.RawQuery...)
	for name, values := range r.Header {
		for _, value := range values {
			headBytes = append(headBytes, "\n"...)
			headBytes = append(headBytes, name...)
			headBytes = append(headBytes, ":"...)
			headBytes = append(headBytes, value...)
		}
	}

	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Body)
		defer r.Body.Close()
		// перевыставляем буфер
		r.Body = io.NopCloser(io.NewSectionReader(bytes.NewReader(bodyBytes), 0, int64(len(bodyBytes))))
	}
	loggingsend(loggingElem{loggingRequest, *u_id, headBytes, bodyBytes, time.Now()})

}

/***********************Хендлеры http*******************************/

// хендлеры для  реквеста
func servHandler(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u_id := generateUniqueID()
		r.Header.Add(pepHeader, u_id)
		logRequest(r, &u_id)
		if nextHandler != nil {
			nextHandler.ServeHTTP(&CustomResponseWriter{w, u_id}, r)
		}
	})
}

// handle при установки обратного проксти
func directionHandler(targetUrl string) func(*http.Request) {
	return func(req *http.Request) {
		target, _ := url.Parse(targetUrl)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}
}

// handle ошибки от прокси
func errHandler(w http.ResponseWriter, r *http.Request, e error) {

	headBytes := []byte("")
	for name, values := range r.Header {
		for _, value := range values {
			headBytes = append(headBytes, "\n"...)
			headBytes = append(headBytes, name...)
			headBytes = append(headBytes, ":"...)
			headBytes = append(headBytes, value...)
		}
	}
	loggingsend(loggingElem{loggingException, r.Header.Get(pepHeader), headBytes, []byte(e.Error()), time.Now()})
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(e.Error()))

}

/******************************************************/
func main() {

	var (
		local     string
		targetURL string
		topicname string
		brokeradr string
		size      int
		count     int
	)

	flag.StringVar(&local, "l", deflocal, "Local ip:port")
	flag.StringVar(&targetURL, "t", defurl, "Target http(s)://ip:port ")
	flag.StringVar(&topicname, "n", deftopicname, "Kafaka topicname")
	flag.StringVar(&brokeradr, "a", defbrokerard, "Kafaka brokeradr ")
	flag.IntVar(&size, "s", defsize, "Channel size ")
	flag.IntVar(&count, "c", defcount, "Channels count ")
	flag.Parse()

	//logging
	logginginit(count, size)
	loggingrun()

	// config kafaka producer
	configkafka(topicname, brokeradr)

	//reverce proxy
	proxy := &httputil.ReverseProxy{
		Director:     directionHandler(targetURL),
		ErrorHandler: errHandler,
	}

	//// to be remove just for test
	//go toberevome_consumer()
	//////////////////////////////

	if err := http.ListenAndServe(local, servHandler(proxy)); err != nil {
		log.Fatal(err)
	}

}
