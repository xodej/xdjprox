// Copyright 2019 Johannes Mueller <circus2@web.de>

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// set defaults for Config
	conf := Config{
		"http://127.0.0.1:7777",
		":8080",
		"2006-02-01 15:04:05",
		false,
		false,
		false,
	}

	// override defaults with cli arguments
	var targetURL string
	flag.StringVar(&targetURL, "o", conf.TargetURL, "Jedox OLAP")

	var entryURL string
	flag.StringVar(&entryURL, "i", conf.EntryURL, "xdjproxy port")

	var logRequest bool
	flag.BoolVar(&logRequest, "req", conf.LogRequest, "log http request (default false)")

	var logResponse bool
	flag.BoolVar(&logResponse, "res", conf.LogResponse, "log OLAP http response (default false)")

	var enableWrite bool
	flag.BoolVar(&enableWrite, "w", conf.EnableWrite, "enable write requests (default false)")

	// parse flags
	flag.Parse()

	// put everything required into config
	conf.TargetURL = targetURL
	conf.EntryURL = entryURL
	conf.LogRequest = logRequest
	conf.LogResponse = logResponse
	conf.EnableWrite = enableWrite

	log.Printf("%#v\n", conf)

	// router and server
	handler := http.NewServeMux()
	srv := &http.Server{
		Addr:           conf.EntryURL,
		Handler:        handler,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// set up server wrapper
	s := Server{
		srv,
		handler,
		conf,
	}

	// Load all routes
	s.Register()

	// Shutting down the proxy
	// see example https://golang.org/pkg/net/http/#Server.Shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := s.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v\n", err)
		}
		// Closing channel jumping to <- idleConnsClosed
		close(idleConnsClosed)
	}()

	// Running the server listening on defined port
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server ListenAndServe: %v\n", err)
	}

	<-idleConnsClosed
}
