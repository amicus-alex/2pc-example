package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PrepareRequest struct {
	TransactionID string `json:"transaction_id"`
	UserID        string `json:"user_id"`
	Amount        uint   `json:"amount"`
}

type CommitOrAbortRequest struct {
	TransactionID string `json:"transaction_id"`
}

func sendPostRequest(url string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	return nil
}

func coordinatorOrderHandler(c *gin.Context) {
	transactionID := uuid.New().String()

	var req struct {
		UserID string `json:"user_id"`
		Amount uint   `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customerReq := PrepareRequest{
		TransactionID: transactionID,
		UserID:        req.UserID,
		Amount:        req.Amount,
	}
	orderReq := PrepareRequest{
		TransactionID: transactionID,
		UserID:        req.UserID,
		Amount:        req.Amount,
	}

	// Phase 1: Prepare phase
	if err := sendPostRequest("http://customer:8081/customer/prepare", customerReq); err != nil {
		abortReq := CommitOrAbortRequest{TransactionID: transactionID}
		sendPostRequest("http://customer:8081/customer/abort", abortReq)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := sendPostRequest("http://order:8082/order/prepare", orderReq); err != nil {
		abortReq := CommitOrAbortRequest{TransactionID: transactionID}
		sendPostRequest("http://customer:8081/customer/abort", abortReq)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Phase 2: Commit phase
	if err := sendPostRequest("http://customer:8081/customer/commit", CommitOrAbortRequest{TransactionID: transactionID}); err != nil {
		sendPostRequest("http://order:8082/order/abort", CommitOrAbortRequest{TransactionID: transactionID})
		sendPostRequest("http://customer:8081/customer/abort", CommitOrAbortRequest{TransactionID: transactionID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := sendPostRequest("http://order:8082/order/commit", CommitOrAbortRequest{TransactionID: transactionID}); err != nil {
		sendPostRequest("http://order:8082/order/abort", CommitOrAbortRequest{TransactionID: transactionID})
		sendPostRequest("http://customer:8081/customer/abort", CommitOrAbortRequest{TransactionID: transactionID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "order created successfully", "transaction_id": transactionID})
}

func main() {
	r := gin.Default()

	r.POST("/coordinator/order", coordinatorOrderHandler)

	time.Sleep(10 * time.Second)

	r.Run(":8080")
}
