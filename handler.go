package rawproxy

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// RawProxy represents the rawproxy plugin for Caddy.
type RawProxy struct {
	Next   httpserver.Handler
	path   string
	to     string
	except []string
}

func (c RawProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) (int,
	error) {
	requestPath := httpserver.Path(r.URL.Path)
	if !requestPath.Matches(c.path) {
		return c.Next.ServeHTTP(w, r)
	}

	for _, exception := range c.except {
		if requestPath.Matches(exception) {
			return c.Next.ServeHTTP(w, r)
		}
	}

	hijacker, _ := w.(http.Hijacker)
	conn, readWriter, err := hijacker.Hijack()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	defer conn.Close()

	remote, err := net.Dial("tcp", c.to)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	defer remote.Close()

	dump, err := httputil.DumpRequest(r, false)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	remote.Write(dump)

	go func() {
		defer conn.Close()
		defer remote.Close()
		io.Copy(remote, readWriter)
	}()

	io.Copy(conn, remote)

	return 0, nil
}

func init() {
	caddy.RegisterPlugin("rawproxy", caddy.Plugin{
		ServerType: "http",
		Action:     setupPlugin,
	})
}

func setupPlugin(c *caddy.Controller) error {
	path, to, except, err := parseRules(c)
	if err != nil {
		return err
	}

	httpserver.GetConfig(c.Key).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return RawProxy{next, path, to, except}
	})

	return nil
}

func parseRules(c *caddy.Controller) (string, string, []string, error) {
	var except []string
	var path, to string

	for c.Next() {
		if !c.NextArg() {
			return "", "", nil, errors.New("rawproxy: missing `path` parameter")
		}

		path = c.Val()

		if !c.NextArg() {
			return "", "", nil, errors.New("rawproxy: missing `to` parameter")
		}

		to = c.Val()

		for c.NextBlock() {
			switch c.Val() {
			case "except":
				except = append(except, c.RemainingArgs()...)
			}
		}
	}

	return path, to, except, nil
}
