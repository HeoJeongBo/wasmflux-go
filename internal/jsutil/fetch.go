//go:build js && wasm

package jsutil

import (
	"encoding/json"
	"fmt"
	"syscall/js"
)

// FetchOption configures a Fetch request.
type FetchOption func(opts js.Value)

// WithMethod sets the HTTP method.
func WithMethod(method string) FetchOption {
	return func(opts js.Value) { opts.Set("method", method) }
}

// WithHeader adds a header to the request.
func WithHeader(key, value string) FetchOption {
	return func(opts js.Value) {
		headers := opts.Get("headers")
		if headers.IsUndefined() {
			headers = js.Global().Get("Headers").New()
			opts.Set("headers", headers)
		}
		headers.Call("set", key, value)
	}
}

// WithBody sets the request body as a string.
func WithBody(body string) FetchOption {
	return func(opts js.Value) { opts.Set("body", body) }
}

// WithJSONBody sets the request body as JSON and adds the Content-Type header.
func WithJSONBody(v any) FetchOption {
	return func(opts js.Value) {
		b, err := json.Marshal(v)
		if err != nil {
			return
		}
		opts.Set("body", string(b))
		headers := opts.Get("headers")
		if headers.IsUndefined() {
			headers = js.Global().Get("Headers").New()
			opts.Set("headers", headers)
		}
		headers.Call("set", "Content-Type", "application/json")
	}
}

// WithCredentials sets the credentials mode ("include", "same-origin", "omit").
func WithCredentials(mode string) FetchOption {
	return func(opts js.Value) { opts.Set("credentials", mode) }
}

// WithSignal sets an AbortController signal for request cancellation.
func WithSignal(signal js.Value) FetchOption {
	return func(opts js.Value) { opts.Set("signal", signal) }
}

// FetchResponse wraps a JS Response object with typed accessors.
type FetchResponse struct {
	raw        js.Value
	Status     int
	StatusText string
	OK         bool
	URL        string
}

// Text reads the response body as a string.
func (r *FetchResponse) Text() (string, error) {
	return awaitString(r.raw.Call("text"))
}

// JSON reads the response body and unmarshals it into dst.
func (r *FetchResponse) JSON(dst any) error {
	text, err := r.Text()
	if err != nil {
		return fmt.Errorf("fetch response text: %w", err)
	}
	return json.Unmarshal([]byte(text), dst)
}

// Bytes reads the response body as []byte.
func (r *FetchResponse) Bytes() ([]byte, error) {
	result, err := awaitJS(r.raw.Call("arrayBuffer"))
	if err != nil {
		return nil, err
	}
	uint8Array := js.Global().Get("Uint8Array").New(result)
	buf := make([]byte, uint8Array.Length())
	js.CopyBytesToGo(buf, uint8Array)
	return buf, nil
}

// Header returns the value of a response header.
func (r *FetchResponse) Header(name string) string {
	return r.raw.Get("headers").Call("get", name).String()
}

// Raw returns the underlying JS Response object.
func (r *FetchResponse) Raw() js.Value {
	return r.raw
}

// Fetch performs an HTTP request using the JS fetch API.
func Fetch(url string, opts ...FetchOption) (*FetchResponse, error) {
	init := js.Global().Get("Object").New()
	for _, opt := range opts {
		opt(init)
	}

	result, err := awaitJS(Global.Call("fetch", url, init))
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}

	return &FetchResponse{
		raw:        result,
		Status:     result.Get("status").Int(),
		StatusText: result.Get("statusText").String(),
		OK:         result.Get("ok").Bool(),
		URL:        result.Get("url").String(),
	}, nil
}

// FetchJSON performs a GET request and unmarshals the JSON response into dst.
func FetchJSON(url string, dst any, opts ...FetchOption) error {
	resp, err := Fetch(url, opts...)
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("fetch %s: %d %s", url, resp.Status, resp.StatusText)
	}
	return resp.JSON(dst)
}

// PostJSON performs a POST request with a JSON body and unmarshals the response into dst.
func PostJSON(url string, body any, dst any, opts ...FetchOption) error {
	allOpts := []FetchOption{WithMethod("POST"), WithJSONBody(body)}
	allOpts = append(allOpts, opts...)
	resp, err := Fetch(url, allOpts...)
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("post %s: %d %s", url, resp.Status, resp.StatusText)
	}
	if dst != nil {
		return resp.JSON(dst)
	}
	return nil
}

func awaitJS(promise js.Value) (js.Value, error) {
	ch := make(chan js.Value, 1)
	errCh := make(chan error, 1)

	onResolve := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			ch <- args[0]
		} else {
			ch <- js.Undefined()
		}
		return nil
	})
	defer onResolve.Release()

	onReject := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			errCh <- fmt.Errorf("%s", args[0].Call("toString").String())
		} else {
			errCh <- fmt.Errorf("promise rejected")
		}
		return nil
	})
	defer onReject.Release()

	promise.Call("then", onResolve).Call("catch", onReject)

	select {
	case v := <-ch:
		return v, nil
	case err := <-errCh:
		return js.Undefined(), err
	}
}

func awaitString(promise js.Value) (string, error) {
	v, err := awaitJS(promise)
	if err != nil {
		return "", err
	}
	return v.String(), nil
}
