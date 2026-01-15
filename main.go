// main.go
package main

import (
	"log"
	"os"

	"finance/database"
	"finance/handlers"
	"finance/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	database.Connect()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Public routes
	app.Post("/auth/register", handlers.Register)
	app.Post("/auth/login", handlers.Login)
	app.Post("/auth/google", handlers.GoogleLogin)
	app.Post("/auth/logout", handlers.Logout)
	app.Static("/uploads", "./uploads")

	// Protected routes
	app.Use(middleware.JWT(handlers.JwtSecret()))

	// Me
	app.Get("/users/me", handlers.GetMe)
	app.Put("/users/me", handlers.UpdateMe)
	app.Put("/users/me/password", handlers.ChangePassword)

	// Categories
	app.Post("/categories", handlers.CreateCategory)
	app.Get("/categories", handlers.GetCategories)
	app.Put("/categories/:id", handlers.UpdateCategory)
	app.Delete("/categories/:id", handlers.DeleteCategory)

	// Transactions
	app.Post("/transactions", handlers.CreateTransaction)
	app.Get("/transactions", handlers.GetTransactions)
	app.Get("/transactions/:id", handlers.GetTransaction)
	app.Put("/transactions/:id", handlers.UpdateTransaction)
	app.Delete("/transactions/:id", handlers.DeleteTransaction)
	// Reports
	app.Get("/reports/summary", handlers.GetSummary)
	app.Get("/reports/monthly", handlers.GetMonthlySummary)
	app.Get("/reports/expense-by-category", handlers.GetExpenseByCategory)

	// Budgets
	app.Post("/budgets", handlers.CreateBudget)
	app.Get("/budgets", handlers.GetBudgets)
	app.Put("/budgets/:id", handlers.UpdateBudget)
	app.Delete("/budgets/:id", handlers.DeleteBudget)
	app.Get("/budgets/status", handlers.GetBudgetStatus)
	app.Get("/budgets/:id/detail", handlers.GetBudgetDetail)

	// Notifications

	app.Post("/notifications", handlers.CreateNotification)
	app.Get("/notifications", handlers.GetNotifications)
	app.Get("/notifications/:id", handlers.GetNotificationDetail)
	app.Delete("/notifications/:id", handlers.DeleteNotification)

	app.Get("/profile", handlers.GetProfile)
	app.Put("/profile", handlers.UpdateProfile)
	app.Post("/profile/photo", handlers.UploadPhoto)
	
	app.Listen(":" + os.Getenv("PORT"))
}


