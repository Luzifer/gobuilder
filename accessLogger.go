package main

import "net/http"

type accessLogResponseWriter struct {
	StatusCode int
	Size       int

	http.ResponseWriter
}

func newAccessLogResponseWriter(res http.ResponseWriter) *accessLogResponseWriter {
	return &accessLogResponseWriter{
		StatusCode:     200,
		Size:           0,
		ResponseWriter: res,
	}
}

func (a *accessLogResponseWriter) Write(out []byte) (int, error) {
	s, err := a.ResponseWriter.Write(out)
	a.Size += s
	return s, err
}

func (a *accessLogResponseWriter) WriteHeader(code int) {
	a.StatusCode = code
	a.ResponseWriter.WriteHeader(code)
}
