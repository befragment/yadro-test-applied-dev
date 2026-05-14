package normalizer

import "net/http"

// ServeMuxPathNormalizer возвращает шаблон маршрута stdlib ServeMux (Go 1.22+),
// аналог chi.RoutePattern для net/http.
type ServeMuxPathNormalizer struct{}

func NewServeMuxPathNormalizer() *ServeMuxPathNormalizer {
	return &ServeMuxPathNormalizer{}
}

func (n *ServeMuxPathNormalizer) Normalize(r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}
	return "unknown"
}
