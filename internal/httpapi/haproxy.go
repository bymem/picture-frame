package httpapi

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// haProxyPort is the local port the HA reverse proxy listens on.
// The kiosk iframe loads http://localhost:haProxyPort/... so that HA's
// resources are same-origin within the iframe context, bypassing the
// Content-Security-Policy frame-ancestors and X-Frame-Options headers that
// HA sends and that WebKit enforces on cross-origin iframes.
const haProxyPort = 8125

// startHAProxy starts a reverse proxy to Home Assistant on localhost:haProxyPort.
// Only started when dashboard_url is configured; safe to call unconditionally.
// The proxy strips the security headers that block iframe embedding and rewrites
// Location redirects so they stay within the proxy origin.
func (s *server) startHAProxy() {
	dashURL := s.running.Display.DashboardURL
	if dashURL == "" {
		return
	}

	target, err := url.Parse(dashURL)
	if err != nil || target.Host == "" {
		s.log.Error("ha proxy: cannot parse dashboard_url", "url", dashURL, "err", err)
		return
	}
	base := &url.URL{Scheme: target.Scheme, Host: target.Host}

	proxy := httputil.NewSingleHostReverseProxy(base)
	// Fix request headers so HA accepts the proxied requests:
	// 1. Clear Host so Go uses req.URL.Host (192.168.10.20:8123) — HA rejects
	//    requests whose Host header doesn't match its own address.
	// 2. Rewrite Origin/Referer from localhost:8125 to HA's host — HA's CSRF
	//    middleware returns 400 when Origin doesn't match the server host.
	origDirector := proxy.Director
	proxyOrigin := base.String() // e.g. "http://192.168.10.20:8123"
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		req.Host = ""
		if req.Header.Get("Origin") != "" {
			req.Header.Set("Origin", proxyOrigin)
		}
		if ref := req.Header.Get("Referer"); ref != "" {
			req.Header.Set("Referer", strings.Replace(ref,
				fmt.Sprintf("localhost:%d", haProxyPort), base.Host, 1))
		}
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Remove headers that block iframe embedding.
		resp.Header.Del("X-Frame-Options")
		if csp := resp.Header.Get("Content-Security-Policy"); csp != "" {
			if filtered := removeCSPDirective(csp, "frame-ancestors"); filtered != "" {
				resp.Header.Set("Content-Security-Policy", filtered)
			} else {
				resp.Header.Del("Content-Security-Policy")
			}
		}
		// Rewrite Location so HA's own redirects (e.g. / → /lovelace/0) stay
		// within the proxy rather than sending the browser to homeassistant.local.
		if loc := resp.Header.Get("Location"); loc != "" {
			rewritten := strings.Replace(loc, target.Host, fmt.Sprintf("localhost:%d", haProxyPort), 1)
			resp.Header.Set("Location", rewritten)
		}
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		s.log.Warn("ha proxy: upstream error", "path", r.URL.Path, "err", err)
		http.Error(w, "Home Assistant unreachable", http.StatusBadGateway)
	}

	addr := fmt.Sprintf(":%d", haProxyPort)
	srv := &http.Server{Addr: addr, Handler: proxy}
	go func() {
		s.log.Info("ha proxy: listening", "addr", addr, "target", base)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Warn("ha proxy: stopped", "err", err)
		}
	}()
}

// dashboardProxyURL returns the local proxy URL for the kiosk iframe, derived
// from the configured dashboard_url. Returns "" when dashboard_url is not set.
func dashboardProxyURL(dashURL string) string {
	if dashURL == "" {
		return ""
	}
	u, err := url.Parse(dashURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return fmt.Sprintf("http://localhost:%d%s", haProxyPort, u.RequestURI())
}

// removeCSPDirective strips one directive (e.g. "frame-ancestors") from a
// Content-Security-Policy header value, returning the rest joined by ";".
func removeCSPDirective(policy, directive string) string {
	parts := strings.Split(policy, ";")
	out := parts[:0]
	for _, p := range parts {
		if !strings.Contains(strings.ToLower(strings.TrimSpace(p)), directive) {
			out = append(out, p)
		}
	}
	return strings.TrimSpace(strings.Join(out, ";"))
}
