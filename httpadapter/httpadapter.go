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

// Start starts the HTTP server on the specified port and listens for incoming requests.  When a request is received,
// it is converted to the appropriate Lambda event type and passed to the handler function.  The response from the
// handler function is then converted to an HTTP response and returned to the client.
func Start(port int, handler interface{}) {
	reflectHandler(handler)

	switch delegateHandlerType {
	case apigwV2HandlerType:
		listenAndServe(port, handleRequestForApigwV2)
	case albHandlerType:
		listenAndServe(port, handleRequestForAlb)
	case apigwHandlerType:
		listenAndServe(port, handleRequestForApigw)
	default:
		panic("unsupported handler type")
	}
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
