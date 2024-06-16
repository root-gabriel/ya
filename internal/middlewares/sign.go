package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/labstack/echo/v4"
	"hash"
	"io"
	"net/http"
)

func CheckSignReq(password string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			req := ctx.Request()
			body, err := io.ReadAll(req.Body)
			if err == nil {
				singPassword := []byte(password)
				bodyHash := GetSign(body, singPassword)
				signR := req.Header.Get("HashSHA256")

				if signR != bodyHash {
					return ctx.String(http.StatusBadRequest, "signature is not valid")
				}
			}
			req.Body = io.NopCloser(bytes.NewReader(body))
			return next(ctx)
		}
	}
}

func GetSign(body []byte, pass []byte) string {
	hashValue := hmac.New(sha256.New, pass)
	hashValue.Write(body)
	sum := hashValue.Sum(nil)
	return hex.EncodeToString(sum)
}

type signResponseWriter struct {
	http.ResponseWriter
	hash hash.Hash
}

func (w signResponseWriter) Write(b []byte) (int, error) {
	w.hash.Write(b)
	w.Header().Set("HashSHA256", hex.EncodeToString(w.hash.Sum(nil)))
	return w.ResponseWriter.Write(b)
}
