package httpadapter

import (
	"github.com/aws/aws-lambda-go/events"
	"io"
	"log/slog"
	"net/http"
	"reflect"
)

func handleRequestForApigwV2(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	var err error

	// Read the content, if any
	var content string
	var contentBytes []byte
	if httpRequest.ContentLength > 0 {
		contentBytes, err = io.ReadAll(httpRequest.Body)
		if err != nil {
			msg := "Could not read content."
			logger.Error(msg, slog.Any("err", err))
			http.Error(httpResponseWriter, msg, http.StatusInternalServerError)
			return
		}
		content = string(contentBytes)
	}

	// Build the lambda request
	lambdaReq := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: httpRequest.Method,
				Path:   httpRequest.RequestURI,
			},
		},
		RawPath: httpRequest.RequestURI,
		Body:    content,
		Headers: formatHeaders(httpRequest),
	}

	// Create the input values
	inValues := make([]reflect.Value, 0)
	if contextRequired {
		inValues = append(inValues, reflect.ValueOf(httpRequest.Context()))
	}
	if inputEventIsPointer {
		inValues = append(inValues, reflect.ValueOf(&lambdaReq))
	} else {
		inValues = append(inValues, reflect.ValueOf(lambdaReq))
	}

	// Invoke the lambda function
	outValues := delegateValue.Call(inValues)

	// Check for an error
	if outValues[1].Interface() != nil {
		err = outValues[1].Interface().(error)
		msg := "Could not invoke lambda."
		logger.Error(msg, slog.Any("err", err))
		http.Error(httpResponseWriter, msg, http.StatusInternalServerError)
		return
	}

	// Return the response
	var lambdaResp *events.APIGatewayV2HTTPResponse
	if outputEventIsPointer {
		lambdaResp = outValues[0].Interface().(*events.APIGatewayV2HTTPResponse)
	} else {
		valLambdaResp := outValues[0].Interface().(events.APIGatewayV2HTTPResponse)
		lambdaResp = &valLambdaResp
	}
	for key, value := range lambdaResp.Headers {
		httpResponseWriter.Header().Add(key, value)
	}
	httpResponseWriter.WriteHeader(lambdaResp.StatusCode)
	_, err = httpResponseWriter.Write([]byte(lambdaResp.Body))
	if err != nil {
		logger.Error("Could not write response.", slog.Any("err", err))
	}
}
