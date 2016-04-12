package ratelimit

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/nbio/st"
	"gopkg.in/vinxi/utils.v0"
)

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(5, 5)

	var called bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	rw := utils.NewWriterStub()
	req := &http.Request{}

	limiter.LimitHTTP(handler)(rw, req)
	rw.WriteHeader(200)
	rw.Write([]byte("foo"))

	st.Expect(t, called, true)
	st.Expect(t, rw.Code, 200)
	st.Expect(t, string(rw.Body), "foo")
	st.Expect(t, rw.Header().Get("X-RateLimit-Limit"), "5")
	st.Expect(t, rw.Header().Get("X-RateLimit-Remaining"), "4")
	st.Expect(t, rw.Header().Get("X-RateLimit-Reset"), "")
}

func TestRateLimiterExceeded(t *testing.T) {
	limiter := NewRateLimiter(5, 5)

	var calls int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	})

	for i := 0; i < 10; i++ {
		rw := utils.NewWriterStub()
		req := &http.Request{}
		limiter.LimitHTTP(handler)(rw, req)

		st.Expect(t, rw.Header().Get("X-RateLimit-Limit"), "5")

		if i < 5 {
			// Write response as it was valid
			rw.WriteHeader(200)
			rw.Write([]byte("foo"))

			st.Expect(t, calls, i+1)
			st.Expect(t, rw.Code, 200)
			st.Expect(t, string(rw.Body), "foo")
			st.Expect(t, rw.Header().Get("X-RateLimit-Remaining"), strconv.Itoa(5-(i+1)))
			st.Expect(t, rw.Header().Get("X-RateLimit-Reset"), "")
		} else {
			st.Expect(t, calls, 5)
			st.Expect(t, rw.Code, 429)
			st.Expect(t, string(rw.Body), "Too Many Requests")
			st.Expect(t, rw.Header().Get("X-RateLimit-Remaining"), "0")
			st.Expect(t, rw.Header().Get("X-RateLimit-Reset"), "0")
		}
	}
}

func TestRateLimiterPassExceptions(t *testing.T) {
	limiter := NewRateLimiter(5, 5)
	limiter.Exception(func(r *http.Request) bool { return r.Method == "GET" })
	limiter.Exception(func(r *http.Request) bool { return r.Method == "PUT" })

	var called bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	// Pass exceptions
	rw := utils.NewWriterStub()
	req := &http.Request{Method: "GET"}

	limiter.LimitHTTP(handler)(rw, req)
	rw.WriteHeader(200)
	rw.Write([]byte("foo"))

	st.Expect(t, called, true)
	st.Expect(t, rw.Code, 200)
	st.Expect(t, string(rw.Body), "foo")
	st.Expect(t, rw.Header().Get("X-RateLimit-Limit"), "")
	st.Expect(t, rw.Header().Get("X-RateLimit-Remaining"), "")
	st.Expect(t, rw.Header().Get("X-RateLimit-Reset"), "")

	// Do not pass exceptions
	rw = utils.NewWriterStub()
	req = &http.Request{Method: "POST"}

	limiter.LimitHTTP(handler)(rw, req)
	rw.WriteHeader(200)
	rw.Write([]byte("foo"))

	st.Expect(t, called, true)
	st.Expect(t, rw.Code, 200)
	st.Expect(t, string(rw.Body), "foo")
	st.Expect(t, rw.Header().Get("X-RateLimit-Limit"), "5")
	st.Expect(t, rw.Header().Get("X-RateLimit-Remaining"), "4")
	st.Expect(t, rw.Header().Get("X-RateLimit-Reset"), "")
}

func TestRateLimiterPassFilters(t *testing.T) {
	limiter := NewRateLimiter(5, 5)
	limiter.Filter(func(r *http.Request) bool { return r.Method == "GET" })
	limiter.Filter(func(r *http.Request) bool { return r.URL.Path == "/" })

	var called bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	// Pass filters
	rw := utils.NewWriterStub()
	req := &http.Request{Method: "GET", URL: &url.URL{Path: "/"}}

	limiter.LimitHTTP(handler)(rw, req)
	rw.WriteHeader(200)
	rw.Write([]byte("foo"))

	st.Expect(t, called, true)
	st.Expect(t, rw.Code, 200)
	st.Expect(t, string(rw.Body), "foo")
	st.Expect(t, rw.Header().Get("X-RateLimit-Limit"), "5")
	st.Expect(t, rw.Header().Get("X-RateLimit-Remaining"), "4")
	st.Expect(t, rw.Header().Get("X-RateLimit-Reset"), "")

	// Do not pass filters
	rw = utils.NewWriterStub()
	req = &http.Request{Method: "POST"}

	limiter.LimitHTTP(handler)(rw, req)
	rw.WriteHeader(200)
	rw.Write([]byte("foo"))

	st.Expect(t, called, true)
	st.Expect(t, rw.Code, 200)
	st.Expect(t, string(rw.Body), "foo")
	st.Expect(t, rw.Header().Get("X-RateLimit-Limit"), "")
	st.Expect(t, rw.Header().Get("X-RateLimit-Remaining"), "")
	st.Expect(t, rw.Header().Get("X-RateLimit-Reset"), "")
}

func TestRateLimiterResponder(t *testing.T) {
	limiter := NewTimeLimiter(time.Second, 1)
	limiter.SetResponder(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte("wait, dude"))
	})

	var called bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	rw := utils.NewWriterStub()
	req := &http.Request{}

	limiter.LimitHTTP(handler)(rw, req)
	st.Expect(t, called, true)

	called = false
	limiter.LimitHTTP(handler)(rw, req)
	st.Expect(t, called, false)
	st.Expect(t, rw.Code, 429)
	st.Expect(t, string(rw.Body), "wait, dude")
	st.Expect(t, rw.Header().Get("X-RateLimit-Limit"), "1")
	st.Expect(t, rw.Header().Get("X-RateLimit-Remaining"), "0")
	st.Expect(t, rw.Header().Get("X-RateLimit-Reset"), "0")
}
