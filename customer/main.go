package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type Customer struct {
	UserID          string `gorm:"primary_key"`
	Balance         uint
	ReservedBalance uint
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
	dsn := "host=customerdb user=youruser password=yourpassword dbname=yourdb port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&Customer{}, &TransactionLog{}); err != nil {
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

	var customer Customer
	if err := db.First(&customer, "user_id = ?", req.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if customer.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient funds"})
		return
	}

	if err := db.Model(&Customer{}).Where("user_id = ?", req.UserID).
		Updates(map[string]interface{}{"balance": gorm.Expr("balance - ?", req.Amount), "reserved_balance": gorm.Expr("reserved_balance + ?", req.Amount)}).Error; err != nil {
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

	var log TransactionLog
	if err := db.First(&log, "transaction_id = ?", req.TransactionID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&Customer{}).Where("user_id = ?", log.UserID).
		Update("reserved_balance", gorm.Expr("reserved_balance - ?", log.Amount)).Error; err != nil {
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

	var log TransactionLog
	if err := db.First(&log, "transaction_id = ?", req.TransactionID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&Customer{}).Where("user_id = ?", log.UserID).
		Updates(map[string]interface{}{"balance": gorm.Expr("balance + ?", log.Amount), "reserved_balance": gorm.Expr("reserved_balance - ?", log.Amount)}).Error; err != nil {
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

func createCustomerHandler(c *gin.Context) {
	var req struct {
		UserID  string `json:"user_id"`
		Balance uint   `json:"balance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer := Customer{
		UserID:          req.UserID,
		Balance:         req.Balance,
		ReservedBalance: 0,
	}

	if err := db.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "customer created"})
}

func main() {
	initDB()

	r := gin.Default()
	r.POST("/customer/prepare", prepareHandler)
	r.POST("/customer/commit", commitHandler)
	r.POST("/customer/abort", abortHandler)
	r.POST("/customer/create", createCustomerHandler)

	r.Run(":8081")
}
