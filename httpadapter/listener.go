package httpadapter

import (
	"context"
	"log/slog"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func listenAndServe(port int, handler func(httpResponseWriter http.ResponseWriter, httpRequest *http.Request), opts *adapterOptions) {
	logger.Info("Starting http listener.",
		slog.Int("port", port))

	http.HandleFunc("/", handler)
	portStr := ":" + strconv.Itoa(port)
	server := &http.Server{
		Addr:    portStr,
		Handler: nil, // Use the typical defaultServerMux with the handler set by http.HandleFunc
	}

	if opts.enableSIGTERM {
		sigtermListenAndServe(server, opts)
	} else {
		defaultListenAndServe(server)
	}
}

func defaultListenAndServe(server *http.Server) {
	// start the server and ignore the error if it is http.ErrServerClosed from Shutdown
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Could not start http listener.", slog.Any("err", err))
		panic(err)
	}
}

func sigtermListenAndServe(server *http.Server, opts *adapterOptions) {
	// Create a context that cancels on SIGTERM or SIGINT
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Start the server in a goroutine
	go func() {
		defaultListenAndServe(server)
	}()

	// Wait for the termination signal
	<-sigCtx.Done()
	logger.Info("SIGTERM: Shutting server down gracefully...")

	// Create a context with a timeout for a limited clean server shutdown (5s should be plenty)
	// Shutdown usually completes much faster, of course.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("SIGTERM: Graceful server shutdown failed", slog.Any("err", err))
	}
	logger.Info("SIGTERM: Server shutdown gracefully")

	// Perform cleanup with a 500ms limit like AWS Lambda SIGTERM behavior. This will help ensure
	// that cleanup functions can be locally tested with the same behavior as in AWS Lambda.
	if len(opts.sigtermFuncs) > 0 {
		logger.Info("SIGTERM: Running shutdown functions (mimicking the 500ms AWS Lambda limit)")
		timedCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		done := make(chan struct{})
		go func() {
			for _, f := range opts.sigtermFuncs {
				f()
			}
			close(done)
		}()
		select {
		case <-done:
			logger.Info("SIGTERM: Shutdown functions completed")
		case <-timedCtx.Done():
			//Panic is more attention-grabbing than an error log.
			panic("SIGKILL: Shutdown functions did not complete within the AWS Lambda limit")
		}
	}

	logger.Info("SIGTERM: Completed, exiting...")
}
