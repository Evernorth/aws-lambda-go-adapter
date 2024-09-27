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

func TestAlbPost(t *testing.T) {

	method := "POST"
	path := "/test"
	reqBody := uuid.NewString()
	contentType := "text/html"

	// Create the mock handler
	respStatusCode := 200
	respBody := "OK"
	reflectHandler(func(ctx context.Context, req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
		assert.Equal(t, method, req.HTTPMethod)
		assert.Equal(t, path, req.Path)
		assert.Equal(t, reqBody, req.Body)

		resp := events.ALBTargetGroupResponse{
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
	handleRequestForAlb(recorder, httpReq)

	assert.Equal(t, respStatusCode, recorder.Code)
	assert.Equal(t, respBody, recorder.Body.String())
	assert.Equal(t, contentType, recorder.Header().Get("Content-Type"))
}

func TestAlbGet(t *testing.T) {

	method := "GET"
	path := "/test"
	contentType := "text/html"

	// Create the mock handler
	respStatusCode := 200
	respBody := "OK"
	reflectHandler(func(ctx context.Context, req *events.ALBTargetGroupRequest) (*events.ALBTargetGroupResponse, error) {
		assert.Equal(t, method, req.HTTPMethod)
		assert.Equal(t, path, req.Path)

		resp := events.ALBTargetGroupResponse{
			StatusCode: respStatusCode,
			Headers:    map[string]string{"Content-Type": contentType},
			Body:       respBody,
		}

		return &resp, nil
	})

	// Run the test
	httpReq := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	handleRequestForAlb(recorder, httpReq)

	assert.Equal(t, respStatusCode, recorder.Code)
	assert.Equal(t, respBody, recorder.Body.String())
	assert.Equal(t, contentType, recorder.Header().Get("Content-Type"))
}

func TestAlbGetNegative(t *testing.T) {

	method := "GET"
	path := "/test"

	// Create the mock handler
	reflectHandler(func(req events.ALBTargetGroupRequest) (*events.ALBTargetGroupResponse, error) {
		assert.Equal(t, method, req.HTTPMethod)
		assert.Equal(t, path, req.Path)

		return nil, errors.New("kaboom")
	})

	// Run the test
	httpReq := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	handleRequestForAlb(recorder, httpReq)

	assert.Equal(t, 500, recorder.Code)
}
