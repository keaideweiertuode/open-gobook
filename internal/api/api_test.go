package api

import (
	"bytes"
	"encoding/json"
	"gobook/internal/db"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestDB(t *testing.T) {
	err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init in-memory DB: %v", err)
	}
}

func teardownTestDB() {
	db.CloseDB()
}

func TestHandleCreateTransaction(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB()

	// 测试正常的创建请求
	reqBody := CreateTxRequest{
		Amount:   20.5,
		Category: "三餐",
		Remark:   "Test dinner",
		RawText:  "晚饭20.5",
	}
	bodyData, _ := json.Marshal(reqBody)
	
	req, err := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(bodyData))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCreateTransaction)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["status"] != "success" {
		t.Errorf("Expected success status, got %v", resp["status"])
	}

	// 测试非法的金额请求（Amount <= 0）
	req, _ = http.NewRequest("POST", "/api/transactions", bytes.NewBuffer([]byte(`{"amount": -5, "category": "三餐"}`)))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for invalid req: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestHandleExport(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB()

	req, _ := http.NewRequest("GET", "/api/export", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExport)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected HTTP 200 OK, got %v", status)
	}
	
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/csv; charset=utf-8" {
		t.Errorf("Expected CSV content type, got %v", contentType)
	}

	// 验证 BOM 存在
	bodyBytes := rr.Body.Bytes()
	if len(bodyBytes) < 3 || bodyBytes[0] != 0xEF || bodyBytes[1] != 0xBB || bodyBytes[2] != 0xBF {
		t.Errorf("Expected UTF-8 BOM in exported CSV")
	}
}
