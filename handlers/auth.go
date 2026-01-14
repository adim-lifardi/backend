// handlers/auth.go
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"finance/database"
	"finance/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// FR-01: Register email/password
func Register(c *fiber.Ctx) error {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if body.Email == "" || body.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "email/password required"})
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
	user := models.User{
		Name:         body.Name,
		Email:        body.Email,
		PasswordHash: string(hash),
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "register failed"})
	}
	return c.Status(201).JSON(fiber.Map{"id": user.ID, "email": user.Email})
}

// FR-02: Login
func Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	var user models.User
	if err := database.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})
	t, _ := token.SignedString([]byte(jwtSecret))
	return c.JSON(fiber.Map{"token": t, "user": fiber.Map{"id": user.ID, "email": user.Email, "name": user.Name}})
}

// FR-02: Logout (client should delete token; server can implement blacklist if needed)
func Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "logout client-side: delete token"})
}

// FR: Google login/register
func GoogleLogin(c *fiber.Ctx) error {
	var body struct {
		IDToken  string `json:"id_token"`
		ClientID string `json:"client_id"` // optional: validate audience
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if body.IDToken == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id_token required"})
	}

	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + body.IDToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.Status(401).JSON(fiber.Map{"error": "google token verify failed"})
	}
	defer resp.Body.Close()

	var info struct {
		Aud     string `json:"aud"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid google token payload"})
	}
	// TODO: validate info.Aud == expected ClientID

	var user models.User
	if err := database.DB.Where("email = ?", info.Email).First(&user).Error; err != nil {
		user = models.User{Name: info.Name, Email: info.Email, PhotoURL: info.Picture}
		database.DB.Create(&user)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})
	t, _ := token.SignedString([]byte(jwtSecret))
	return c.JSON(fiber.Map{"token": t, "user": fiber.Map{"id": user.ID, "email": user.Email, "name": user.Name, "photo_url": user.PhotoURL}})
}
