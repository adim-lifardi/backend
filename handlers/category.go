// handlers/category.go
package handlers

import (
	"finance/database"
	"finance/models"
	"finance/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// FR-06..FR-08
func CreateCategory(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var body struct {
		Name string `json:"name"`
		Type string `json:"type"` // income/expense
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	body.Type = strings.ToLower(body.Type)
	if body.Type != "income" && body.Type != "expense" {
		return c.Status(400).JSON(fiber.Map{"error": "type must be income or expense"})
	}
	if strings.TrimSpace(body.Name) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name cannot be empty"})
	}

	// Cek duplikat kategori untuk user
	var existing models.Category
	if err := database.DB.Where("user_id = ? AND LOWER(name) = LOWER(?) AND type = ?", uid, body.Name, body.Type).First(&existing).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "category already exists"})
	}

	cat := models.Category{UserID: uid, Name: body.Name, Type: body.Type}
	if err := database.DB.Create(&cat).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "create failed", "detail": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"id":   cat.ID,
		"name": cat.Name,
		"type": cat.Type,
	})
}

func GetCategories(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var cats []models.Category
	if err := database.DB.Where("user_id = ?", uid).Find(&cats).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query failed", "detail": err.Error()})
	}

	return c.JSON(cats)
}

func UpdateCategory(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")

	var cat models.Category
	if err := database.DB.Where("id = ? AND user_id = ?", id, uid).First(&cat).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}

	var body struct {
		Name *string `json:"name"`
		Type *string `json:"type"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if body.Name != nil {
		if strings.TrimSpace(*body.Name) == "" {
			return c.Status(400).JSON(fiber.Map{"error": "name cannot be empty"})
		}
		// Cek duplikat nama kategori
		var existing models.Category
		if err := database.DB.Where("user_id = ? AND LOWER(name) = LOWER(?) AND type = ?", uid, *body.Name, cat.Type).First(&existing).Error; err == nil && existing.ID != cat.ID {
			return c.Status(400).JSON(fiber.Map{"error": "category already exists"})
		}
		cat.Name = *body.Name
	}
	if body.Type != nil {
		t := strings.ToLower(*body.Type)
		if t != "income" && t != "expense" {
			return c.Status(400).JSON(fiber.Map{"error": "type must be income or expense"})
		}
		cat.Type = t
	}

	if err := database.DB.Save(&cat).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update failed", "detail": err.Error()})
	}

	return c.JSON(fiber.Map{
		"id":   cat.ID,
		"name": cat.Name,
		"type": cat.Type,
	})
}

func DeleteCategory(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")

	// Cek apakah kategori masih dipakai di transaksi
	var count int64
	database.DB.Model(&models.Transaction{}).Where("category_id = ? AND user_id = ?", id, uid).Count(&count)
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "category is in use"})
	}

	tx := database.DB.Where("id = ? AND user_id = ?", id, uid).Delete(&models.Category{})
	if tx.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "delete failed", "detail": tx.Error.Error()})
	}
	if tx.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}

	return c.JSON(fiber.Map{"message": "deleted"})
}
