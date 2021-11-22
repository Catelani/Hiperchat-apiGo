package router

import (
	"net/http"

	"github.com/Mateus-pilo/go-whats-opt/hlp"
)

// HandlerNotFound Function
func handlerNotFound(w http.ResponseWriter, r *http.Request) {
	if (*r).Method == "OPTIONS" {
		return
	}
	
	hlp.LogPrintln(hlp.LogLevelWarn, "http-access", "not found method "+r.Method+" at URI "+r.RequestURI)
	ResponseNotFound(w, "not found method "+r.Method+" at URI "+r.RequestURI)
}

// HandlerMethodNotAllowed Function
func handlerMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	if (*r).Method == "OPTIONS" {
		return
	}
	hlp.LogPrintln(hlp.LogLevelWarn, "http-access", "not allowed method "+r.Method+" at URI "+r.RequestURI)
	ResponseMethodNotAllowed(w, "not allowed method "+r.Method+" at URI "+r.RequestURI)
}

// HandlerFavIcon Function
func handlerFavIcon(w http.ResponseWriter, r *http.Request) {
	if (*r).Method == "OPTIONS" {
		return
	}
	ResponseNoContent(w)
}
