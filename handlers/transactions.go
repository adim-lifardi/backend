// handlers/transaction.go
package handlers

import (
	"fmt"
	"time"

	"finance/database"
	"finance/models"
	"finance/utils"

	"github.com/gofiber/fiber/v2"
)

// FR-09..FR-13, FR-27..FR-28
func CreateTransaction(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	// payload
	var body struct {
		CategoryID uint    `json:"category_id"`
		Amount     float64 `json:"amount"`
		Date       string  `json:"date"`
		Note       string  `json:"note"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if body.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "amount must be greater than 0"})
	}
	if body.Note == "" {
		return c.Status(400).JSON(fiber.Map{"error": "note cannot be empty"})
	}

	// validasi kategori
	var cat models.Category
	if err := database.DB.Where("id = ? AND user_id = ?", body.CategoryID, uid).First(&cat).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid category"})
	}

	// parse tanggal
	parsed, err := time.Parse(time.RFC3339, body.Date)
	if err != nil {
		parsed = time.Now()
	}

	// buat transaksi
	trx := models.Transaction{
		UserID:     uid,
		CategoryID: body.CategoryID,
		Amount:     body.Amount,
		Date:       parsed,
		Note:       body.Note,
	}
	if err := database.DB.Create(&trx).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "create failed", "detail": err.Error()})
	}

	// cek budget terkait
	var budget models.Budget
	if err := database.DB.Where("category_id = ? AND user_id = ?", trx.CategoryID, uid).First(&budget).Error; err == nil {
		var totalExpense float64
		database.DB.Model(&models.Transaction{}).
			Where("category_id = ? AND user_id = ? AND date BETWEEN ? AND ?", trx.CategoryID, uid, budget.StartDate, budget.EndDate).
			Select("COALESCE(SUM(amount),0)").Scan(&totalExpense)

		status := calculateStatus(totalExpense, budget.LimitAmount)
		if status == "Over Budget" {
			// buat notifikasi
			if err := AddBudgetNotification(uid, cat.Name); err != nil {
				fmt.Println("Gagal simpan notifikasi:", err)
			}
		}

		return c.Status(201).JSON(fiber.Map{
			"transaction":   trx,
			"budget_status": status,
			"total_expense": totalExpense,
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"transaction":   trx,
		"budget_status": "Safe",
		"total_expense": trx.Amount,
	})
}

func GetTransactions(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	// filters
	start := c.Query("start_date")
	end := c.Query("end_date")
	categoryID := c.Query("category_id")
	keyword := c.Query("keyword")
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	// query with JOIN
	var results []struct {
		ID           uint    `json:"id"`
		Amount       float64 `json:"amount"`
		Note         string  `json:"note"`
		Date         string  `json:"date"`
		CategoryID   uint    `json:"category_id"`
		CategoryName string  `json:"category_name"`
		CategoryType string  `json:"category_type"`
	}

	query := `
		SELECT t.id, t.amount, t.note, t.date,
		       c.id AS category_id, c.name AS category_name, c.type AS category_type
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = ?
	`

	args := []interface{}{uid}

	if categoryID != "" {
		query += " AND c.id = ?"
		args = append(args, categoryID)
	}
	if start != "" && end != "" {
		query += " AND t.date BETWEEN ? AND ?"
		args = append(args, start, end)
	}
	if keyword != "" {
		query += " AND t.note LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	query += " ORDER BY t.date DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	if err := database.DB.Raw(query, args...).Scan(&results).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query failed", "detail": err.Error()})
	}

	return c.JSON(results)
}

func GetTransaction(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")
	var trx models.Transaction
	if err := database.DB.Where("id = ? AND user_id = ?", id, uid).First(&trx).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(trx)
}

func UpdateTransaction(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")

	var trx models.Transaction
	if err := database.DB.Where("id = ? AND user_id = ?", id, uid).First(&trx).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}

	var body struct {
		CategoryID *uint    `json:"category_id"`
		Amount     *float64 `json:"amount"`
		Date       *string  `json:"date"`
		Note       *string  `json:"note"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if body.CategoryID != nil {
		var cat models.Category
		if err := database.DB.Where("id = ? AND user_id = ?", *body.CategoryID, uid).First(&cat).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid category"})
		}
		trx.CategoryID = *body.CategoryID
	}
	if body.Amount != nil {
		trx.Amount = *body.Amount
	}
	if body.Date != nil {
		if parsed, err := time.Parse(time.RFC3339, *body.Date); err == nil {
			trx.Date = parsed
		}
	}
	if body.Note != nil {
		trx.Note = *body.Note
	}

	if err := database.DB.Save(&trx).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update failed"})
	}
	return c.JSON(trx)
}

func DeleteTransaction(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")
	if err := database.DB.Where("id = ? AND user_id = ?", id, uid).Delete(&models.Transaction{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "delete failed"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}
