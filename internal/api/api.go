package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv" // 【新增】用于解析字符串到整型ID
	"time"
	"gobook/internal/db"
	"gobook/internal/models"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"status": "error", "message": msg})
}

func StartWebServer() {
	http.Handle("/", http.FileServer(http.Dir("./web/static")))

	http.HandleFunc("GET /api/stats/daily", handleDailyStats)
	http.HandleFunc("GET /api/stats/category", handleCategoryStats)
	http.HandleFunc("POST /api/transactions", handleCreateTransaction)
	http.HandleFunc("GET /api/stats/years", handleAvailableYears) 
	
	// 【新增】流水明细与物理删除接口
	http.HandleFunc("GET /api/transactions/list", handleListTransactions)
	http.HandleFunc("POST /api/transactions/delete", handleDeleteTransaction)
	http.HandleFunc("DELETE /api/transactions/delete", handleDeleteTransaction)

	// 【新增】数据导出接口
	http.HandleFunc("GET /api/export", handleExport)

	port := ":8080"
	fmt.Printf("🚀 Web UI & API 服务已启动: http://localhost%s\n", port)
	if err := http.ListenAndServe(port, corsMiddleware(http.DefaultServeMux)); err != nil {
		fmt.Printf("Web 服务启动失败: %v\n", err)
	}
}

// 【新增】处理流水明细列表获取
func handleListTransactions(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = time.Now().Format("2006-01-02")
	}
	
	list, err := db.QueryTransactions(period)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "获取明细列表失败")
		return
	}
	if list == nil {
		list = []models.TransactionDetail{}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// 【新增】处理物理删除路由 (支持 DELETE 和 POST 方法)
func handleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "无效的账单 ID")
		return
	}

	if err := db.DeleteTransaction(id); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success","message":"账单删除成功"}`))
}

type CreateTxRequest struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Remark   string  `json:"remark"`
	RawText  string  `json:"raw_text"`
}

func handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "解析 JSON 失败")
		return
	}
	if req.Amount <= 0 || req.Category == "" {
		jsonError(w, http.StatusBadRequest, "金额或分类不能为空")
		return
	}
	catID, catType, _ := db.GetCategoryDetailsByName(req.Category)
	tx := &models.Transaction{
		Amount:          req.Amount,
		Type:            catType,
		AccountID:       1,
		CategoryID:      catID,
		TransactionDate: time.Now(),
		RawText:         req.RawText,
		Remark:          req.Remark,
	}
	if err := db.SaveTransaction(tx); err != nil {
		jsonError(w, http.StatusInternalServerError, "写入数据库失败")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"status":"success","message":"账单已同步","type":"%s","matched_category_id":%d}`, catType, catID)))
}

func handleDailyStats(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" { period = time.Now().Format("2006-01-02") }
	stats, err := db.QueryTrendStats(period)
	if err != nil { jsonError(w, http.StatusInternalServerError, "查询失败"); return }
	if stats == nil { stats = []models.TrendStat{} }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleCategoryStats(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" { period = time.Now().Format("2006-01-02") }
	stats, err := db.QueryCategoryStats(period)
	if err != nil { jsonError(w, http.StatusInternalServerError, "查询失败"); return }
	if stats == nil { stats = []models.CategoryStat{} }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleAvailableYears(w http.ResponseWriter, r *http.Request) {
	years, err := db.QueryAvailableYears()
	if err != nil { jsonError(w, http.StatusInternalServerError, "查询年份失败"); return }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(years)
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	list, err := db.QueryTransactions(period)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "获取导出数据失败")
		return
	}

	filename := "gobook_export.csv"
	if period != "" {
		filename = fmt.Sprintf("gobook_export_%s.csv", period)
	}

	if format == "json" {
		if period != "" {
			filename = fmt.Sprintf("gobook_export_%s.json", period)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		json.NewEncoder(w).Encode(list)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	
	// 写入 UTF-8 BOM，防止 Excel 打开中文乱码
	w.Write([]byte("\xEF\xBB\xBF"))

	cw := csv.NewWriter(w)
	// 表头
	cw.Write([]string{"ID", "交易日期", "收支类型", "分类", "金额", "备注"})
	
	for _, t := range list {
		typeStr := "支出"
		if t.Type == "income" {
			typeStr = "收入"
		}
		cw.Write([]string{
			fmt.Sprintf("%d", t.ID),
			t.TransactionDate,
			typeStr,
			t.CategoryName,
			fmt.Sprintf("%.2f", t.Amount),
			t.Remark,
		})
	}
	cw.Flush()
}