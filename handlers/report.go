// handlers/report.go
package handlers

import (
	"finance/database"
	"finance/utils"

	"github.com/gofiber/fiber/v2"
)

// FR-14..FR-20
func GetSummary(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var totalIncome, totalExpense float64

	database.DB.Raw(`
        SELECT COALESCE(SUM(t.amount),0)
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = ? AND c.type = 'income'
    `, uid).Scan(&totalIncome)

	database.DB.Raw(`
        SELECT COALESCE(SUM(t.amount),0)
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = ? AND c.type = 'expense'
    `, uid).Scan(&totalExpense)

	return c.JSON(fiber.Map{
		"total_income":  totalIncome,
		"total_expense": totalExpense,
		"saldo":         totalIncome - totalExpense,
	})
}

func GetMonthlySummary(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var results []struct {
		Month        string  `json:"month"`
		TotalIncome  float64 `json:"total_income"`
		TotalExpense float64 `json:"total_expense"`
	}

	// ✅ gunakan DATE_FORMAT untuk MySQL
	database.DB.Raw(`
        SELECT DATE_FORMAT(t.date, '%Y-%m') AS month,
               SUM(CASE WHEN c.type='income' THEN t.amount ELSE 0 END) AS total_income,
               SUM(CASE WHEN c.type='expense' THEN t.amount ELSE 0 END) AS total_expense
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = ?
        GROUP BY DATE_FORMAT(t.date, '%Y-%m')
        ORDER BY month
    `, uid).Scan(&results)

	return c.JSON(results)
}

func GetExpenseByCategory(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var results []struct {
		CategoryName string  `json:"category_name"`
		TotalExpense float64 `json:"total_expense"`
	}

	database.DB.Raw(`
        SELECT c.name AS category_name, COALESCE(SUM(t.amount),0) AS total_expense
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = ? AND c.type = 'expense'
        GROUP BY c.name
    `, uid).Scan(&results)

	return c.JSON(results)
}
