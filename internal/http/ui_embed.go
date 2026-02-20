package http

import (
	"embed"
	"io/fs"
	httpstd "net/http"
)

//go:embed all:ui
var uiFiles embed.FS

func (s *Server) uiHandler() httpstd.Handler {
	sub, err := fs.Sub(uiFiles, "ui")
	if err != nil {
		return httpstd.HandlerFunc(func(w httpstd.ResponseWriter, _ *httpstd.Request) {
			writeError(w, httpstd.StatusInternalServerError, "ui_assets_missing", nil)
		})
	}
	return httpstd.StripPrefix("/ui/", httpstd.FileServer(httpstd.FS(sub)))
}

func (s *Server) handleUIRoot(w httpstd.ResponseWriter, r *httpstd.Request) {
	if r.URL.Path != "/" {
		writeError(w, httpstd.StatusNotFound, "not_found", nil)
		return
	}
	httpstd.Redirect(w, r, "/ui/", httpstd.StatusTemporaryRedirect)
}
