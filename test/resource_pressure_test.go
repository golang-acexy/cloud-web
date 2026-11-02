package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

type auditRequest struct {
	method, path, body string
	status             int
}

// TestHTTPResourceAudit exercises real local HTTP connections without a database.
func TestHTTPResourceAudit(t *testing.T) {
	if os.Getenv("CLOUD_WEB_AUDIT") != "1" {
		t.Skip("set CLOUD_WEB_AUDIT=1")
	}
	log := logger.Logrus()
	previous := log.Out
	log.SetOutput(io.Discard)
	defer log.SetOutput(previous)
	server := httptest.NewServer(ginstarter.RawGinEngine())
	defer server.Close()
	transport := &http.Transport{MaxIdleConns: 16, MaxIdleConnsPerHost: 16}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	cases := []auditRequest{
		{"POST", "/usr/user/save", `{"name":"save","userId":999}`, 200},
		{"GET", "/usr/user/by-id/1", "", 200},
		{"POST", "/adm/user/query-one", `{"name":"one"}`, 200},
		{"POST", "/usr/user/query", `{"name":"list","userId":999}`, 200},
		{"POST", "/usr/user/query-by-page", `{"number":1,"size":16,"condition":{"name":"page"},"timeRanges":[{"field":"createdAt","start":1000,"end":2000}]}`, 200},
		{"PUT", "/usr/user/by-id/1", `{"name":"update","userId":999}`, 200},
		{"DELETE", "/usr/user/by-id/1", "", 200},
		{"GET", "/usr/user/by-id/invalid", "", 400},
		{"POST", "/usr/user/query", `{"name":"failure"}`, 500},
		{"POST", "/usr/user/query", `{"unexpected":1}`, 400},
		{"POST", "/usr/user/query", `{`, 400},
	}
	request := func(c auditRequest) error {
		req, err := http.NewRequest(c.method, server.URL+c.path, strings.NewReader(c.body))
		if err != nil {
			return err
		}
		if c.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var data restResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		}
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			return err
		}
		if resp.StatusCode != 200 || data.Status.StatusCode != c.status {
			return fmt.Errorf("%s HTTP=%d REST=%d want=%d", c.path, resp.StatusCode, data.Status.StatusCode, c.status)
		}
		return nil
	}
	batch := func(inputs []auditRequest, iterations int) {
		var wg sync.WaitGroup
		failures := make(chan error, 8)
		for worker := 0; worker < 8; worker++ {
			wg.Add(1)
			go func(offset int) {
				defer wg.Done()
				for i := 0; i < iterations; i++ {
					if err := request(inputs[(i+offset)%len(inputs)]); err != nil {
						failures <- err
						return
					}
				}
			}(worker)
		}
		wg.Wait()
		close(failures)
		for err := range failures {
			t.Fatal(err)
		}
	}
	sample := func(label string) runtime.MemStats {
		client.CloseIdleConnections()
		server.CloseClientConnections()
		userBizService.reset()
		// Allow closed network goroutines to finish before observing the quiet process.
		time.Sleep(50 * time.Millisecond)
		runtime.GC()
		runtime.GC()
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		t.Logf("%s heap=%d objects=%d goroutines=%d", label, s.HeapAlloc, s.HeapObjects, runtime.NumGoroutine())
		return s
	}
	batch(cases, 50)
	sample("warm")
	for round := 0; round < 6; round++ {
		before := sample(fmt.Sprintf("round%d_before", round))
		start := time.Now()
		batch(cases, 500)
		elapsed := time.Since(start)
		after := sample(fmt.Sprintf("round%d_after", round))
		t.Logf("round=%d requests=4000 elapsed=%s bytes/request=%d", round, elapsed, (after.TotalAlloc-before.TotalAlloc)/4000)
	}
	for round := 0; round < 4; round++ {
		// The fixture and the service's last-request copies are released between bursts.
		func() {
			body := `{"name":"` + strings.Repeat("x", 1<<20) + `"}`
			batch([]auditRequest{{"POST", "/usr/user/query", body, 200}}, 10)
		}()
		sample(fmt.Sprintf("large%d_after", round))
	}
}

// BenchmarkHTTPQuery measures the local engine pipeline without socket overhead.
func BenchmarkHTTPQuery(b *testing.B) {
	body := `{"name":"list","userId":999}`
	engine := ginstarter.RawGinEngine()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/usr/user/query", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		if rec.Code != 200 {
			b.Fatal(rec.Code)
		}
	}
}
