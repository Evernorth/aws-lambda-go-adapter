package httpadapter

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestSigtermListenAndServe(t *testing.T) {

	opts := &adapterOptions{
		enableSIGTERM: true,
	}
	sigtermListenAndServeTest(t, opts, false)
}

func TestSigtermListenAndServe_WithFuncs(t *testing.T) {

	opts := &adapterOptions{
		enableSIGTERM: true,
		sigtermFuncs: []func(){
			func() {
				t.Log("Running cleanup function 1")
			},
			func() {
				t.Log("Running cleanup function 2")
			},
		},
	}
	sigtermListenAndServeTest(t, opts, false)
}

func TestSigtermListenAndServe_ExpectFuncTimeout(t *testing.T) {
	opts := &adapterOptions{
		enableSIGTERM: true,
		sigtermFuncs: []func(){
			func() {
				t.Log("Long running cleanup function")
				time.Sleep(502 * time.Millisecond) // Simulate a long-running function (>500ms)
			},
		},
	}
	sigtermListenAndServeTest(t, opts, true)
}

func sigtermListenAndServeTest(t *testing.T, opts *adapterOptions, expectPanic bool) {
	// Create a dummy HTTP server that does nothing, just to test the SIGTERM handling
	server := &http.Server{
		Addr: ":0", // Use a port no one should care about
	}

	// Create a channel to signal when the server has started
	done := make(chan struct{})
	go func() {
		// Use defer to ensure that we recover from a timeout panic to assert
		// behavior of AWS Lambda where cleanup functions take too long
		defer func() {
			if r := recover(); r == nil && expectPanic {
				t.Fatal("Expected panic due to cleanup function timeout, but no panic occurred")
			}
		}()

		sigtermListenAndServe(server, opts)
		close(done)
	}()

	// Wait for the server to start
	time.Sleep(20 * time.Millisecond)

	// Send SIGINT to self
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)

	// Wait for the server to exit or timeout
	select {
	case <-done:
		t.Log("server successfully exited on SIGINT")
	case <-time.After(1 * time.Second):
		if !expectPanic {
			t.Fatal("server did not exit on SIGINT")
		}
	}
}
