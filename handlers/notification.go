package handlers

import (
	"fmt"
	"time"

	"finance/database"
	"finance/models"
	"finance/utils"

	"github.com/gofiber/fiber/v2"
)

// Tambah notifikasi otomatis (misalnya dipanggil dari budget handler)
func AddBudgetNotification(userID uint, category string) error {
	notif := models.Notification{
		UserID:    userID,
		Title:     "Budget Alert",
		Message:   fmt.Sprintf("Kategori %s sudah melebihi batas!", category),
		CreatedAt: time.Now(),
	}
	if err := database.DB.Create(&notif).Error; err != nil {
		return err
	}
	return nil
}

// Endpoint: buat notifikasi manual (POST /notifications)
func CreateNotification(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var body struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	notif := models.Notification{
		UserID:    uid,
		Title:     body.Title,
		Message:   body.Message,
		CreatedAt: time.Now(),
	}
	if err := database.DB.Create(&notif).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "create failed", "detail": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(notif)
}

// Endpoint: ambil semua notifikasi user (GET /notifications)
// handlers/notification.go
func GetNotifications(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var notifs []models.Notification
	if err := database.DB.Where("user_id = ?", uid).
		Order("created_at desc").
		Find(&notifs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "query failed", "detail": err.Error()})
	}

	return c.JSON(notifs)
}

// Endpoint: detail notifikasi (GET /notifications/:id)
func GetNotificationDetail(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")

	var notif models.Notification
	if err := database.DB.Where("id = ? AND user_id = ?", id, uid).First(&notif).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(notif)
}

// Endpoint: hapus notifikasi (DELETE /notifications/:id)
func DeleteNotification(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id := c.Params("id")

	tx := database.DB.Where("id = ? AND user_id = ?", id, uid).Delete(&models.Notification{})
	if tx.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "delete failed", "detail": tx.Error.Error()})
	}
	if tx.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}
