package main

//Для теста
//curl -X POST "http://localhost:4589/post" -H "accept: application/json" -d '{"test:test"}'
//curl -X POST "http://localhost:4589/post" -H "accept: application/json" --data-binary "@C:/Users/ssha/Downloads/demo-big-20170815.sql"
// .\goproxy.exe -l "localhost:4589" -t "https://httpbin.org"

import (
	"bytes"
	"flag"
	"goproxy/common"
	"goproxy/logger"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"strconv"
	"strings"
	"time"
)

/***********************log response *******************************/
func logResponse(rs *http.Response) {

	headBytes := []byte("")
	for name, values := range rs.Header {
		for _, value := range values {
			headBytes = append(append(append(append(headBytes, name...), ":"...), value...), "\n"...)
		}
	}
	var bodyBytes []byte
	if !common.RSP_ExcepBodyPath[rs.Request.URL.Path] {
		if rs.Body != nil {
			bodyBytes, _ = io.ReadAll(rs.Body)
			defer rs.Body.Close()
			// перевыставляем буфер
			rs.Body = io.NopCloser(io.NewSectionReader(bytes.NewReader(bodyBytes), 0, int64(len(bodyBytes))))
		}
	}
	logger.Loggingsend(common.LoggingElem{
		LType: common.LogResponse,
		SID:   rs.Request.Header.Get(common.PepHeader),
		Head:  headBytes,
		Body:  bodyBytes,
		Times: time.Now()})

}

/***********************log request *******************************/

func logRequest(r *http.Request, u_id *string) {

	headBytes := append(append(append([]byte("path:"), r.URL.Path...), "\nquery:"...), r.URL.RawQuery...)
	for name, values := range r.Header {
		for _, value := range values {
			headBytes = append(append(append(append(headBytes, "\n"...), name...), ":"...), value...)
		}
	}
	var bodyBytes []byte
	if !common.RGS_ExcepBodyPath[r.URL.Path] {
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			defer r.Body.Close()
			r.Body = io.NopCloser(io.NewSectionReader(bytes.NewReader(bodyBytes), 0, int64(len(bodyBytes))))
		}
	}
	logger.Loggingsend(common.LoggingElem{
		LType: common.LogRequest,
		SID:   *u_id,
		Head:  headBytes,
		Body:  bodyBytes,
		Times: time.Now()})

}

/******************************************************/
func main() {
	var (
		local     string
		targetURL string
		size      int
		count     int
		pgurl     string
		rgssplit  string
		rspsplit  string
	)

	flag.StringVar(&local, "l", common.Deflocal, "local ip:port")
	flag.StringVar(&targetURL, "t", common.Defurl, "target http(s)://ip:port ")
	flag.IntVar(&size, "s", common.Defsize, "one channel size ")
	flag.IntVar(&count, "c", common.Defcount, "channels count ")
	flag.StringVar(&pgurl, "g", common.Defpgurl, "postgres url")
	flag.StringVar(&rgssplit, "i", "", "skip body requets  \"path1,path2,path..\"")
	flag.StringVar(&rspsplit, "o", "", "skip body response \"path1,path2,path..\"")
	flag.IntVar(&common.DEBUG_LEVEL, "d", common.DEBUG_LEVEL, "debug level (0,1,2,3)")
	flag.Parse()

	for _, v := range strings.Split(rgssplit, ",") {
		common.RGS_ExcepBodyPath[v] = true
	}
	for _, v := range strings.Split(rspsplit, ",") {
		common.RSP_ExcepBodyPath[v] = true
	}

	//logging set
	logger.Logginginit(count, size)
	logger.Loggingrun(pgurl)

	//reverce proxy set
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			target, _ := url.Parse(targetURL)
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, e error) {
			headBytes := []byte("")
			for name, values := range r.Header {
				for _, value := range values {
					headBytes = append(append(append(append(headBytes, "\n"...), name...), ":"...), value...)
				}
			}
			logger.Loggingsend(common.LoggingElem{LType: common.LogException,
				SID:   r.Header.Get(common.PepHeader),
				Head:  headBytes,
				Body:  []byte(e.Error()),
				Times: time.Now()})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(e.Error()))

		},
		ModifyResponse: func(rs *http.Response) error {
			rs.Header.Add(common.PepHeader, rs.Request.Header.Get(common.PepHeader))
			logResponse(rs)
			return nil
		},
	}

	if err := http.ListenAndServe(local,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u_id := strconv.FormatInt(time.Now().UnixNano(), 10)
			r.Header.Add(common.PepHeader, u_id)
			logRequest(r, &u_id)
			proxy.ServeHTTP(w, r)
		})); err != nil {
		log.Fatal(err)
	}

}
