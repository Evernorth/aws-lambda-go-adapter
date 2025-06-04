package httpadapter

import (
	"context"
	"log/slog"
	"net/http"
	"os"
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

	if !opts.enableSIGTERM {
		defaultListenAndServe(server) //port, handler)
	} else {
		sigtermListenAndServe(server, opts)
	}
}

func defaultListenAndServe(server *http.Server) { //port int, handler func(httpResponseWriter http.ResponseWriter, httpRequest *http.Request)) {

	//http.HandleFunc("/", handler)
	//portStr := ":" + strconv.Itoa(port)
	//err := http.ListenAndServe(portStr, nil)
	//err := server.ListenAndServe()
	//if err != nil {
	//	logger.Error("Could not start http listener.",
	//		slog.Any("err", err))
	//	panic(err)
	//}
	if err := server.ListenAndServe(); err != nil { //&& err != http.ErrServerClosed {
		logger.Error("Could not start http listener.", slog.Any("err", err))
		panic(err)
	}
}

func sigtermListenAndServe(server *http.Server, opts *adapterOptions) {

	registerSigtermFuncs(opts.sigtermFuncs)

	// Create a context that cancels on SIGTERM or SIGINT
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	//http.HandleFunc("/", handler)
	//portStr := ":" + strconv.Itoa(port)
	//server := &http.Server{
	//	Addr:    portStr,
	//	Handler: nil, // Use the typical defaultServerMux with the handler set by http.HandleFunc
	//}
	go func() {
		//if err := server.ListenAndServe(); err != nil { //&& err != http.ErrServerClosed {
		//	logger.Error("Could not start http listener.", slog.Any("err", err))
		//	panic(err)
		//}
		defaultListenAndServe(server)
	}()

	// Wait for the termination signal
	<-sigCtx.Done()
	logger.Info("SIGTERM: Shutting server down gracefully...")

	// Perform cleanup with a timeout with a 500ms limit like AWS Lambda SIGTERM behavior
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Graceful server shutdown failed", slog.Any("err", err))
	}

	logger.Info("Server shutdown gracefully")

	//TODO: Verify app cleanup is done after the server is shutdown.
	// If not, will need to call cleanupFuncs here.
	logger.Info("Application cleanup completed")
}

// registerSigtermFuncs configures an optional list of sigtermHandlers to run on HTTP Server shutdown.
func registerSigtermFuncs(sigtermFuncs []func()) {
	// optionally register SIGTERM handlers
	if len(sigtermFuncs) > 0 {
		signaled := make(chan os.Signal, 1)
		signal.Notify(signaled, syscall.SIGTERM)
		go func() {
			<-signaled
			for _, f := range sigtermFuncs {
				f()
			}
		}()
	}
}
