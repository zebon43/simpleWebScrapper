package analyzer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

// TestDetectHTMLVersion verifies that the custom doctype parser correctly identifies HTML5 and older versions.
func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "HTML5 Doctype",
			html:     `<!DOCTYPE html><html></html>`,
			expected: "HTML5",
		},
		{
			name:     "HTML 4.01 Doctype",
			html:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd"><html></html>`,
			expected: `-//W3C//DTD HTML 4.01//EN`,
		},
		{
			name:     "No Doctype",
			html:     `<html></html>`,
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse test HTML: %v", err)
			}

			result := detectHTMLVersion(node)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestCheckAccessibility uses a mock HTTP local server to test parallel link scanning.
func TestCheckAccessibility(t *testing.T) {
	// Create a mock server to simulate web endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/good" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/bad" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	links := []string{
		server.URL + "/good",                 // Accessible
		server.URL + "/bad",                  // Inaccessible (404)
		"http://localhost:9999/doesnotexist", // Inaccessible (Connection refused)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// We expect 2 links to be inaccessible (/bad and the refused connection)
	expectedInaccessible := 2
	result := checkAccessibility(ctx, links)

	if result != expectedInaccessible {
		t.Errorf("expected %d inaccessible links, got %d", expectedInaccessible, result)
	}
}

// TestAnalyzeFullPage tests the overall orchestrator by mocking an entire target website.
func TestAnalyzeFullPage(t *testing.T) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>Test Target Page</title></head>
			<body>
				<h1>Main Heading</h1>
				<h2>Subheading 1</h2>
				<h2>Subheading 2</h2>
				<a href="https://google.com">External Link</a>
				<a href="/internal-route">Internal Link</a>
				<form>
					<input type="text" name="username">
					<input type="password" name="pwd">
				</form>
			</body>
			</html>
		`))
	}))
	defer targetServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := Analyze(ctx, targetServer.URL)
	if err != nil {
		t.Fatalf("Analyze failed unexpectedly: %v", err)
	}

	if result.HTMLVersion != "HTML5" {
		t.Errorf("Expected HTML5, got %s", result.HTMLVersion)
	}

	if result.Title != "Test Target Page" {
		t.Errorf("Expected title 'Test Target Page', got '%s'", result.Title)
	}

	if result.Headings["h1"] != 1 {
		t.Errorf("Expected 1 h1, got %d", result.Headings["h1"])
	}
	if result.Headings["h2"] != 2 {
		t.Errorf("Expected 2 h2s, got %d", result.Headings["h2"])
	}

	if result.ExternalLinks != 1 {
		t.Errorf("Expected 1 external link, got %d", result.ExternalLinks)
	}
	if result.InternalLinks != 1 {
		t.Errorf("Expected 1 internal link, got %d", result.InternalLinks)
	}

	if !result.HasLoginForm {
		t.Errorf("Expected HasLoginForm to be true, got false")
	}
}

// TestAnalyzeErrors ensures correct HTTP error strings surface gracefully.
func TestAnalyzeErrors(t *testing.T) {
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer errorServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Analyze(ctx, errorServer.URL)
	if err == nil {
		t.Fatal("Expected an error for HTTP 403 response, but got nil")
	}

	if !strings.Contains(err.Error(), "HTTP 403") {
		t.Errorf("Expected error to contain HTTP 403, got: %v", err)
	}
}
