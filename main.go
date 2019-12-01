package main

import (
	"flag"
	"fmt"
	"html"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
	uuid "github.com/gofrs/uuid"
)

// Dump http request
func dumpRequest(r *http.Request, contextLoggerWithReqID *log.Entry) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		contextLoggerWithReqID.Printf("failed to dump http request %v", err)
		return
	}
	contextLoggerWithReqID.WithFields(log.Fields{"type": "request"}).Printf("%s", dump)
}

// Dump http response
func dumpResponse(contextLoggerWithReqID *log.Entry) func(*http.Response) (err error) {
	return func(r *http.Response) (err error) {
		defer func() {
			err := r.Body.Close()
			if err != nil {
				contextLoggerWithReqID.Printf("failed to close http response %v", err)
				return
			}
		}()

		dump, myErr := httputil.DumpResponse(r, true)
		if myErr != nil {
			contextLoggerWithReqID.Printf("failed to dump http response %v", myErr)
			return nil
		}
		contextLoggerWithReqID.WithFields(log.Fields{"type": "response"}).Printf("%s", dump)

		return nil
	}
}

// Generate UUID per request
func generateRequestUUID() string {
	id, err := uuid.NewV4()
	if err != nil {
		log.Printf("failed to generate UUID for request %v", err)
		return ""
	}

	return id.String()
}

// Block http request
func blockRequest(conf *Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		contextLoggerWithReqID := log.WithFields(log.Fields{"request_id": generateRequestUUID(), "session": r.URL.Query().Get("sid")})
		contextLoggerWithReqID.WithFields(log.Fields{"type": "block"}).Printf("blocked request %s", r.URL.Path)

		w.WriteHeader(http.StatusBadRequest)
		_, err := fmt.Fprintf(w, "%s", `1009;"not authorized for operation";"`+html.EscapeString(r.URL.Path)+` (blocked by xdjprox)";`)
		if err != nil {
			contextLoggerWithReqID.Printf("failed to send block message to client %v", err)
			return
		}
	}
}

// Handle http request
func handleRequest(conf *Config, p *httputil.ReverseProxy) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		contextLoggerWithReqID := log.WithFields(log.Fields{"request_id": generateRequestUUID(), "session": r.URL.Query().Get("sid")})
		contextLoggerWithReqID.WithFields(log.Fields{"type": "forward"}).Printf("forwarded request %s", r.URL.Path)

		targetURL, _ := url.Parse(conf.TargetURL)

		// add headers to response
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Keep-Alive", "timeout=5, max=100")

		// add headers to request
		r.Header.Set("Accept-Charset", "utf-8")
		r.Header.Set("Content-Type", "text/plain")
		r.Header.Set("Connection", "keep-alive")
		r.Header.Set("Accept-Encoding", "gzip, deflate")
		r.URL.Host = targetURL.Host
		r.URL.Scheme = targetURL.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = targetURL.Host

		// log request if enabled
		if conf.LogRequest {
			dumpRequest(r, contextLoggerWithReqID)
		}

		// log response if enabled
		if conf.LogResponse {
			p.ModifyResponse = dumpResponse(contextLoggerWithReqID)
		}

		// ...and go...
		p.ServeHTTP(w, r)
	}
}

