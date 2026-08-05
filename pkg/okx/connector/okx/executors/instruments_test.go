package executors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/okx/connector/okx/models"
)

func TestFetchInstrumentsReturnsFetchErrorWithoutPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != instrumentsPath {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("instId") {
		case "BAD-SWAP":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"500","msg":"boom","data":[]}`))
		case "BTC-USDT-SWAP":
			_, _ = w.Write([]byte(`{"code":"0","data":[{"instId":"BTC-USDT-SWAP","instType":"SWAP","ctVal":"0.01","ctMult":"1"}]}`))
		default:
			t.Fatalf("unexpected instId: %s", r.URL.Query().Get("instId"))
		}
	}))
	defer server.Close()

	instruments, err := FetchInstruments(resty.New(), server.URL, map[string]models.Instrumentidentifier{
		"BAD-SWAP":      {InstID: "BAD-SWAP", InstType: "SWAP"},
		"BTC-USDT-SWAP": {InstID: "BTC-USDT-SWAP", InstType: "SWAP"},
	})
	if err == nil {
		t.Fatal("expected fetch error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected 500 error, got %v", err)
	}

	if _, ok := instruments["BAD-SWAP"]; ok {
		t.Fatal("did not expect failed instrument to be stored")
	}
	got, ok := instruments["BTC-USDT-SWAP"]
	if !ok {
		t.Fatal("expected successful instrument to be stored")
	}
	if got.CtVal != "0.01" {
		t.Fatalf("expected ctVal 0.01, got %q", got.CtVal)
	}
}
