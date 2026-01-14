package handlers

import (
	"finance/database"
	"finance/models"
	"finance/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

// helper untuk status budget agar konsisten
func calculateStatus(total, limit float64) string {
	if total >= limit {
		return "Over Budget"
	} else if total >= limit*0.8 {
		return "Near Limit"
	}
	return "Safe"
}

// CreateBudget
func CreateBudget(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var body struct {
		CategoryID  uint    `json:"category_id"`
		LimitAmount float64 `json:"limit_amount"`
		StartDate   string  `json:"start_date"`
		EndDate     string  `json:"end_date"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	// validasi kategori
	if body.CategoryID != 0 {
		var cat models.Category
		if err := database.DB.Where("id = ? AND user_id = ?", body.CategoryID, uid).First(&cat).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid category"})
		}
	}

	// validasi tanggal
	sd, err := time.Parse("2006-01-02", body.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid start_date"})
	}
	ed, err := time.Parse("2006-01-02", body.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid end_date"})
	}
	if ed.Before(sd) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "end_date must be after start_date"})
	}

	b := models.Budget{
		UserID:      uid,
		CategoryID:  body.CategoryID,
		LimitAmount: body.LimitAmount,
		StartDate:   sd,
		EndDate:     ed,
	}
	if err := database.DB.Create(&b).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "create failed", "detail": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(b)
}

// GetBudgets dengan pagination
func GetBudgets(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	var budgets []models.Budget
	if err := database.DB.Where("user_id = ?", uid).Limit(limit).Offset(offset).Find(&budgets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "query failed", "detail": err.Error()})
	}
	return c.JSON(budgets)
}

// UpdateBudget
func UpdateBudget(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")

	var b models.Budget
	if err := database.DB.Where("id = ? AND user_id = ?", id, uid).First(&b).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}

	var body struct {
		LimitAmount *float64 `json:"limit_amount"`
		StartDate   *string  `json:"start_date"`
		EndDate     *string  `json:"end_date"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	if body.LimitAmount != nil {
		b.LimitAmount = *body.LimitAmount
	}
	if body.StartDate != nil {
		if sd, err := time.Parse("2006-01-02", *body.StartDate); err == nil {
			b.StartDate = sd
		}
	}
	if body.EndDate != nil {
		if ed, err := time.Parse("2006-01-02", *body.EndDate); err == nil {
			if ed.Before(b.StartDate) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "end_date must be after start_date"})
			}
			b.EndDate = ed
		}
	}

	if err := database.DB.Save(&b).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "update failed", "detail": err.Error()})
	}
	return c.JSON(b)
}

// DeleteBudget
func DeleteBudget(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")

	tx := database.DB.Where("id = ? AND user_id = ?", id, uid).Delete(&models.Budget{})
	if tx.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "delete failed", "detail": tx.Error.Error()})
	}
	if tx.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// GetBudgetStatus
func GetBudgetStatus(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var results []struct {
		BudgetID     uint    `json:"budget_id"`
		CategoryName string  `json:"category_name"`
		LimitAmount  float64 `json:"limit_amount"`
		TotalExpense float64 `json:"total_expense"`
		Status       string  `json:"status"`
	}

	database.DB.Raw(`
        SELECT b.id AS budget_id,
               c.name AS category_name,
               b.limit_amount,
               COALESCE(SUM(t.amount),0) AS total_expense,
               CASE 
                 WHEN COALESCE(SUM(t.amount),0) >= b.limit_amount THEN 'Over Budget'
                 WHEN COALESCE(SUM(t.amount),0) >= b.limit_amount * 0.8 THEN 'Near Limit'
                 ELSE 'Safe'
               END AS status
        FROM budgets b
        LEFT JOIN categories c ON b.category_id = c.id
        LEFT JOIN transactions t ON t.category_id = b.category_id
           AND t.user_id = b.user_id
           AND t.date BETWEEN b.start_date AND b.end_date
        WHERE b.user_id = ?
        GROUP BY b.id, c.name, b.limit_amount
    `, uid).Scan(&results)

	return c.JSON(results)
}

// GetBudgetDetail
func GetBudgetDetail(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")

	var budget models.Budget
	if err := database.DB.Where("id = ? AND user_id = ?", id, uid).First(&budget).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}

	var transactions []models.Transaction
	database.DB.Where("user_id = ? AND category_id = ? AND date BETWEEN ? AND ?", uid, budget.CategoryID, budget.StartDate, budget.EndDate).Find(&transactions)

	var total float64
	for _, t := range transactions {
		total += t.Amount
	}
	status := calculateStatus(total, budget.LimitAmount)

	var cat models.Category
	database.DB.First(&cat, budget.CategoryID)

	return c.JSON(fiber.Map{
		"budget":        budget,
		"category_name": cat.Name,
		"transactions":  transactions,
		"total_expense": total,
		"status":        status,
	})
}

// GetBudgetSummary - ringkasan semua budget
func GetBudgetSummary(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var result struct {
		TotalLimit   float64 `json:"total_limit"`
		TotalExpense float64 `json:"total_expense"`
	}
	database.DB.Raw(`
        SELECT COALESCE(SUM(b.limit_amount),0) AS total_limit,
               COALESCE(SUM(t.amount),0) AS total_expense
        FROM budgets b
        LEFT JOIN transactions t ON t.category_id = b.category_id
           AND t.user_id = b.user_id
           AND t.date BETWEEN b.start_date AND b.end_date
        WHERE b.user_id = ?
    `, uid).Scan(&result)

	// hitung persentase total penggunaan
	percentUsed := 0.0
	if result.TotalLimit > 0 {
		percentUsed = (result.TotalExpense / result.TotalLimit) * 100
	}

	return c.JSON(fiber.Map{
		"total_limit":   result.TotalLimit,
		"total_expense": result.TotalExpense,
		"percent_used":  percentUsed,
	})
}
