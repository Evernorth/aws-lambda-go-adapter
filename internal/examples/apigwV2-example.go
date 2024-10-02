package main

import (
	"context"
	"github.com/Evernorth/aws-lambda-go-adapter/httpadapter"
	"github.com/Evernorth/aws-lambda-go-adapter/pkg/util"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
)

// Handler is the Lambda handler function.  It returns a 200 status code with a "Hello, World!" message for GET requests,
// and a 405 status code for all other requests.
func Handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	if request.RequestContext.HTTP.Method == http.MethodGet {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       "Hello, World!",
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusMethodNotAllowed,
		Body:       "Method not allowed",
	}, nil
}

func main() {
	if util.IsLambdaRuntime() {
		lambda.Start(Handler)
	} else {
		httpadapter.Start(8080, Handler)
	}
}
