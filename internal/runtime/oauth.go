package runtime

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type OAuthCallbackResult struct {
	Code  string
	State string
	Err   string
}

func StartOAuthCallbackServer(ctx context.Context, addr string, expectedState string) (string, <-chan OAuthCallbackResult, func() error, error) {
	if addr == "" {
		addr = "127.0.0.1:0"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", nil, nil, err
	}

	resultCh := make(chan OAuthCallbackResult, 1)
	var once sync.Once
	sendResult := func(r OAuthCallbackResult) {
		once.Do(func() {
			resultCh <- r
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		state := q.Get("state")
		if expectedState != "" && state != expectedState {
			sendResult(OAuthCallbackResult{State: state, Err: "invalid state"})
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}
		if errText := q.Get("error"); errText != "" {
			sendResult(OAuthCallbackResult{State: state, Err: errText})
			http.Error(w, errText, http.StatusBadRequest)
			return
		}
		code := q.Get("code")
		if code == "" {
			sendResult(OAuthCallbackResult{State: state, Err: "missing code"})
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}
		sendResult(OAuthCallbackResult{Code: code, State: state})
		_, _ = w.Write([]byte("OAuth connection complete. You can close this window."))
	})

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	go func() {
		if serveErr := srv.Serve(ln); serveErr != nil && serveErr != http.ErrServerClosed {
			sendResult(OAuthCallbackResult{Err: serveErr.Error()})
		}
		close(resultCh)
	}()

	callbackURL := url.URL{
		Scheme: "http",
		Host:   ln.Addr().String(),
		Path:   "/oauth/callback",
	}

	stop := func() error {
		return srv.Shutdown(context.Background())
	}

	return callbackURL.String(), resultCh, stop, nil
}

func WaitForOAuthCode(ctx context.Context, resultCh <-chan OAuthCallbackResult) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res, ok := <-resultCh:
		if !ok {
			return "", fmt.Errorf("oauth callback channel closed")
		}
		if res.Err != "" {
			return "", errors.New(res.Err)
		}
		if res.Code == "" {
			return "", fmt.Errorf("missing oauth code")
		}
		return res.Code, nil
	}
}
