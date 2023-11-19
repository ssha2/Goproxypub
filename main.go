package main

import (
	"bytes"
	"flag"
	"goproxy/common"
	"goproxy/logger"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/google/uuid"
)

//Для теста
//curl -X POST "http://localhost:4589/post" -H "accept: application/json" -d '{"test:test"}'
// .\goproxy.exe -l "localhost:4589" -t "https://httpbin.org"

/***********************writer для proxy + логирование response*******************************/

type CustomResponseWriter struct {
	http.ResponseWriter
	u_id string
}

func (rr *CustomResponseWriter) WriteHeader(statusCode int) {
	rr.Header().Add(common.PepHeader, rr.u_id)
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *CustomResponseWriter) Write(b []byte) (int, error) {

	// режет большие объемы flushем - логирование перенесено в modify
	// headBytes := []byte("")
	// for name, values := range rr.Header() {
	// 	for _, value := range values {
	// 		headBytes = append(headBytes, name...)
	// 		headBytes = append(headBytes, ":"...)
	// 		headBytes = append(headBytes, value...)
	// 		headBytes = append(headBytes, "\n"...)
	// 	}
	// }
	// logger.Loggingsend(common.LoggingElem{LType: common.LogResponse, SID: rr.u_id, Head: headBytes, Body: b, Times: time.Now()})

	return rr.ResponseWriter.Write(b)
}

func logResponse(rs *http.Response) {

	headBytes := []byte("")
	for name, values := range rs.Header {
		for _, value := range values {
			headBytes = append(append(append(append(headBytes, name...), ":"...), value...), "\n"...)
		}
	}
	var bodyBytes []byte
	if rs.Body != nil {
		bodyBytes, _ = io.ReadAll(rs.Body)
		defer rs.Body.Close()
		// перевыставляем буфер
		rs.Body = io.NopCloser(io.NewSectionReader(bytes.NewReader(bodyBytes), 0, int64(len(bodyBytes))))
	}
	logger.Loggingsend(common.LoggingElem{LType: common.LogResponse, SID: rs.Header.Get(common.PepHeader), Head: headBytes, Body: bodyBytes, Times: time.Now()})

}

/***********************Логирование request *******************************/

func generateUniqueID() string {
	id := uuid.New()
	return id.String()
}

func logRequest(r *http.Request, u_id *string) {

	headBytes := append(append(append([]byte("path:"), r.URL.Path...), "\nquery:"...), r.URL.RawQuery...)
	for name, values := range r.Header {
		for _, value := range values {
			headBytes = append(append(append(append(headBytes, "\n"...), name...), ":"...), value...)
		}
	}
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Body)
		defer r.Body.Close()
		r.Body = io.NopCloser(io.NewSectionReader(bytes.NewReader(bodyBytes), 0, int64(len(bodyBytes))))
	}
	logger.Loggingsend(common.LoggingElem{LType: common.LogRequest, SID: *u_id, Head: headBytes, Body: bodyBytes, Times: time.Now()})

}

/***********************Хендлеры http*******************************/

// handle для  реквеста
func servHandler(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u_id := generateUniqueID()
		r.Header.Add(common.PepHeader, u_id)
		logRequest(r, &u_id)
		if nextHandler != nil {
			nextHandler.ServeHTTP(&CustomResponseWriter{w, u_id}, r)
		}
	})
}

// handle direction
func directionHandler(targetUrl string) func(*http.Request) {
	return func(req *http.Request) {
		target, _ := url.Parse(targetUrl)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}
}

// handle modify
func modifyHandler() func(*http.Response) error {
	return func(rs *http.Response) error {
		logResponse(rs)
		return nil
	}
}

// handle ошибки от прокси
func errHandler(w http.ResponseWriter, r *http.Request, e error) {
	headBytes := []byte("")
	for name, values := range r.Header {
		for _, value := range values {
			headBytes = append(append(append(append(headBytes, "\n"...), name...), ":"...), value...)
		}
	}
	logger.Loggingsend(common.LoggingElem{LType: common.LogException, SID: r.Header.Get(common.PepHeader), Head: headBytes, Body: []byte(e.Error()), Times: time.Now()})
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(e.Error()))

}

/******************************************************/
func main() {
	var (
		local     string
		targetURL string
		size      int
		count     int
		pgurl     string
	)

	flag.StringVar(&local, "l", common.Deflocal, "Local ip:port")
	flag.StringVar(&targetURL, "t", common.Defurl, "Target http(s)://ip:port ")
	flag.IntVar(&size, "s", common.Defsize, "Channel size ")
	flag.IntVar(&count, "c", common.Defcount, "Channels count ")
	flag.StringVar(&pgurl, "g", common.Defpgurl, "Posg url")
	flag.Parse()

	//logging
	logger.Logginginit(count, size)
	logger.Loggingrun(&pgurl)

	//reverce proxy
	proxy := &httputil.ReverseProxy{
		Director:       directionHandler(targetURL),
		ErrorHandler:   errHandler,
		ModifyResponse: modifyHandler(),
	}

	if err := http.ListenAndServe(local, servHandler(proxy)); err != nil {
		log.Fatal(err)
	}

}
