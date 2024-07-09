package repository

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type customerRepository struct {
	database *gorm.DB
}

func (c *customerRepository) Prepare(ctx context.Context, userId string, amount uint) TxI {
	select {
	case <-ctx.Done():
		return TxI{
			DB:  nil,
			Err: fmt.Errorf("timeout"),
		}
	default:
		return func() TxI {
			tx := c.database.Model(Customer{}).Begin()
			var customer Customer
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userId).First(&customer).Error; err != nil {
				tx.Rollback()
				return TxI{
					DB:  nil,
					Err: err,
				}
			}
			if customer.Balance < amount {
				return TxI{
					DB:  nil,
					Err: fmt.Errorf("insufficient funds"),
				}
			}
			customer.Balance -= amount
			if err := tx.Where("user_id = ?", userId).Save(&customer).Error; err != nil {
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

type CustomerRepository interface {
	Prepare(ctx context.Context, userId string, amount uint) TxI
}

func NewCustomerRepo(db *gorm.DB) CustomerRepository {
	return &customerRepository{database: db}
}