// Register routes for proxy
func registerRoutes(conf *Config, proxy *httputil.ReverseProxy) {
	// handle all requests read + write
	if conf.EnableWrite {
		http.HandleFunc("/", handleRequest(conf, proxy))
		return
	}

	// white-listed read-only routes handled by proxy
	http.HandleFunc("/server/change_password", handleRequest(conf, proxy))
	http.HandleFunc("/server/databases", handleRequest(conf, proxy))
	http.HandleFunc("/server/info", handleRequest(conf, proxy))
	http.HandleFunc("/server/licenses", handleRequest(conf, proxy))
	http.HandleFunc("/server/login", handleRequest(conf, proxy))
	http.HandleFunc("/server/logout", handleRequest(conf, proxy))
	http.HandleFunc("/server/user_info", handleRequest(conf, proxy))

	http.HandleFunc("/database/cubes", handleRequest(conf, proxy))
	http.HandleFunc("/database/dimensions", handleRequest(conf, proxy))
	http.HandleFunc("/database/info", handleRequest(conf, proxy))

	http.HandleFunc("/dimension/cubes", handleRequest(conf, proxy))
	http.HandleFunc("/dimension/dfilter", handleRequest(conf, proxy))
	http.HandleFunc("/dimension/element", handleRequest(conf, proxy))
	http.HandleFunc("/dimension/elements", handleRequest(conf, proxy))
	http.HandleFunc("/dimension/info", handleRequest(conf, proxy))

	http.HandleFunc("/element/info", handleRequest(conf, proxy))

	http.HandleFunc("/cube/holds", handleRequest(conf, proxy))
	http.HandleFunc("/cube/info", handleRequest(conf, proxy))
	http.HandleFunc("/cube/locks", handleRequest(conf, proxy))
	http.HandleFunc("/cube/rules", handleRequest(conf, proxy))

	http.HandleFunc("/cell/area", handleRequest(conf, proxy))
	http.HandleFunc("/cell/drillthrough", handleRequest(conf, proxy))
	http.HandleFunc("/cell/export", handleRequest(conf, proxy))

	http.HandleFunc("/cell/value", handleRequest(conf, proxy))
	http.HandleFunc("/cell/values", handleRequest(conf, proxy))

	http.HandleFunc("/rule/functions", handleRequest(conf, proxy))
	http.HandleFunc("/rule/info", handleRequest(conf, proxy))
	http.HandleFunc("/rule/parse", handleRequest(conf, proxy))

	http.HandleFunc("/svs/info", handleRequest(conf, proxy))

	http.HandleFunc("/view/calculate", handleRequest(conf, proxy))

	http.HandleFunc("/meta-sp", handleRequest(conf, proxy))

	http.HandleFunc("/api", handleRequest(conf, proxy))
	http.HandleFunc("/inc/", handleRequest(conf, proxy))
	http.HandleFunc("/favicon.ico", handleRequest(conf, proxy))

	// default block of all non-white-listed routes
	http.HandleFunc("/", blockRequest(conf))
}

// Parse cli flags to override configuration defaults
func parseFlags(conf *Config) {
	var targetURL string
	flag.StringVar(&targetURL, "o", conf.TargetURL, "Jedox OLAP server address")

	var entryURL string
	flag.StringVar(&entryURL, "i", conf.EntryURL, "xdjprox local port")

	var logFile string
	flag.StringVar(&logFile, "log-file", conf.LogFile, "log file name")

	var logRequest bool
	flag.BoolVar(&logRequest, "log-req", conf.LogRequest, "enable logging of client http request (default false)")

	var logResponse bool
	flag.BoolVar(&logResponse, "log-res", conf.LogResponse, "enable logging of OLAP http response (default false)")

	var logAll bool
	flag.BoolVar(&logAll, "log-all", conf.LogAll, "enable logging everything (default false)")

	var enableWrite bool
	flag.BoolVar(&enableWrite, "w", conf.EnableWrite, "enable write requests (default false)")

	// parse flags
	flag.Parse()

	// put everything required into config
	conf.TargetURL = targetURL
	conf.EntryURL = entryURL
	conf.EnableWrite = enableWrite
	conf.LogFile = logFile

	// everything possible shall be logged
	if logAll {
		logRequest = true
		logResponse = true
	}
	conf.LogRequest = logRequest
	conf.LogResponse = logResponse
}

func main() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// set defaults for Config
	conf := &Config{
		"http://127.0.0.1:7777",
		":8080",
		"2006-02-01 15:04:05",
		false,
		false,
		false,
		"",
		false,
	}

	// Read flags from cli start
	parseFlags(conf)

	// set up log file
	if logFunc := conf.SetupLogFile(); logFunc != nil {
		defer logFunc()
	}

	log.Printf("xdjprox started with config %#v", conf)

	// set up target url
	tgtURL, _ := url.Parse(conf.TargetURL)

	// set-up reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(tgtURL)

	// immediate output of response stream
	proxy.FlushInterval = time.Millisecond * -1

	// register routes for proxy
	registerRoutes(conf, proxy)

	// run proxy
	if err := http.ListenAndServe(conf.EntryURL, nil); err != nil {
		panic(err)
	}
}
