# aws-lambda-go-adapter

[![Go Report Card](https://goreportcard.com/badge/github.com/Evernorth/aws-lambda-go-adapter)](https://goreportcard.com/report/github.com/Evernorth/aws-lambda-go-adapter)
[![GoDoc](https://godoc.org/github.com/Evernorth/aws-lambda-go-adapter?status.svg)](https://godoc.org/github.com/Evernorth/aws-lambda-go-adapter)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/Evernorth/aws-lambda-go-adapter)](https://github.com/Evernorth/aws-lambda-go-adapter/releases)

## Description
This module makes it easy to test a Golang AWS Lambda outside of AWS, without needing Docker.  Currently, 
only HTTP triggers are supported, but additional triggers may be supported in the future.

### How it works
Lambda Function in AWS:
1. The main function calls the lambda.Start function, passing in the Handler function.
2. AWS Lambda invokes the Handler function when an HTTP request is received.
![diagram1](docs/images/diagram1.png)

Testing a Lambda Function outside of AWS:
1. The main function calls the httpadapter.Start function, passing in the Handler function and the port number to listen on.
2. The httpadapter listens for incoming HTTP requests on the specified port and invokes the Handler function when a request is received.
![diagram2](docs/images/diagram2.png)

   
## Features
* Supports APIGatewayV2HTTP, APIGatewayProxy, and ALBTargetGroup events.
* Supports handler functions with and without context.Context parameters.
* Supports both values and pointers for handler function request events.
* Supports both values and pointers for handler function response events.

## Installation
```go get -u github.com/Evernorth/aws-lambda-go-adapter```

## Usage
```
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
```

## Dependencies
See the [go.mod](go.mod) file.

## Support
If you have questions, concerns, bug reports, etc. See [CONTRIBUTING](CONTRIBUTING.md).

## License
aws-lambda-go-adapter is open source software released under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0.html).

## Original Contributors
- Steve Sefton, Evernorth
- Ben Lilley, Evernorth