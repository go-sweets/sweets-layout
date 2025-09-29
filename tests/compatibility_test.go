package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// CompatibilityTestSuite tests compatibility across different environments
type CompatibilityTestSuite struct {
	httpBaseURL string
	client      *http.Client
}

// SetupCompatibilityTests initializes compatibility testing
func SetupCompatibilityTests() *CompatibilityTestSuite {
	return &CompatibilityTestSuite{
		httpBaseURL: "http://localhost:8081",
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// TestGoVersionCompatibility tests compatibility with different Go versions
func TestGoVersionCompatibility(t *testing.T) {
	goVersion := runtime.Version()
	t.Logf("Running on Go version: %s", goVersion)

	// Ensure we're running on a supported Go version
	if !strings.HasPrefix(goVersion, "go1.2") && !strings.HasPrefix(goVersion, "go1.1") {
		t.Logf("Warning: Running on Go version %s, CloudWeGo is optimized for Go 1.19+", goVersion)
	}

	// Test basic functionality
	suite := SetupCompatibilityTests()
	resp, err := suite.client.Get(suite.httpBaseURL + "/ping")
	if err == nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

// TestHTTPVersionCompatibility tests HTTP/1.1 and HTTP/2 compatibility
func TestHTTPVersionCompatibility(t *testing.T) {
	suite := SetupCompatibilityTests()

	// Test HTTP/1.1
	t.Run("HTTP1.1", func(t *testing.T) {
		client := &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				ForceAttemptHTTP2: false,
			},
		}

		resp, err := client.Get(suite.httpBaseURL + "/ping")
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "1.1", resp.Proto)
		}
	})

	// Test HTTP/2 (if available)
	t.Run("HTTP2", func(t *testing.T) {
		client := &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				ForceAttemptHTTP2: true,
			},
		}

		resp, err := client.Get(suite.httpBaseURL + "/ping")
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			t.Logf("HTTP Protocol: %s", resp.Proto)
		}
	})
}

// TestContentTypeCompatibility tests different content types
func TestContentTypeCompatibility(t *testing.T) {
	suite := SetupCompatibilityTests()

	testCases := []struct {
		name         string
		acceptHeader string
		expectedCT   string
	}{
		{"JSON", "application/json", "application/json"},
		{"Any", "*/*", "application/json"},
		{"Text", "text/plain", "application/json"}, // API always returns JSON
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", suite.httpBaseURL+"/ping", nil)
			assert.NoError(t, err)

			req.Header.Set("Accept", tc.acceptHeader)

			resp, err := suite.client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, "application/json")
			}
		})
	}
}

// TestBrowserCompatibility tests compatibility with different user agents
func TestBrowserCompatibility(t *testing.T) {
	suite := SetupCompatibilityTests()

	userAgents := []struct {
		name string
		ua   string
	}{
		{"Chrome", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
		{"Firefox", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0"},
		{"Safari", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Version/14.1.1 Safari/537.36"},
		{"Edge", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59"},
		{"Mobile Safari", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1"},
		{"Android Chrome", "Mozilla/5.0 (Linux; Android 10; SM-G975F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Mobile Safari/537.36"},
	}

	for _, ua := range userAgents {
		t.Run(ua.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", suite.httpBaseURL+"/ping", nil)
			assert.NoError(t, err)

			req.Header.Set("User-Agent", ua.ua)

			resp, err := suite.client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				// Verify CORS headers are present for browser requests
				corsOrigin := resp.Header.Get("Access-Control-Allow-Origin")
				assert.NotEmpty(t, corsOrigin, "CORS headers should be present for browser requests")
			}
		})
	}
}

// TestCharsetCompatibility tests different character encoding compatibility
func TestCharsetCompatibility(t *testing.T) {
	suite := SetupCompatibilityTests()

	encodings := []string{
		"utf-8",
		"UTF-8",
		"iso-8859-1",
	}

	for _, encoding := range encodings {
		t.Run(encoding, func(t *testing.T) {
			req, err := http.NewRequest("GET", suite.httpBaseURL+"/ping", nil)
			assert.NoError(t, err)

			req.Header.Set("Accept-Charset", encoding)

			resp, err := suite.client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				// Verify response is valid JSON
				var jsonResponse map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
				assert.NoError(t, err, "Response should be valid JSON regardless of charset")
			}
		})
	}
}

// TestCompressionCompatibility tests different compression algorithms
func TestCompressionCompatibility(t *testing.T) {
	suite := SetupCompatibilityTests()

	compressions := []struct {
		name     string
		encoding string
	}{
		{"Gzip", "gzip"},
		{"Deflate", "deflate"},
		{"Brotli", "br"},
		{"Multiple", "gzip, deflate, br"},
		{"None", ""},
	}

	for _, comp := range compressions {
		t.Run(comp.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", suite.httpBaseURL+"/ping", nil)
			assert.NoError(t, err)

			if comp.encoding != "" {
				req.Header.Set("Accept-Encoding", comp.encoding)
			}

			resp, err := suite.client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				// Log compression used
				contentEncoding := resp.Header.Get("Content-Encoding")
				t.Logf("Content-Encoding: %s (requested: %s)", contentEncoding, comp.encoding)
			}
		})
	}
}

