package rawproxy

import (
	"errors"
	"github.com/mholt/caddy/caddy/setup"
	"github.com/mholt/caddy/middleware"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
)

type RawProxy struct {
	Next   middleware.Handler
	path   string
	to     string
	except []string
}

func (c *RawProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) (int,
	error) {
	requestPath := middleware.Path(r.URL.Path)
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

func Setup(c *setup.Controller) (middleware.Middleware, error) {
	path, to, except, err := parseRules(c)
	if err != nil {
		return nil, err
	}

	return func(next middleware.Handler) middleware.Handler {
		return &RawProxy{next, path, to, except}
	}, nil
}

func parseRules(c *setup.Controller) (string, string, []string, error) {
	var except []string
	var path, to string

	for c.Next() {
		if !c.NextArg() {
			if c.NextBlock() {
				switch c.Val() {
				case "except":
					except = append(except, c.RemainingArgs()...)
				}
				break
			}

			return "", "", nil, errors.New("rawproxy: missing `path` parameter")
		}

		path = c.Val()

		if !c.NextArg() {
			return "", "", nil, errors.New("rawproxy: missing `to` parameter")
		}

		to = c.Val()
	}

	return path, to, except, nil
}
