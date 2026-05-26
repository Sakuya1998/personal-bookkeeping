package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Ledger 账本
type Ledger struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Name         string     `gorm:"size:100;not null" json:"name"`
	Description  *string    `gorm:"size:500" json:"description"`
	BaseCurrency string     `gorm:"size:10;default:CNY" json:"base_currency"`
	Icon         *string    `gorm:"size:50" json:"icon"`
	Color        *string    `gorm:"size:20" json:"color"`
	IsArchived   bool       `gorm:"default:false" json:"is_archived"`
	SortOrder    int        `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (l *Ledger) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// Category 分类
type Category struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	LedgerID  *uuid.UUID `gorm:"type:uuid;index" json:"ledger_id"`
	Name      string     `gorm:"size:50;not null" json:"name"`
	Type      string     `gorm:"size:10;not null;check:type IN ('income','expense')" json:"type"`
	Icon      *string    `gorm:"size:50" json:"icon"`
	Color     *string    `gorm:"size:20" json:"color"`
	ParentID  *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	SortOrder int        `gorm:"default:0" json:"sort_order"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// Transaction 账单记录
type Transaction struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	LedgerID        uuid.UUID  `gorm:"type:uuid;index;not null" json:"ledger_id"`
	UserID          uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	CategoryID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"category_id"`
	Type            string     `gorm:"size:10;not null;check:type IN ('income','expense')" json:"type"`
	Amount          float64    `gorm:"type:decimal(18,2);not null" json:"amount"`
	Currency        string     `gorm:"size:10;default:CNY" json:"currency"`
	ExchangeRate    float64    `gorm:"type:decimal(18,8);default:1.0" json:"exchange_rate"`
	BaseAmount      float64    `gorm:"type:decimal(18,2);not null" json:"base_amount"`
	Description     *string    `gorm:"type:text" json:"description"`
	TransactionDate string     `gorm:"type:date;index;not null" json:"transaction_date"`
	Tags            *string    `gorm:"type:text" json:"tags"` // JSON array stored as string
	IsReconciled    bool       `gorm:"default:false" json:"is_reconciled"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	User     User     `gorm:"foreignKey:UserID" json:"-"`
	Ledger   Ledger   `gorm:"foreignKey:LedgerID" json:"-"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// ExchangeRate 汇率
type ExchangeRate struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FromCurrency string    `gorm:"size:10;not null" json:"from_currency"`
	ToCurrency   string    `gorm:"size:10;not null" json:"to_currency"`
	Rate         float64   `gorm:"type:decimal(18,8);not null" json:"rate"`
	Date         string    `gorm:"type:date;not null" json:"date"`
	Source       *string   `gorm:"size:50" json:"source"`
	CreatedAt    time.Time `json:"created_at"`
}

func (e *ExchangeRate) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (e *ExchangeRate) TableName() string {
	return "exchange_rates"
}