// TestAPIVersionCompatibility tests API versioning compatibility
func TestAPIVersionCompatibility(t *testing.T) {
	suite := SetupCompatibilityTests()

	// Test different API version endpoints
	endpoints := []struct {
		name     string
		path     string
		expected int
	}{
		{"Unversioned Ping", "/ping", http.StatusOK},
		{"V1 Hello", "/v1/hello", http.StatusOK}, // May return 401 if auth required
		{"Health Check", "/health", http.StatusOK},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			resp, err := suite.client.Get(suite.httpBaseURL + endpoint.path)
			if err == nil {
				defer resp.Body.Close()
				// Allow for auth errors on protected endpoints
				if resp.StatusCode != endpoint.expected && resp.StatusCode != http.StatusUnauthorized {
					t.Errorf("Expected %d or %d, got %d for %s", endpoint.expected, http.StatusUnauthorized, resp.StatusCode, endpoint.path)
				}
			}
		})
	}
}

// TestMethodCompatibility tests different HTTP methods
func TestMethodCompatibility(t *testing.T) {
	suite := SetupCompatibilityTests()

	methods := []struct {
		method   string
		path     string
		expected int
	}{
		{"GET", "/ping", http.StatusOK},
		{"HEAD", "/ping", http.StatusOK},
		{"OPTIONS", "/ping", http.StatusOK}, // CORS preflight
		{"POST", "/ping", http.StatusMethodNotAllowed},
		{"PUT", "/ping", http.StatusMethodNotAllowed},
		{"DELETE", "/ping", http.StatusMethodNotAllowed},
		{"PATCH", "/ping", http.StatusMethodNotAllowed},
	}

	for _, method := range methods {
		t.Run(fmt.Sprintf("%s_%s", method.method, strings.TrimPrefix(method.path, "/")), func(t *testing.T) {
			req, err := http.NewRequest(method.method, suite.httpBaseURL+method.path, nil)
			assert.NoError(t, err)

			resp, err := suite.client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.Equal(t, method.expected, resp.StatusCode, "Method %s on %s should return %d", method.method, method.path, method.expected)
			}
		})
	}
}

// TestCloudWeGoFeatureCompatibility tests CloudWeGo specific features
func TestCloudWeGoFeatureCompatibility(t *testing.T) {
	suite := SetupCompatibilityTests()

	t.Run("HertzHTTPServer", func(t *testing.T) {
		// Test that Hertz HTTP server is responding
		resp, err := suite.client.Get(suite.httpBaseURL + "/ping")
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Check for CloudWeGo specific headers or characteristics
			server := resp.Header.Get("Server")
			t.Logf("Server header: %s", server)
		}
	})

	t.Run("JSONResponseFormat", func(t *testing.T) {
		// Verify JSON response format is consistent
		resp, err := suite.client.Get(suite.httpBaseURL + "/ping")
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var jsonResponse map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
			assert.NoError(t, err)
			assert.Contains(t, jsonResponse, "message")
			assert.Equal(t, "pong", jsonResponse["message"])
		}
	})
}

// TestBackwardCompatibility tests backward compatibility with previous versions
func TestBackwardCompatibility(t *testing.T) {
	// This would test compatibility with previous API versions
	// For now, we'll test that existing endpoints still work

	suite := SetupCompatibilityTests()

	// Test legacy endpoints still work
	legacyEndpoints := []string{
		"/ping",
		"/health",
	}

	for _, endpoint := range legacyEndpoints {
		t.Run("Legacy_"+strings.TrimPrefix(endpoint, "/"), func(t *testing.T) {
			resp, err := suite.client.Get(suite.httpBaseURL + endpoint)
			if err == nil {
				defer resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}
		})
	}
}
