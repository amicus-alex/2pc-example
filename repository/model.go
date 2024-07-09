package repository

import "gorm.io/gorm"

type Customer struct {
	*gorm.Model
	UserID  string `gorm:"type:varchar(255);column:user_id;not null"`
	Balance uint   `gorm:"type:int unsigned;column:balance;not null"`
}

type Order struct {
	*gorm.Model
	UserID string `gorm:"type:varchar(255);column:user_id;not null"`
	Amount uint   `gorm:"type:int unsigned;column:number;not null"`
}
