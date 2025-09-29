package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// E2ETestSuite defines the end-to-end test suite
type E2ETestSuite struct {
	suite.Suite
	httpBaseURL string
	rpcAddress  string
	ctx         context.Context
	cancel      context.CancelFunc
}

// SetupSuite initializes the test environment before running tests
func (suite *E2ETestSuite) SetupSuite() {
	// Note: The application should be running separately for E2E tests
	// Run the application with: make run
	// Then run tests with: go test ./tests/...

	suite.httpBaseURL = "http://localhost:8080"
	suite.rpcAddress = "localhost:8888"

	// Create context for test operations
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), time.Minute*5)

	// Wait for application to be ready
	suite.waitForApplicationReady()
}

// TearDownSuite cleans up after all tests are completed
func (suite *E2ETestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// waitForApplicationReady waits for the application to be ready to serve requests
func (suite *E2ETestSuite) waitForApplicationReady() {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(suite.httpBaseURL + "/ping")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Second)
	}
	suite.Fail("Application did not become ready in time")
}

// TestPing tests the ping endpoint
func (suite *E2ETestSuite) TestPing() {
	resp, err := http.Get(suite.httpBaseURL + "/ping")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)
}

// TestHealth tests the health check endpoint
func (suite *E2ETestSuite) TestHealth() {
	resp, err := http.Get(suite.httpBaseURL + "/health")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)
}

// TestHelloEndpoint tests the hello endpoint
func (suite *E2ETestSuite) TestHelloEndpoint() {
	resp, err := http.Get(suite.httpBaseURL + "/hello/1")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)
}

// TestHelloEndpointWithInvalidID tests the hello endpoint with invalid ID
func (suite *E2ETestSuite) TestHelloEndpointWithInvalidID() {
	resp, err := http.Get(suite.httpBaseURL + "/hello/invalid")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	// Should return bad request or similar error
	suite.True(resp.StatusCode >= 400)
}

// TestConcurrentRequests tests handling of concurrent requests
func (suite *E2ETestSuite) TestConcurrentRequests() {
	concurrentRequests := 10
	done := make(chan bool, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func(id int) {
			resp, err := http.Get(fmt.Sprintf("%s/hello/%d", suite.httpBaseURL, id))
			suite.NoError(err)
			if resp != nil {
				resp.Body.Close()
				suite.Equal(http.StatusOK, resp.StatusCode)
			}
			done <- true
		}(i + 1)
	}

	// Wait for all requests to complete
	for i := 0; i < concurrentRequests; i++ {
		select {
		case <-done:
		case <-time.After(time.Second * 10):
			suite.Fail("Concurrent requests timed out")
		}
	}
}

// TestE2E is the entry point for the test suite
func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}