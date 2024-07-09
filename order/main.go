package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type Order struct {
	OrderID   string `gorm:"primary_key"`
	UserID    string
	Amount    uint
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TransactionLog struct {
	TransactionID string `gorm:"primary_key"`
	UserID        string
	Amount        uint
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

var db *gorm.DB

func initDB() {
	var err error
	dsn := "host=orderdb user=youruser password=yourpassword dbname=yourdb port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&Order{}, &TransactionLog{}); err != nil {
		panic(err)
	}
}

func prepareHandler(c *gin.Context) {
	var req struct {
		TransactionID string `json:"transaction_id"`
		UserID        string `json:"user_id"`
		Amount        uint   `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("JSON data: %s\n", req)

	order := Order{
		OrderID:   req.TransactionID,
		UserID:    req.UserID,
		Amount:    req.Amount,
		Status:    "PENDING",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log := TransactionLog{
		TransactionID: req.TransactionID,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Status:        "PREPARE",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := db.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "prepared"})
}

func commitHandler(c *gin.Context) {
	var req struct {
		TransactionID string `json:"transaction_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&Order{}).Where("order_id = ?", req.TransactionID).
		Updates(map[string]interface{}{"status": "COMMITTED", "updated_at": time.Now()}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&TransactionLog{}).Where("transaction_id = ?", req.TransactionID).
		Updates(map[string]interface{}{"status": "COMMIT", "updated_at": time.Now()}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "committed"})
}

func abortHandler(c *gin.Context) {
	var req struct {
		TransactionID string `json:"transaction_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&Order{}).Where("order_id = ?", req.TransactionID).
		Updates(map[string]interface{}{"status": "ABORTED", "updated_at": time.Now()}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&TransactionLog{}).Where("transaction_id = ?", req.TransactionID).
		Updates(map[string]interface{}{"status": "ABORT", "updated_at": time.Now()}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "aborted"})
}

func main() {
	initDB()

	r := gin.Default()
	r.POST("/order/prepare", prepareHandler)
	r.POST("/order/commit", commitHandler)
	r.POST("/order/abort", abortHandler)

	r.Run(":8082")
}
