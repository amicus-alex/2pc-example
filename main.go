package main

import (
	"context"
	_pc "dt/2pc"
	"dt/domain"
	repo2 "dt/repository"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func main() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	db := repo2.InitDatabase()
	cTx := repo2.NewCustomerRepo(db)
	oTx := repo2.NewOrderRepo(db)
	tpcCoordinator := _pc.NewCoordinator(cTx, oTx)
	router.POST("/2pc/order", func(c *gin.Context) {
		var orderData domain.OrderRequest
		err := c.BindJSON(&orderData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(c, 10*time.Second)
		defer cancel()

		err = tpcCoordinator.Run(ctx, orderData)
		if err != nil {
			c.JSON(500, gin.H{
				"message": fmt.Sprintf("err: %s", err.Error()),
			})
		} else {
			c.JSON(200, gin.H{
				"message": "success",
			})
		}
	})

	router.Run(":8080")
}
