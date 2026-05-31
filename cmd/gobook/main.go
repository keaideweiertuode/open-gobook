package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"gobook/internal/api"
	"gobook/internal/db"
	"gobook/internal/models"
)

func main() {
	dbFile := "data/my_money.db"
	if err := db.InitDB(dbFile); err != nil {
		slog.Error("初始化数据库失败", "error", err)
		os.Exit(1)
	}
	defer db.CloseDB()

	args := os.Args[1:]
	if len(args) > 0 {
		// 【新增】：如果在终端输入的是 go run . rm 12 或 go run . del 12，则执行删除
		if args[0] == "rm" || args[0] == "del" {
			if len(args) < 2 {
				fmt.Println("❌ 请指定要删除的账单 ID。示例: go run . rm 15")
				return
			}
			id, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Println("❌ 无效的账单 ID 格式，必须为数字。")
				return
			}
			if err := db.DeleteTransaction(id); err != nil {
				fmt.Printf("❌ 从数据库删除失败: %v\n", err)
				return
			}
			fmt.Printf("✅ 物理删除成功！账单 ID [%d] 已彻底从本地抹除。\n", id)
			return
		}

		// 否则，正常进行大白话录入逻辑
		rawInput := strings.Join(args, " ")
		handleCLIRecord(rawInput)
		return
	}

	api.StartWebServer()
}

func handleCLIRecord(input string) {
	fmt.Printf("正在通过 AI 解析: \"%s\"...\n", input)
	category, amount, remark, err := mockCallAI(input)
	if err != nil {
		fmt.Printf("❌ AI 解析失败: %v\n", err)
		return
	}
	catID, catType, _ := db.GetCategoryDetailsByName(category)
	tx := &models.Transaction{
		Amount:          amount,
		Type:            catType,
		AccountID:       1,
		CategoryID:      catID,
		TransactionDate: time.Now(),
		RawText:         input,
		Remark:          remark,
	}
	if err := db.SaveTransaction(tx); err != nil {
		fmt.Printf("❌ 写入数据库失败: %v\n", err)
		return
	}
	typeStr := "支出"
	if catType == "income" { typeStr = "收入" }
	fmt.Printf("✅ %s录入成功！[%s] ¥%.2f (备注: %s) -- %s\n", typeStr, category, amount, remark, tx.TransactionDate.Format("2006-01-02 15:04"))
}

func mockCallAI(input string) (category string, amount float64, remark string, err error) {
	parts := strings.Fields(input)
	if len(parts) >= 2 {
		remark = parts[0]
		amount, err = strconv.ParseFloat(parts[1], 64)
		if err == nil {
			if strings.Contains(remark, "饭") || strings.Contains(remark, "面") || strings.Contains(remark, "餐") {
				category = "三餐"
			} else if strings.Contains(remark, "工资") {
				category = "工资"
			} else {
				category = "其它"
			}
			return category, amount, remark, nil
		}
	}
	return "其它", 0, input, fmt.Errorf("暂时无法识别的输入格式")
}