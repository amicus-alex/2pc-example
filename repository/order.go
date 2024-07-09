package repository

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

type orderRepository struct {
	database *gorm.DB
}

func (o *orderRepository) Prepare(ctx context.Context, userId string, amount uint) TxI {
	select {
	case <-ctx.Done():
		return TxI{
			DB:  nil,
			Err: fmt.Errorf("timeout"),
		}
	default:
		return func() TxI {
			tx := o.database.Model(Order{}).Begin()
			var order = Order{
				UserID: userId,
				Amount: amount,
			}
			if err := tx.Create(&order).Error; err != nil {
				tx.Rollback()
				return TxI{
					DB:  nil,
					Err: err,
				}
			}
			return TxI{
				DB:  tx,
				Err: nil,
			}
		}()
	}
}

type OrderRepository interface {
	Prepare(ctx context.Context, userId string, amount uint) TxI
}

func NewOrderRepo(db *gorm.DB) OrderRepository {
	return &orderRepository{database: db}
}
