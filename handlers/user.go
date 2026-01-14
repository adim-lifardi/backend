// handlers/user.go
package handlers

import (
	"finance/database"
	"finance/models"
	"finance/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// FR-04: Get/Update profile
func GetMe(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "photo_url": user.PhotoURL})
}

func UpdateMe(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var body struct {
		Name     *string `json:"name"`
		Email    *string `json:"email"`
		PhotoURL *string `json:"photo_url"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	if body.Name != nil {
		user.Name = *body.Name
	}
	if body.Email != nil {
		user.Email = *body.Email
	}
	if body.PhotoURL != nil {
		user.PhotoURL = *body.PhotoURL
	}
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update failed"})
	}
	return c.JSON(fiber.Map{"message": "updated"})
}

// FR-05: Change password
func ChangePassword(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&body); err != nil || body.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	if user.PasswordHash != "" && bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.OldPassword)) != nil {
		return c.Status(401).JSON(fiber.Map{"error": "old password mismatch"})
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 12)
	user.PasswordHash = string(hash)
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update failed"})
	}
	return c.JSON(fiber.Map{"message": "password changed"})
}
