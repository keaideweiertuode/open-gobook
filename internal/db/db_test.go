package db

import (
	"gobook/internal/models"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) {
	// 使用 SQLite 纯内存模式，速度极快且不产生垃圾文件
	err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init in-memory DB: %v", err)
	}
}

func teardownTestDB() {
	CloseDB()
}

func TestGetCategoryDetailsByName(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB()

	// 1. 测试正确的内置分类
	id, catType, err := GetCategoryDetailsByName("三餐")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if catType != "expense" || id == 0 {
		t.Errorf("Failed to get correct category, got id=%d, type=%s", id, catType)
	}

	// 2. 测试分类兜底逻辑（找不到时应该返回 "其它" 的信息）
	id, catType, err = GetCategoryDetailsByName("不存在的分类名称")
	if err != nil {
		t.Errorf("Unexpected error for fallback: %v", err)
	}
	if catType != "expense" || id == 0 {
		t.Errorf("Fallback category should be 'expense' and valid id, got id=%d, type=%s", id, catType)
	}
}

func TestTransactionCRUD(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB()

	catID, catType, _ := GetCategoryDetailsByName("三餐")
	
	// 1. 创建账单 (Create)
	tx := &models.Transaction{
		Amount:          15.5,
		Type:            catType,
		AccountID:       1,
		CategoryID:      catID,
		TransactionDate: time.Now(),
		RawText:         "吃饭15.5",
		Remark:          "午饭",
	}

	err := SaveTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to save transaction: %v", err)
	}

	// 2. 查询账单 (Read)
	today := time.Now().Format("2006-01-02")
	list, err := QueryTransactions(today)
	if err != nil {
		t.Fatalf("Failed to query transactions: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(list))
	}

	if list[0].Amount != 15.5 {
		t.Errorf("Expected amount 15.5, got %v", list[0].Amount)
	}

	// 3. 物理删除 (Delete)
	err = DeleteTransaction(list[0].ID)
	if err != nil {
		t.Fatalf("Failed to delete transaction: %v", err)
	}

	// 4. 测试删除不存在的账单（应返回错误）
	err = DeleteTransaction(999999)
	if err == nil {
		t.Error("Expected error when deleting non-existent ID, got nil")
	}
}
