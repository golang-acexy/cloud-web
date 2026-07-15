package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-acexy/starter-gin/ginstarter"
)

type restResponse struct {
	Status struct {
		StatusCode int `json:"statusCode"`
	} `json:"status"`
	Data json.RawMessage `json:"data"`
}

func performRequest(t *testing.T, method, path, body string) restResponse {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	ginstarter.RawGinEngine().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("HTTP 状态码不正确: path=%s status=%d body=%s", path, recorder.Code, recorder.Body.String())
	}
	var response restResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v body=%s", err, recorder.Body.String())
	}
	return response
}

func assertRestStatus(t *testing.T, response restResponse, expected int) {
	t.Helper()
	if response.Status.StatusCode != expected {
		t.Fatalf("REST 状态码不正确: actual=%d expected=%d data=%s", response.Status.StatusCode, expected, response.Data)
	}
}

func TestAuthorityFieldsOverrideClientValues(t *testing.T) {
	userBizService.reset()
	response := performRequest(t, http.MethodPost, "/usr/user/save", `{"userId":999,"name":"save"}`)
	assertRestStatus(t, response, int(ginstarter.StatusCodeSuccess))
	if userBizService.lastSave.UserID != 12345 {
		t.Fatalf("保存权限字段未覆盖客户端值: %+v", userBizService.lastSave)
	}

	response = performRequest(t, http.MethodPost, "/usr/user/query", `{"userId":999,"name":"query"}`)
	assertRestStatus(t, response, int(ginstarter.StatusCodeSuccess))
	if userBizService.lastCondition["user_id"] != uint64(12345) {
		t.Fatalf("查询权限条件未覆盖客户端值: %+v", userBizService.lastCondition)
	}

	response = performRequest(t, http.MethodPut, "/usr/user/by-id/1", `{"userId":999,"name":"modify"}`)
	assertRestStatus(t, response, int(ginstarter.StatusCodeSuccess))
	if userBizService.lastCondition["user_id"] != uint64(12345) || userBizService.lastUpdate["user_id"] != uint64(12345) {
		t.Fatalf("更新权限字段未被系统强制覆盖: condition=%+v update=%+v", userBizService.lastCondition, userBizService.lastUpdate)
	}
}

func TestPageValidationAndAuthority(t *testing.T) {
	response := performRequest(t, http.MethodPost, "/usr/user/query-by-page", `{"number":0,"size":10}`)
	assertRestStatus(t, response, int(ginstarter.StatusCodeBadRequestParameters))
	response = performRequest(t, http.MethodPost, "/usr/user/query-by-page", `{"number":1,"size":2001}`)
	assertRestStatus(t, response, int(ginstarter.StatusCodeBadRequestParameters))

	response = performRequest(t, http.MethodPost, "/usr/user/query-by-page", `{"number":1,"size":10,"condition":{"userId":999}}`)
	assertRestStatus(t, response, int(ginstarter.StatusCodeSuccess))
	if userBizService.lastCondition["user_id"] != uint64(12345) {
		t.Fatalf("分页权限条件未覆盖客户端值: %+v", userBizService.lastCondition)
	}
	var pager struct {
		Records []UserDTO `json:"records"`
		Total   int64     `json:"total"`
	}
	if err := json.Unmarshal(response.Data, &pager); err != nil || len(pager.Records) != 1 || pager.Total != 1 {
		t.Fatalf("分页响应不正确: pager=%+v err=%v", pager, err)
	}
}

func TestResponseAndErrorSemantics(t *testing.T) {
	response := performRequest(t, http.MethodGet, "/adm/user/by-id/not-number", "")
	assertRestStatus(t, response, int(ginstarter.StatusCodeBadRequestParameters))

	response = performRequest(t, http.MethodGet, "/adm/user/by-id/404", "")
	assertRestStatus(t, response, int(ginstarter.StatusCodeSuccess))
	if len(response.Data) != 0 && string(response.Data) != "null" {
		t.Fatalf("单条不存在应返回成功空数据: %s", response.Data)
	}

	response = performRequest(t, http.MethodPost, "/adm/user/query", `{"name":"empty"}`)
	assertRestStatus(t, response, int(ginstarter.StatusCodeSuccess))
	if string(response.Data) != "[]" {
		t.Fatalf("空列表应返回 []: %s", response.Data)
	}

	response = performRequest(t, http.MethodPut, "/adm/user/by-id/404", `{"name":"modify"}`)
	assertRestStatus(t, response, int(ginstarter.StatusCodeBadRequestParameters))
	response = performRequest(t, http.MethodDelete, "/adm/user/by-id/404", "")
	assertRestStatus(t, response, int(ginstarter.StatusCodeBadRequestParameters))

	response = performRequest(t, http.MethodGet, "/adm/user/by-id/500", "")
	assertRestStatus(t, response, int(ginstarter.StatusCodeException))
	response = performRequest(t, http.MethodPut, "/adm/user/by-id/500", `{"name":"modify"}`)
	assertRestStatus(t, response, int(ginstarter.StatusCodeException))
	response = performRequest(t, http.MethodDelete, "/adm/user/by-id/500", "")
	assertRestStatus(t, response, int(ginstarter.StatusCodeException))
}
