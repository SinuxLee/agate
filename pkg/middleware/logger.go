package middleware

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/valyala/bytebufferpool"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytebufferpool.ByteBuffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	_, _ = w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w responseWriter) WriteString(s string) (int, error) {
	_, _ = w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		rw := &responseWriter{body: bytebufferpool.Get(), ResponseWriter: c.Writer}
		defer bytebufferpool.Put(rw.body)

		c.Writer = rw
		body, _ := c.GetRawData()
		_ = c.Request.Body.Close()
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		begin := time.Now()

		c.Next()

		log.Info().Str("remote", c.Request.RemoteAddr).
			// Str("referer", c.Request.Referer()).
			Str("method", c.Request.Method).
			Str("uri", c.Request.URL.String()).
			Interface("header", c.Request.Header).
			Interface("param", c.Params).
			RawJSON("body", body).
			Interface("rspHeader", rw.Header()).
			Int("statusCode", rw.Status()).
			Int("contentLen", rw.Size()).
			RawJSON("response", rw.body.Bytes()).
			TimeDiff("cost", time.Now(), begin).
			Msg("logger")
	}
}
