package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gellen89/mam-update/internal/app"
)

func TestApp_Run_Success(t *testing.T) {
	srv := startHTTPServer(t, getMux())

	baseDir := t.TempDir()

	t.Setenv("MAM_UPDATE_DIR", baseDir)
	t.Setenv("MAM_SEEDBOX_URL", fmt.Sprintf("%s/mam", srv.URL))
	t.Setenv("IP_URL", fmt.Sprintf("%s/ip", srv.URL))

	args := []string{"-mam-id", "1234"}
	err := app.Run(context.Background(), args)
	if err != nil {
		t.Fatalf("failed to run app: %v", err)
	}

	lastupdateFile := fmt.Sprintf("%s/last_update_time", baseDir)
	_, err = os.ReadFile(lastupdateFile)
	if err != nil {
		t.Fatalf("last run time file does not exist: %v", err)
	}

	ipAddrFile := fmt.Sprintf("%s/MAM.ip", baseDir)
	ipAddrBits, err := os.ReadFile(ipAddrFile)
	if err != nil {
		t.Fatalf("MAM.ip file does not exist: %v", err)
	}

	if testIpAddr != string(ipAddrBits) {
		t.Fatal("ip address was not correct when written to file")
	}

	cookieFile := fmt.Sprintf("%s/MAM.cookie", baseDir)
	_, err = os.ReadFile(cookieFile)
	if err != nil {
		t.Fatalf("MAM.cookie file does not exist: %v", err)
	}
}

type dynamicSeedboxResponse struct {
	Success bool   `json:"Success"`
	Msg     string `json:"msg"`
}

const (
	testIpAddr = "192.168.1.100"
)

func getMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, testIpAddr)
	})
	mux.HandleFunc("/mam", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := dynamicSeedboxResponse{Success: true, Msg: "ok"}
		bits, _ := json.Marshal(&resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(bits)
	})
	return mux
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewServer(h)
	tb.Cleanup(srv.Close)
	return srv
}
