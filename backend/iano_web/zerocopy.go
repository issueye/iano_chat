package web

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ZeroCopyWriter interface {
	http.ResponseWriter
	io.ReaderFrom
}

type writeBuffer struct {
	buf []byte
}

func newWriteBuffer(size int) *writeBuffer {
	return &writeBuffer{
		buf: make([]byte, 0, size),
	}
}

func (wb *writeBuffer) Write(p []byte) (n int, err error) {
	wb.buf = append(wb.buf, p...)
	return len(p), nil
}

func (wb *writeBuffer) Bytes() []byte {
	return wb.buf
}

func (wb *writeBuffer) Reset() {
	wb.buf = wb.buf[:0]
}

func (c *Context) SendFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	c.SetHeader("Content-Length", strconv.FormatInt(stat.Size(), 10))

	bufWriter := bufio.NewWriter(c.Writer)
	_, err = io.Copy(bufWriter, file)
	if err != nil {
		return err
	}
	return bufWriter.Flush()
}

func (c *Context) SendFileRange(filepath string, offset, length int64) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	c.SetHeader("Content-Length", strconv.FormatInt(length, 10))
	c.SetHeader("Accept-Ranges", "bytes")
	c.SetHeader("Content-Range", "bytes "+strconv.FormatInt(offset, 10)+"-"+strconv.FormatInt(offset+length-1, 10)+"/"+strconv.FormatInt(length, 10))
	c.Status(http.StatusPartialContent)

	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = io.CopyN(c.Writer, file, length)
	return err
}

func (c *Context) ServeContent(name string, modTime int64, content io.ReadSeeker) {
	c.SetHeader("Last-Modified", time.Unix(modTime, 0).Format(time.RFC1123))

	if ims := c.GetHeader("If-Modified-Since"); ims != "" {
		if t, err := http.ParseTime(ims); err == nil {
			if t.Unix() >= modTime {
				c.Status(http.StatusNotModified)
				return
			}
		}
	}

	if rs, ok := content.(io.Reader); ok {
		io.Copy(c.Writer, rs)
	}
}

func (c *Context) WriteString(s string) (int, error) {
	return c.Writer.Write([]byte(s))
}

func CopyBuffer(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := acquireBytes()
	defer releaseBytes(buf)

	if cap(buf) < 32*1024 {
		buf = make([]byte, 32*1024)
	}

	return io.CopyBuffer(dst, src, buf)
}

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
	written    bool
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Body:           make([]byte, 0, 1024),
	}
}

func (r *ResponseRecorder) WriteHeader(code int) {
	if !r.written {
		r.StatusCode = code
		r.written = true
		r.ResponseWriter.WriteHeader(code)
	}
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	r.Body = append(r.Body, b...)
	return r.ResponseWriter.Write(b)
}

func (r *ResponseRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *ResponseRecorder) Reset() {
	r.StatusCode = http.StatusOK
	r.Body = r.Body[:0]
	r.written = false
}

func CopyWithPool(dst io.Writer, src io.Reader) (int64, error) {
	buf := acquireBytes()
	defer releaseBytes(buf)
	return io.CopyBuffer(dst, src, buf)
}
