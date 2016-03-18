# Raw Proxy
A Caddy extension which directly proxies the TCP stream of connections without any additional processing done by Caddy (except for the parsing of the headers). The proxy only connects to the remote server as HTTP/1.1.

This was a quick extension written because Caddy was interfering with one of my web servers which required direct instant receiving and responses without any buffering.

I've only tested this extension with basic HTTP GET requests. I'm not sure if it will work for POST or any other requests. Use at your own risk. I'm more than happy to fix and improve on this extension to support all kinds of requests. Feel free to let me know about any issues.

## Installation
1. `go get -u github.com/caddyserver/caddyext`
2. `caddyext install --after log rawproxy:github.com/1lann/rawproxy`
3. `caddyext build`
4. `./caddy`

## Usage
In your caddy file, you can specify the following directive:

```
rawproxy from to {
	except ignored_paths...
}
```

**from** is the path to match to proxy from.<br>
**to** is the IP and port of the server to proxy to.<br>
**ignored_paths** is an optional space seperated list of paths to exclude from being proxied.

### Example configurations
```
rawproxy / 127.0.0.1:9001
```
Proxies everything to 127.0.0.1:9001.

```
rawproxy /cat 127.0.0.1:9001 {
	except /cat/dog
}
```
Proxies everything matching `/cat` except for paths matching `/cat/dog` to 127.0.0.1:9001.

## License
Raw Proxy is licensed under The MIT License which can be found [here](/LICENSE).
