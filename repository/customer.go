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

func (p *customerRepository) Credit(ctx context.Context, userId string, amount uint) error {
	tx := p.database.Model(Customer{}).Begin()
	err := error(nil)
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()
	var customer Customer
	if err = tx.Where("user_id = ?", userId).First(&customer).Error; err != nil {
		return err
	}
	customer.Balance += amount
	if err = tx.Where("user_id = ?", userId).Save(&customer).Error; err != nil {
		return err
	}
	return nil
}

func (p *customerRepository) Debit(ctx context.Context, userId string, amount uint) error {
	//time.Sleep(time.Duration(6) * time.Second)
	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout")
	default:
		return func() error {
			tx := p.database.Model(Customer{}).Begin()
			err := error(nil)
			defer func() {
				if err != nil {
					tx.Rollback()
				}
				tx.Commit()
			}()
			var customer Customer
			if err = tx.Where("user_id = ?", userId).First(&customer).Error; err != nil {
				return err
			}
			if customer.Balance < amount {
				return fmt.Errorf("insufficient funds")
			}
			customer.Balance -= amount
			if err = tx.Where("user_id = ?", userId).Save(&customer).Error; err != nil {
				return err
			}
			return nil
		}()
	}
}

func (p *customerRepository) GetById(ctx context.Context, userId string) (Customer, error) {
	var (
		customer = Customer{}
	)
	err := p.database.Model(Customer{}).Where("user_id = ?", userId).First(&customer).Error
	return customer, err
}

func (p *customerRepository) Prepare(ctx context.Context, userId string, amount uint) TxI {
	select {
	case <-ctx.Done():
		return TxI{
			DB:  nil,
			Err: fmt.Errorf("timeout"),
		}
	default:
		return func() TxI {
			tx := p.database.Model(Customer{}).Begin()
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
					Err: fmt.Errorf("thieu tien"),
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
	GetById(ctx context.Context, userId string) (Customer, error)
	Credit(ctx context.Context, userId string, amount uint) error
	Debit(ctx context.Context, userId string, amount uint) error
}

func NewCustomerRepo(db *gorm.DB) CustomerRepository {
	return &customerRepository{database: db}
}
