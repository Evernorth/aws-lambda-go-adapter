package httpadapter

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApigwV2Post(t *testing.T) {

	method := "POST"
	path := "/test"
	reqBody := uuid.NewString()
	contentType := "text/plain"

	// Create the mock handler
	respStatusCode := 200
	respBody := "OK"
	reflectHandler(func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		assert.Equal(t, method, req.RequestContext.HTTP.Method)
		assert.Equal(t, path, req.RequestContext.HTTP.Path)
		assert.Equal(t, reqBody, req.Body)

		resp := events.APIGatewayV2HTTPResponse{
			StatusCode: respStatusCode,
			Headers:    map[string]string{"Content-Type": contentType},
			Body:       respBody,
		}

		return resp, nil
	})

	// Run the test
	httpReq := httptest.NewRequest(method, path, strings.NewReader(reqBody))
	httpReq.Header.Add("Content-Type", contentType)
	recorder := httptest.NewRecorder()
	handleRequestForApigwV2(recorder, httpReq)

	assert.Equal(t, respStatusCode, recorder.Code)
	assert.Equal(t, respBody, recorder.Body.String())
	assert.Equal(t, contentType, recorder.Header().Get("Content-Type"))
}

func TestApigwV2Get(t *testing.T) {

	method := "GET"
	path := "/test"
	contentType := "text/plain"

	// Create the mock handler
	respStatusCode := 200
	respBody := "OK"
	reflectHandler(func(ctx context.Context, req *events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
		assert.Equal(t, method, req.RequestContext.HTTP.Method)
		assert.Equal(t, path, req.RequestContext.HTTP.Path)

		resp := events.APIGatewayV2HTTPResponse{
			StatusCode: respStatusCode,
			Headers:    map[string]string{"Content-Type": contentType},
			Body:       respBody,
		}

		return &resp, nil
	})

	// Run the test
	httpReq := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	handleRequestForApigwV2(recorder, httpReq)

	assert.Equal(t, respStatusCode, recorder.Code)
	assert.Equal(t, respBody, recorder.Body.String())
	assert.Equal(t, contentType, recorder.Header().Get("Content-Type"))
}

func TestApigwV2GetNegative(t *testing.T) {

	method := "GET"
	path := "/test"

	// Create the mock handler
	reflectHandler(func(req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
		assert.Equal(t, method, req.RequestContext.HTTP.Method)
		assert.Equal(t, path, req.RequestContext.HTTP.Path)

		return nil, errors.New("kaboom")
	})

	// Run the test
	httpReq := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	handleRequestForApigwV2(recorder, httpReq)

	assert.Equal(t, 500, recorder.Code)
}
