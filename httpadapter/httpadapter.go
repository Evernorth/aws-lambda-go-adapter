/*
Package httpadapter makes it easy to test a Golang AWS Lambda outside of AWS, without needing Docker, by providing HTTP trigger support.

See documentation: https://pkg.go.dev/github.com/Evernorth/aws-lambda-go-adapter
*/
package httpadapter

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"log/slog"
	"reflect"
)

type handlerType int

const (
	apigwV2HandlerType handlerType = 0
	albHandlerType     handlerType = 1
	apigwHandlerType   handlerType = 2
)

var (
	logger               *slog.Logger
	delegateValue        reflect.Value
	delegateHandlerType  handlerType
	contextRequired      bool
	inputEventIsPointer  bool
	outputEventIsPointer bool
)

func init() {
	logger = slog.Default()
}

func SetLogger(l *slog.Logger) {
	logger = l
}

type adapterOptions struct {
	//baseContext                      context.Context
	enableSIGTERM bool
	sigtermFuncs  []func()
}

type Option func(*adapterOptions)

// Start starts the HTTP server on the specified port and listens for incoming requests.  When a request is received,
// it is converted to the appropriate Lambda event type and passed to the handler function.  The response from the
// handler function is then converted to an HTTP response and returned to the client.
// See StartWithOptions for other start options.
func Start(port int, handler interface{}) {
	StartWithOptions(port, handler)
}

// StartWithOptions starts the HTTP server on the specified port and listens for incoming requests.  When a request is received,
// it is converted to the appropriate Lambda event type and passed to the handler function.  The response from the
// handler function is then converted to an HTTP response and returned to the client.
func StartWithOptions(port int, handler interface{}, options ...Option) {

	reflectHandler(handler)

	opts := &adapterOptions{}
	for _, option := range options {
		option(opts)
	}

	switch delegateHandlerType {
	case apigwV2HandlerType:
		listenAndServe(port, handleRequestForApigwV2, opts)
	case albHandlerType:
		listenAndServe(port, handleRequestForAlb, opts)
	case apigwHandlerType:
		listenAndServe(port, handleRequestForApigw, opts)
	default:
		panic("unsupported handler type")
	}
}

// WithEnableSIGTERM enables SIGTERM behavior with the HTTP server for graceful shutdown with the optional handler function(s).
// Graceful shutdown of the HTTP server before the provided functions are invoked with a ~500ms timeout. If the functions
// do not complete within the timeout limit, a "SIGKILL" panic will occur. This enables testing in a local environment,
// mimicking AWS Lambda's SIGTERM and SIGKILL behavior.
//
// Usage:
//
//	httpadapter.StartWithOptions(8080, Handler,
//		lambda.WithEnableSIGTERM(func() {
//			log.Print("Cleaning up application components...")
//		})
//	)
func WithEnableSIGTERM(sigtermFuncs ...func()) Option {
	return Option(func(opts *adapterOptions) {
		opts.sigtermFuncs = append(opts.sigtermFuncs, sigtermFuncs...)
		opts.enableSIGTERM = true
	})
}

// reflectHandler reflects the handler function to determine the input and output event types, and whether a context is
// required.  It panics if the function signature is not supported.
func reflectHandler(handler interface{}) {

	handlerType := reflect.TypeOf(handler)

	// Function validation
	if handlerType.Kind() != reflect.Func || handlerType.NumIn() > 2 ||
		handlerType.NumOut() != 2 || handlerType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		panic("unsupported function signature")
	}

	// Check if the function has a context as the first argument
	inEventIndex := 0
	contextRequired = false
	if handlerType.NumIn() == 2 {
		inEventIndex = 1
		if handlerType.In(0) == reflect.TypeOf((*context.Context)(nil)).Elem() {
			contextRequired = true
		} else {
			panic("unsupported function signature")
		}
	}

	// Reflect function input
	reqEventType := handlerType.In(inEventIndex)
	switch reqEventType {
	case reflect.TypeOf((*events.APIGatewayV2HTTPRequest)(nil)).Elem():
		delegateHandlerType = apigwV2HandlerType
		inputEventIsPointer = false
	case reflect.TypeOf((*events.APIGatewayV2HTTPRequest)(nil)):
		delegateHandlerType = apigwV2HandlerType
		inputEventIsPointer = true
	case reflect.TypeOf((*events.ALBTargetGroupRequest)(nil)).Elem():
		delegateHandlerType = albHandlerType
		inputEventIsPointer = false
	case reflect.TypeOf((*events.ALBTargetGroupRequest)(nil)):
		delegateHandlerType = albHandlerType
		inputEventIsPointer = true
	case reflect.TypeOf((*events.APIGatewayProxyRequest)(nil)).Elem():
		delegateHandlerType = apigwHandlerType
		inputEventIsPointer = false
	case reflect.TypeOf((*events.APIGatewayProxyRequest)(nil)):
		delegateHandlerType = apigwHandlerType
		inputEventIsPointer = true
	default:
		panic("unsupported input event")
	}

	// Reflect function output
	respEventType := handlerType.Out(0)
	switch respEventType {
	case reflect.TypeOf((*events.APIGatewayV2HTTPResponse)(nil)).Elem():
		outputEventIsPointer = false
		if delegateHandlerType != apigwV2HandlerType {
			panic("unsupported output event")
		}
	case reflect.TypeOf((*events.APIGatewayV2HTTPResponse)(nil)):
		outputEventIsPointer = true
		if delegateHandlerType != apigwV2HandlerType {
			panic("unsupported output event")
		}
	case reflect.TypeOf((*events.ALBTargetGroupResponse)(nil)).Elem():
		outputEventIsPointer = false
		if delegateHandlerType != albHandlerType {
			panic("unsupported output event")
		}
	case reflect.TypeOf((*events.ALBTargetGroupResponse)(nil)):
		outputEventIsPointer = true
		if delegateHandlerType != albHandlerType {
			panic("unsupported output event")
		}
	case reflect.TypeOf((*events.APIGatewayProxyResponse)(nil)).Elem():
		outputEventIsPointer = false
		if delegateHandlerType != apigwHandlerType {
			panic("unsupported output event")
		}
	case reflect.TypeOf((*events.APIGatewayProxyResponse)(nil)):
		outputEventIsPointer = true
		if delegateHandlerType != apigwHandlerType {
			panic("unsupported output event")
		}
	default:
		panic("unsupported output event")
	}

	// Reflect the handler
	delegateValue = reflect.ValueOf(handler)

}
