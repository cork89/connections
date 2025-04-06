package main

import (
	"compress/gzip"
	"context"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	brotli "github.com/andybalholm/brotli"
	uuid "github.com/google/uuid"
)

const CookieName = "Session"

type HttpContext string

const (
	SessionCtx HttpContext = "SessionCtx"
)

type Middleware func(http.Handler) http.Handler

func CreateStack(mw ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			next = mw[i](next)
		}
		return next
	}
}

type scWriter struct {
	http.ResponseWriter
	statusCode int
}

func (mw *scWriter) WriteHeader(statusCode int) {
	mw.ResponseWriter.WriteHeader(statusCode)
	mw.statusCode = statusCode
}

type CompressionWriter interface {
	Write([]byte) (int, error)
}

type compressWriter struct {
	http.ResponseWriter
	compressionWriter CompressionWriter
}

func (cw *compressWriter) Write(bytes []byte) (int, error) {
	return cw.compressionWriter.Write(bytes)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		scWriter := &scWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(scWriter, r)
		log.Println(scWriter.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}

func Session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil {
			temp, err := uuid.NewV7()

			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				next.ServeHTTP(w, r)
				return
			}

			cookie = &http.Cookie{
				Name:     CookieName,
				Value:    temp.String(),
				Path:     "/",
				MaxAge:   int(time.Duration(2160 * time.Hour).Seconds()),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}

			http.SetCookie(w, cookie)
		} else if cookie.MaxAge < int(time.Duration(168*time.Hour).Seconds()) {
			cookie.MaxAge = int(time.Duration(2160 * time.Hour).Seconds())
			http.SetCookie(w, cookie)
		}

		ctx := context.WithValue(r.Context(), SessionCtx, cookie.Value)
		r = r.Clone(ctx)

		next.ServeHTTP(w, r)
	})
}

func CacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2400")
		next.ServeHTTP(w, r)
	})
}

var noCompressionFiles = []string{"webp", "jpeg", "woff2", "mpeg", "mp4", "webm", "common.js", "mygames.js"}

func StaticCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get("Accept-Encoding")
		if slices.Contains(noCompressionFiles, r.RequestURI) {
			next.ServeHTTP(w, r)
			return
		}
		var writer http.ResponseWriter
		if strings.Contains(acceptEncoding, "br") {
			w.Header().Set("Content-Encoding", "br")
			brWriter := brotli.NewWriterV2(w, 4)
			defer brWriter.Close()
			writer = &compressWriter{ResponseWriter: w, compressionWriter: brWriter}
		} else if strings.Contains(acceptEncoding, "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()
			writer = &compressWriter{ResponseWriter: w, compressionWriter: gzipWriter}
		}

		next.ServeHTTP(writer, r)
	})
}
