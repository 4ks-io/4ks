package fetcher

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	requestTimeout        = 20 * time.Second
	responseHeaderTimeout = 10 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	dialTimeout           = 5 * time.Second
	userAgent             = "4ks-fetcher/1.0 (+https://4ks.io)"
)

func initCollector(ctx context.Context, validated *validatedURL, _ bool) (*colly.Collector, error) {
	c := colly.NewCollector(colly.AllowedDomains(validated.Hostname))
	c.UserAgent = userAgent
	c.SetRequestTimeout(requestTimeout)

	if err := c.Limit(&colly.LimitRule{
		Parallelism: 1,
		Delay:       500 * time.Millisecond,
	}); err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: requestTimeout,
		Transport: &http.Transport{
			Proxy:                 nil,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   tlsHandshakeTimeout,
			ResponseHeaderTimeout: responseHeaderTimeout,
			ExpectContinueTimeout: time.Second,
			DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				if isIPLiteral(host) {
					return nil, errIPLiteralNotAllowed
				}

				resolveCtx, cancel := context.WithTimeout(ctx, dialTimeout)
				defer cancel()

				ipAddrs, err := net.DefaultResolver.LookupIPAddr(resolveCtx, host)
				if err != nil {
					return nil, err
				}
				if err := validateResolvedIPs(ipAddrs); err != nil {
					return nil, err
				}

				dialer := &net.Dialer{Timeout: dialTimeout}
				var lastErr error
				for _, ipAddr := range ipAddrs {
					conn, dialErr := dialer.DialContext(resolveCtx, network, net.JoinHostPort(ipAddr.IP.String(), port))
					if dialErr == nil {
						return conn, nil
					}
					lastErr = dialErr
				}
				if lastErr == nil {
					lastErr = errHostResolutionFailed
				}
				return nil, lastErr
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errUnexpectedRedirectCount
			}
			_, err := validateFetchURL(req.Context(), req.URL.String())
			return err
		},
	}

	c.SetClient(client)
	c.SetRedirectHandler(client.CheckRedirect)

	return c, nil
}
