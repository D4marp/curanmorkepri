package httpx

import "net/http"

// Middleware membungkus http.Handler.
type Middleware func(http.Handler) http.Handler

// Router adalah wrapper tipis di atas http.ServeMux (Go 1.22+, mendukung
// pola "METHOD /path/{param}") dengan dukungan grup middleware bergaya
// chi/gin, tanpa dependensi eksternal.
type Router struct {
	mux    *http.ServeMux
	prefix string
	mws    []Middleware
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

// Group membuat sub-router dengan prefix path & middleware tambahan yang
// hanya berlaku pada grup tersebut (mis. grup /api/v1/wa dengan API-key auth).
func (r *Router) Group(prefix string, mws ...Middleware) *Router {
	combined := make([]Middleware, 0, len(r.mws)+len(mws))
	combined = append(combined, r.mws...)
	combined = append(combined, mws...)
	return &Router{mux: r.mux, prefix: r.prefix + prefix, mws: combined}
}

func (r *Router) handle(method, path string, h http.HandlerFunc) {
	full := method + " " + r.prefix + path
	var handler http.Handler = h
	for i := len(r.mws) - 1; i >= 0; i-- {
		handler = r.mws[i](handler)
	}
	r.mux.Handle(full, handler)
}

func (r *Router) Get(path string, h http.HandlerFunc)    { r.handle(http.MethodGet, path, h) }
func (r *Router) Post(path string, h http.HandlerFunc)   { r.handle(http.MethodPost, path, h) }
func (r *Router) Put(path string, h http.HandlerFunc)    { r.handle(http.MethodPut, path, h) }
func (r *Router) Patch(path string, h http.HandlerFunc)  { r.handle(http.MethodPatch, path, h) }
func (r *Router) Delete(path string, h http.HandlerFunc) { r.handle(http.MethodDelete, path, h) }

// Handle mendaftarkan handler mentah tanpa prefix method (mis. untuk static file serving).
func (r *Router) Handle(pattern string, h http.Handler) {
	var handler http.Handler = h
	for i := len(r.mws) - 1; i >= 0; i-- {
		handler = r.mws[i](handler)
	}
	r.mux.Handle(r.prefix+pattern, handler)
}

func (r *Router) ServeMux() *http.ServeMux { return r.mux }
