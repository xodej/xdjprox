// Copyright 2019 Johannes Mueller <circus2@web.de>

package main

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
)

// Server handles requests
type Server struct {
	*http.Server
	router *http.ServeMux
	config Config
}

func (s *Server) blockRequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("blocked request %s for session %s\n", r.URL.Path, r.URL.Query().Get("sid"))

	w.WriteHeader(http.StatusBadRequest)
	_, err := fmt.Fprintf(w, "%s", `1009;"not authorized for operation";"`+html.EscapeString(r.URL.Path)+` (blocked by xdjproxy)";`)
	if err != nil {
		log.Printf("failed to send block message to client\n")
		return
	}
}

func (s *Server) logResponse(r *http.Response) (err error) {
	b, erro := ioutil.ReadAll(r.Body)
	if erro != nil {
		log.Printf("failed to log OLAP response\n")
		return
	}

	erro = r.Body.Close()
	if erro != nil {
		log.Printf("failed to close http response\n")
		return
	}

	// log response body
	log.Printf("%s\n", string(b))

	body := ioutil.NopCloser(bytes.NewReader(b))
	r.Body = body
	r.ContentLength = int64(len(b))
	r.Header.Set("Content-Length", strconv.Itoa(len(b)))

	return nil
}

func (s *Server) logRequest(r *http.Request) {
	if r.Method == "GET" {
		log.Printf("%s\n", r.URL.String())
		return
	}

	buf, erro := ioutil.ReadAll(r.Body)
	if erro != nil {
		log.Printf("failed to log request\n")
		return
	}

	readerShow := ioutil.NopCloser(bytes.NewBuffer(buf))
	readerKeep := ioutil.NopCloser(bytes.NewBuffer(buf))
	log.Printf("%q\n", readerShow)
	r.Body = readerKeep
}

func (s *Server) forwardRequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("forwarded request %s for session %s\n", r.URL.Path, r.URL.Query().Get("sid"))

	// add headers to response
	w.Header().Add("Connection", "Keep-Alive")
	w.Header().Add("Keep-Alive", "timeout=5, max=100")

	// add headers to request
	r.Header.Set("Accept-Charset", "utf-8")
	r.Header.Set("Content-Type", "text/plain")

	// if logging is disabled add gzip compression header to request
	if !s.config.LogResponse {
		r.Header.Set("Accept-Encoding", "gzip,deflate")
	}

	if s.config.LogRequest {
		s.logRequest(r)
	}

	myURL, _ := url.Parse(s.config.TargetURL)

	proxy := httputil.NewSingleHostReverseProxy(myURL)
	proxy.FlushInterval = time.Millisecond * -1
	if s.config.LogResponse {
		proxy.ModifyResponse = s.logResponse
	}
	proxy.ServeHTTP(w, r)
}

// Register registers all special treated URL paths
func (s *Server) Register() {
	// whitelist of API calls
	s.router.HandleFunc("/server/change_password", s.forwardRequestHandler)
	s.router.HandleFunc("/server/databases", s.forwardRequestHandler)
	s.router.HandleFunc("/server/info", s.forwardRequestHandler)
	s.router.HandleFunc("/server/licenses", s.forwardRequestHandler)
	s.router.HandleFunc("/server/login", s.forwardRequestHandler)
	s.router.HandleFunc("/server/logout", s.forwardRequestHandler)
	s.router.HandleFunc("/server/user_info", s.forwardRequestHandler)

	s.router.HandleFunc("/database/cubes", s.forwardRequestHandler)
	s.router.HandleFunc("/database/dimensions", s.forwardRequestHandler)
	s.router.HandleFunc("/database/info", s.forwardRequestHandler)

	s.router.HandleFunc("/dimension/cubes", s.forwardRequestHandler)
	s.router.HandleFunc("/dimension/dfilter", s.forwardRequestHandler)
	s.router.HandleFunc("/dimension/element", s.forwardRequestHandler)
	s.router.HandleFunc("/dimension/elements", s.forwardRequestHandler)
	s.router.HandleFunc("/dimension/info", s.forwardRequestHandler)

	s.router.HandleFunc("/element/info", s.forwardRequestHandler)

	s.router.HandleFunc("/cube/holds", s.forwardRequestHandler)
	s.router.HandleFunc("/cube/info", s.forwardRequestHandler)
	s.router.HandleFunc("/cube/locks", s.forwardRequestHandler)
	s.router.HandleFunc("/cube/rules", s.forwardRequestHandler)

	s.router.HandleFunc("/cell/area", s.forwardRequestHandler)
	s.router.HandleFunc("/cell/drillthrough", s.forwardRequestHandler)
	s.router.HandleFunc("/cell/export", s.forwardRequestHandler)

	s.router.HandleFunc("/cell/value", s.forwardRequestHandler)
	s.router.HandleFunc("/cell/values", s.forwardRequestHandler)

	s.router.HandleFunc("/rule/functions", s.forwardRequestHandler)
	s.router.HandleFunc("/rule/info", s.forwardRequestHandler)
	s.router.HandleFunc("/rule/parse", s.forwardRequestHandler)

	s.router.HandleFunc("/svs/info", s.forwardRequestHandler)

	s.router.HandleFunc("/view/calculate", s.forwardRequestHandler)

	s.router.HandleFunc("/meta-sp", s.forwardRequestHandler)

	s.router.HandleFunc("/api", s.forwardRequestHandler)
	s.router.HandleFunc("/inc/", s.forwardRequestHandler)
	s.router.HandleFunc("/favicon.ico", s.forwardRequestHandler)

	// if writing is enabled pass everything to OLAP
	if s.config.EnableWrite {
		s.router.HandleFunc("/", s.forwardRequestHandler)
		return
	}

	// block all requests not in whitelist
	// potentially harmful to server integrity
	s.router.HandleFunc("/", s.blockRequestHandler)
}
