// models/models.go
package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Name         string    `gorm:"size:100;not null"`
	Email        string    `gorm:"size:150;unique;not null"`
	PasswordHash string    `gorm:"size:255"`
	PhotoURL     string    `json:"PhotoURL"`
	PhoneNumber  string    `json:"PhoneNumber"`
	Instagram    string    `json:"Instagram"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

type Category struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Name      string    `gorm:"size:100;not null"`
	Type      string    `gorm:"size:20;not null"` // "income" or "expense"
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Transaction struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"not null;index"`
	CategoryID uint      `gorm:"not null;index"`
	Amount     float64   `gorm:"type:decimal(15,2);not null"`
	Date       time.Time `gorm:"not null;index"`
	Note       string    `gorm:"type:text"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

type Budget struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null;index"`
	CategoryID  uint      `gorm:"not null;index"`
	LimitAmount float64   `gorm:"type:decimal(15,2);not null"`
	StartDate   time.Time `gorm:"not null"`
	EndDate     time.Time `gorm:"not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`

	SpentAmount float64 `gorm:"-" json:"SpentAmount"`
	Status      string  `gorm:"-" json:"Status"`
}

type Backup struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	BackupURL string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Notification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
