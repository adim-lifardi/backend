// services/budget_service.go
package services

import (
	"finance/models"

	"gorm.io/gorm"
)

func CalculateSpentAmount(db *gorm.DB, budget *models.Budget) {
	var total float64
	db.Model(&models.Transaction{}).
		Where("user_id = ? AND category_id = ? AND date BETWEEN ? AND ?",
			budget.UserID, budget.CategoryID, budget.StartDate, budget.EndDate).
		Select("SUM(amount)").Scan(&total)

	budget.SpentAmount = total
	if total > budget.LimitAmount {
		budget.Status = "Over Budget"
	} else {
		budget.Status = "On Track"
	}
}
