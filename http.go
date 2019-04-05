package sectiontrace

import "net/http"

func WrapHandler(section Section, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx, sec := section.Begin(req.Context())
		defer sec.End(nil)
		req = req.WithContext(ctx)

		next.ServeHTTP(w, req)
	})
}
