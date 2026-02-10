package web

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkContextPool(b *testing.B) {
	e := New()
	e.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "hello")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		e.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkContextWithoutPool(b *testing.B) {
	e := New()
	e.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "hello")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c := &Context{
			Writer: w,
			Request: req,
		}
		e.router.handleRequest(c)
		w.Body.Reset()
	}
}

func BenchmarkCopyWithPool(b *testing.B) {
	buf := make([]byte, 32*1024)
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	reader := bytes.NewReader(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader.Seek(0, io.SeekStart)
		_, err := io.CopyBuffer(io.Discard, reader, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCopyWithoutPool(b *testing.B) {
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	reader := bytes.NewReader(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader.Seek(0, io.SeekStart)
		_, err := io.Copy(io.Discard, reader)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteString(b *testing.B) {
	e := New()
	e.GET("/test", func(c *Context) {
		for i := 0; i < 100; i++ {
			c.WriteString("hello world ")
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		e.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkWriteJSON(b *testing.B) {
	e := New()
	e.GET("/test", func(c *Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"message": "hello",
			"code":    200,
			"data": map[string]interface{}{
				"id":   1,
				"name": "test",
			},
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		e.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkResponseRecorder(b *testing.B) {
	e := New()
	e.GET("/test", func(c *Context) {
		rec := NewResponseRecorder(c.Writer)
		rec.Write([]byte("hello"))
		rec.Write([]byte(" world"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		e.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkRouterWithParams(b *testing.B) {
	e := New()
	e.GET("/user/:id", func(c *Context) {
		c.String(http.StatusOK, "%s", c.Param("id"))
	})
	e.GET("/user/:id/post/:postId", func(c *Context) {
		c.String(http.StatusOK, "%s%s", c.Param("id"), c.Param("postId"))
	})
	e.GET("/v1/users/:id", func(c *Context) {
		c.String(http.StatusOK, "%s", c.Param("id"))
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/user/123", nil)
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
	}
}

func BenchmarkConcurrentRequests(b *testing.B) {
	e := New()
	e.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "hello")
	})

	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		for pb.Next() {
			e.ServeHTTP(w, req)
			w.Body.Reset()
		}
	})
}

func BenchmarkMiddlewareChaining(b *testing.B) {
	e := New()
	e.Use(func(c *Context) {
		c.Set("start", "value")
	})
	e.Use(func(c *Context) {
		c.Set("middleware", "test")
	})
	e.Use(func(c *Context) {
		c.Set("third", "data")
	})
	e.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "hello")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		e.ServeHTTP(w, req)
		w.Body.Reset()
	}
}
