package performance

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

// CompressionMiddleware returns gzip compression middleware
func CompressionMiddleware(cfg *viper.Viper) echo.MiddlewareFunc {
	level := cfg.GetInt("performance.compression_level")
	if level == 0 {
		level = gzip.DefaultCompression
	}

	minSize := cfg.GetInt("performance.compression_min_size")
	if minSize == 0 {
		minSize = 1024 // 1KB
	}

	return middleware.GzipWithConfig(middleware.GzipConfig{
		Level: level,
		MinLength: minSize,
		Skipper: func(c echo.Context) bool {
			// Skip compression for already compressed content
			contentType := c.Response().Header().Get("Content-Type")
			return strings.Contains(contentType, "image/") ||
				strings.Contains(contentType, "video/") ||
				strings.Contains(contentType, "audio/") ||
				strings.HasSuffix(c.Request().URL.Path, ".woff") ||
				strings.HasSuffix(c.Request().URL.Path, ".woff2")
		},
	})
}

// CacheControlMiddleware adds cache headers for static resources
func CacheControlMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			// Set cache headers based on resource type
			if strings.HasPrefix(path, "/static/") {
				// Static assets - cache for 1 year
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else if strings.HasPrefix(path, "/api/") {
				// API responses - no cache
				c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				c.Response().Header().Set("Pragma", "no-cache")
				c.Response().Header().Set("Expires", "0")
			} else if strings.HasSuffix(path, ".html") || path == "/" {
				// HTML pages - cache for 5 minutes
				c.Response().Header().Set("Cache-Control", "public, max-age=300")
			}

			return next(c)
		}
	}
}

// ETagger generates ETags for responses
type ETagger struct {
	cache sync.Map
}

// NewETagger creates a new ETag generator
func NewETagger() *ETagger {
	return &ETagger{}
}

// Middleware returns ETag middleware
func (e *ETagger) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip for non-GET requests
			if c.Request().Method != http.MethodGet {
				return next(c)
			}

			// Create custom response writer to capture response
			rec := &responseRecorder{
				ResponseWriter: c.Response().Writer,
				body:           new(bytes.Buffer),
			}
			c.Response().Writer = rec

			// Process request
			if err := next(c); err != nil {
				return err
			}

			// Generate ETag from response body
			if rec.body.Len() > 0 && c.Response().Status == http.StatusOK {
				etag := generateETag(rec.body.Bytes())
				c.Response().Header().Set("ETag", etag)

				// Check if client has matching ETag
				if match := c.Request().Header.Get("If-None-Match"); match == etag {
					return c.NoContent(http.StatusNotModified)
				}

				// Write the captured body
				_, err := c.Response().Write(rec.body.Bytes())
				return err
			}

			return nil
		}
	}
}

// responseRecorder captures response body for ETag generation
type responseRecorder struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// generateETag generates an ETag from content
func generateETag(content []byte) string {
	// Simple ETag generation using content length and first/last bytes
	// In production, use a proper hash like MD5 or SHA1
	if len(content) == 0 {
		return `"0"`
	}
	
	return `"` + fmt.Sprintf("%d", len(content)) + `-` + 
		string(content[0]) + string(content[len(content)-1]) + `"`
}

// RequestCoalescingMiddleware prevents duplicate concurrent requests
type RequestCoalescer struct {
	mu       sync.Mutex
	inflight map[string]*coalescedRequest
}

type coalescedRequest struct {
	done   chan struct{}
	result interface{}
	err    error
}

// NewRequestCoalescer creates a request coalescer
func NewRequestCoalescer() *RequestCoalescer {
	return &RequestCoalescer{
		inflight: make(map[string]*coalescedRequest),
	}
}

// Middleware returns request coalescing middleware
func (rc *RequestCoalescer) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only coalesce GET requests
			if c.Request().Method != http.MethodGet {
				return next(c)
			}

			// Generate request key
			key := c.Request().URL.String()

			rc.mu.Lock()
			if req, exists := rc.inflight[key]; exists {
				rc.mu.Unlock()
				
				// Wait for inflight request to complete
				<-req.done
				
				if req.err != nil {
					return req.err
				}
				
				// Return cached result
				return c.JSON(http.StatusOK, req.result)
			}

			// Create new inflight request
			req := &coalescedRequest{
				done: make(chan struct{}),
			}
			rc.inflight[key] = req
			rc.mu.Unlock()

			// Execute request
			err := next(c)
			
			// Store result and notify waiters
			req.err = err
			close(req.done)

			// Cleanup
			rc.mu.Lock()
			delete(rc.inflight, key)
			rc.mu.Unlock()

			return err
		}
	}
}

// ResponseBufferingMiddleware buffers small responses to set Content-Length
func ResponseBufferingMiddleware(maxSize int) echo.MiddlewareFunc {
	if maxSize == 0 {
		maxSize = 10 * 1024 // 10KB default
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create buffered response writer
			w := &bufferedResponseWriter{
				ResponseWriter: c.Response().Writer,
				buffer:         new(bytes.Buffer),
				maxSize:        maxSize,
			}
			c.Response().Writer = w

			// Process request
			if err := next(c); err != nil {
				return err
			}

			// Flush buffer if needed
			return w.Flush()
		}
	}
}

// bufferedResponseWriter buffers response up to maxSize
type bufferedResponseWriter struct {
	http.ResponseWriter
	buffer     *bytes.Buffer
	maxSize    int
	unbuffered bool
}

func (w *bufferedResponseWriter) Write(p []byte) (int, error) {
	if w.unbuffered {
		return w.ResponseWriter.Write(p)
	}

	// Check if adding this would exceed max size
	if w.buffer.Len()+len(p) > w.maxSize {
		// Switch to unbuffered mode
		w.unbuffered = true
		
		// Write buffered content
		if w.buffer.Len() > 0 {
			if _, err := w.ResponseWriter.Write(w.buffer.Bytes()); err != nil {
				return 0, err
			}
		}
		
		// Write current chunk
		return w.ResponseWriter.Write(p)
	}

	// Buffer the content
	return w.buffer.Write(p)
}

func (w *bufferedResponseWriter) Flush() error {
	if !w.unbuffered && w.buffer.Len() > 0 {
		// Set Content-Length header
		w.Header().Set("Content-Length", fmt.Sprintf("%d", w.buffer.Len()))
		
		// Write buffered content
		_, err := w.ResponseWriter.Write(w.buffer.Bytes())
		return err
	}
	return nil
}