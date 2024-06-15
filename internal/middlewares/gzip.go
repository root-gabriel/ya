package middlewares

import (
	"compress/gzip"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipUnpacking() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			req := ctx.Request()
			rw := ctx.Response().Writer
			header := req.Header

			zap.S().Infof("Request Headers before gzip processing: %v", header)

			if strings.Contains(header.Get("Accept-Encoding"), "gzip") {
				cw := newCompressWriter(rw)
				ctx.Response().Writer = cw
				defer cw.Close()
			}

			if strings.Contains(header.Get("Content-Encoding"), "gzip") {
				cr, err := newCompressReader(req.Body)
				if err != nil {
					return ctx.String(http.StatusInternalServerError, "")
				}
				ctx.Request().Body = cr
				defer cr.Close()
			}

			if err = next(ctx); err != nil {
				ctx.Error(err)
			}

			return err
		}
	}
}

