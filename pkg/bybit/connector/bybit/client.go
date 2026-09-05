package bybit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	BaseURL = "https://api.bybit.com"

	defaultTimeout    = 20 * time.Second
	defaultRecvWindow = "10000"

	publicPathPrefix = "/v5/market/"

	CodeTimestampOutOfWindow = 10002
	CodeRateLimit            = 10006
	CodeServerError          = 10016
	CodeIPRateLimit          = 10018
	CodeUnifiedOnly          = 10028
)

type Credentials struct {
	APIKey string
	Secret string
}

type authState struct {
	creds  Credentials
	offset atomic.Int64
}

var attached sync.Map

func isPublicPath(path string) bool {
	return strings.HasPrefix(path, publicPathPrefix)
}

func AttachAuth(client *resty.Client, creds Credentials) {
	if prev, ok := attached.Load(client); ok && prev.(*authState).creds == creds {
		return
	}
	state := &authState{creds: creds}
	state.offset.Store(serverTimeOffset())
	attached.Store(client, state)

	client.SetPreRequestHook(func(_ *resty.Client, req *http.Request) error {
		if isPublicPath(req.URL.Path) {
			return nil
		}
		ts := strconv.FormatInt(time.Now().UnixMilli()+state.offset.Load(), 10)
		req.Header.Set("X-BAPI-API-KEY", creds.APIKey)
		req.Header.Set("X-BAPI-TIMESTAMP", ts)
		req.Header.Set("X-BAPI-RECV-WINDOW", defaultRecvWindow)
		req.Header.Set("X-BAPI-SIGN", Sign(creds.Secret, ts, creds.APIKey, defaultRecvWindow, req.URL.RawQuery))
		return nil
	})
}

func NewBaseClient() *resty.Client {
	return resty.New().
		SetTimeout(defaultTimeout).
		SetRetryCount(0)
}

func Sign(secret, timestamp, apiKey, recvWindow, query string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(timestamp + apiKey + recvWindow + query))
	return hex.EncodeToString(h.Sum(nil))
}

type serverTime struct {
	TimeSecond string `json:"timeSecond"`
	TimeNano   string `json:"timeNano"`
}

func serverTimeOffset() int64 {
	before := time.Now().UnixMilli()
	st, err := DoGet[serverTime](NewBaseClient(), "/v5/market/time", nil)
	if err != nil {
		return 0
	}
	after := time.Now().UnixMilli()
	nano, convErr := strconv.ParseInt(st.TimeNano, 10, 64)
	if convErr != nil || nano == 0 {
		sec, secErr := strconv.ParseInt(st.TimeSecond, 10, 64)
		if secErr != nil || sec == 0 {
			return 0
		}
		nano = sec * 1_000_000_000
	}
	return nano/1_000_000 - (before+after)/2
}

func resyncTime(client *resty.Client) {
	if v, ok := attached.Load(client); ok {
		v.(*authState).offset.Store(serverTimeOffset())
	}
}

type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
	RetryAfter time.Duration
}

func (e *HTTPError) Error() string {
	if strings.TrimSpace(e.Body) == "" {
		return fmt.Sprintf("bybit: unexpected status %s", e.Status)
	}
	return fmt.Sprintf("bybit: unexpected status %s: %s", e.Status, strings.TrimSpace(e.Body))
}

func (e *HTTPError) IPBanned() bool {
	return e.StatusCode == http.StatusForbidden
}

type APIError struct {
	Code    int
	Msg     string
	Path    string
	ResetAt time.Time
}

func (e *APIError) Error() string {
	return fmt.Sprintf("bybit: error %d: %s (%s)", e.Code, e.Msg, e.Path)
}

type envelope struct {
	RetCode int             `json:"retCode"`
	RetMsg  string          `json:"retMsg"`
	Result  json.RawMessage `json:"result"`
	Time    int64           `json:"time"`
}

func DoGet[T any](client *resty.Client, path string, params map[string]string) (T, error) {
	result, err := doGet[T](client, path, params)
	var apiErr *APIError
	if err != nil && asAPIError(err, &apiErr) && apiErr.Code == CodeTimestampOutOfWindow {
		resyncTime(client)
		return doGet[T](client, path, params)
	}
	return result, err
}

func asAPIError(err error, target **APIError) bool {
	e, ok := err.(*APIError)
	if ok {
		*target = e
	}
	return ok
}

func doGet[T any](client *resty.Client, path string, params map[string]string) (T, error) {
	var zero T

	Limiter.Acquire(path)

	req := client.R()
	for k, v := range params {
		req.SetQueryParam(k, v)
	}

	resp, err := req.Get(BaseURL + path)
	if err != nil {
		return zero, err
	}

	Limiter.Observe(path, resp.Header())

	body := resp.Body()
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		httpErr := &HTTPError{
			StatusCode: resp.StatusCode(),
			Status:     resp.Status(),
			Body:       string(body),
		}
		if ra := resp.Header().Get("Retry-After"); ra != "" {
			if secs, convErr := strconv.Atoi(ra); convErr == nil {
				httpErr.RetryAfter = time.Duration(secs) * time.Second
			}
		}
		return zero, httpErr
	}

	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return zero, fmt.Errorf("bybit: decode %s: %w", path, err)
	}
	if env.RetCode != 0 {
		apiErr := &APIError{Code: env.RetCode, Msg: env.RetMsg, Path: path}
		if env.RetCode == CodeRateLimit {
			apiErr.ResetAt = Limiter.ResetAt(path)
		}
		return zero, apiErr
	}

	var result T
	if len(env.Result) == 0 || string(env.Result) == "null" {
		return result, nil
	}
	if err := json.Unmarshal(env.Result, &result); err != nil {
		return zero, fmt.Errorf("bybit: decode %s result: %w", path, err)
	}
	return result, nil
}
