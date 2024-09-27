package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
)

func listenAndServe(port int, handler func(httpResponseWriter http.ResponseWriter, httpRequest *http.Request)) {
	logger.Info("Starting http listener.",
		slog.Int("port", port))

	http.HandleFunc("/", handler)
	portStr := ":" + strconv.Itoa(port)
	err := http.ListenAndServe(portStr, nil)
	if err != nil {
		logger.Error("Could not start http listener.",
			slog.Any("err", err))
		panic(err)
	}
}
