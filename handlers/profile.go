// handlers/profile.go
package handlers

import (
	"finance/database"
	"finance/models"
	"finance/utils"

	"github.com/gofiber/fiber/v2"
)

func GetProfile(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(fiber.Map{
		"ID":          user.ID,
		"Name":        user.Name,
		"Email":       user.Email,
		"PhotoURL":    user.PhotoURL,
		"PhoneNumber": user.PhoneNumber,
		"Instagram":   user.Instagram,
	})
}

func UpdateProfile(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var body struct {
		Name        string `json:"Name"`
		PhotoURL    string `json:"PhotoURL"`
		PhoneNumber string `json:"PhoneNumber"`
		Instagram   string `json:"Instagram"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	if body.Name != "" {
		user.Name = body.Name
	}
	if body.PhotoURL != "" {
		user.PhotoURL = body.PhotoURL
	}
	if body.PhoneNumber != "" {
		user.PhoneNumber = body.PhoneNumber
	}
	if body.Instagram != "" {
		user.Instagram = body.Instagram
	}

	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "update failed"})
	}

	return c.JSON(user)
}

func UploadPhoto(c *fiber.Ctx) error {
	uid, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	file, err := c.FormFile("photo")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no file uploaded"})
	}

	// Simpan file ke folder uploads
	savePath := "./uploads/" + file.Filename
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save file"})
	}

	// Buat URL publik yang bisa diakses client
	publicURL := "/uploads/" + file.Filename

	// Update user PhotoURL di database
	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	user.PhotoURL = publicURL
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update user photo"})
	}

	// Return JSON ke client
	return c.JSON(fiber.Map{"PhotoURL": publicURL})
}
