package test

import "testing"

// TestFieldValidationHTTP preserves route rejection, null values, and authority injection.
func TestFieldValidationHTTP(t *testing.T) {
	t.Cleanup(userBizService.reset)
	for _, tc := range []struct{ method, path, body string }{
		{"POST", "/adm/user/query", `{}`},
		{"POST", "/adm/user/query-one", `null`},
		{"POST", "/adm/user/query", `{"name":"ok","unknown":1}`},
		{"POST", "/adm/user/query-one", `{"unknown":1}`},
		{"POST", "/adm/user/query-by-page", `{"number":1,"size":10,"condition":{"unknown":1}}`},
		{"PUT", "/adm/user/by-id/1", `{}`},
		{"PUT", "/adm/user/by-id/1", `{"name":"ok","id":1}`},
	} {
		userBizService.reset()
		response := performRequest(t, tc.method, tc.path, tc.body)
		assertRestStatus(t, response, 400)
		if userBizService.lastCondition != nil || userBizService.lastUpdate != nil {
			t.Fatalf("rejected request reached service: %s", tc.path)
		}
	}
	for _, path := range []string{"/usr/user/query", "/adm/user/query-one"} {
		userBizService.reset()
		assertRestStatus(t, performRequest(t, "POST", path, `{"name":null,"userId":0}`), 200)
		value, exists := userBizService.lastCondition["name"]
		if !exists || value != nil {
			t.Fatal("null changed")
		}
		if path == "/usr/user/query" && userBizService.lastCondition["user_id"] != uint64(12345) {
			t.Fatal("authority not applied after conversion")
		}
	}
	userBizService.reset()
	assertRestStatus(t, performRequest(t, "PUT", "/usr/user/by-id/1", `{"name":null,"userId":0}`), 200)
	if value, exists := userBizService.lastUpdate["name"]; !exists || value != nil {
		t.Fatal("null modification changed")
	}
	if userBizService.lastUpdate["user_id"] != uint64(12345) || userBizService.lastCondition["user_id"] != uint64(12345) {
		t.Fatal("authority assignments or predicates changed")
	}
	userBizService.reset()
	assertRestStatus(t, performRequest(t, "POST", "/usr/user/query-by-page", `{"number":1,"size":10}`), 200)
	if userBizService.lastCondition["user_id"] != uint64(12345) {
		t.Fatal("empty page condition lost authority")
	}
}
