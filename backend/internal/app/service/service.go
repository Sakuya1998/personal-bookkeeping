package service

import (
	"personal-bookkeeping/internal/infra/database"
	"personal-bookkeeping/internal/infra/cache"
	"personal-bookkeeping/internal/infra/queue"

	"gorm.io/gorm"
)

// Service 是所有业务服务的基类，通过 DI 注入依赖。
type Service struct {
	DB    *gorm.DB
	Cache cache.Cache
	Queue queue.Queue
}

// NewService 创建 Service 实例。
func NewService() *Service {
	return &Service{
		DB:    database.GetDB(),
		Cache: cache.GetDefault(),
		Queue: queue.GetDefault(),
	}
}

// TransactionService 交易相关业务逻辑。
type TransactionService struct {
	*Service
}

func NewTransactionService(s *Service) *TransactionService {
	return &TransactionService{Service: s}
}

// LedgerService 账本相关业务逻辑。
type LedgerService struct {
	*Service
}

func NewLedgerService(s *Service) *LedgerService {
	return &LedgerService{Service: s}
}

// CategoryService 分类相关业务逻辑。
type CategoryService struct {
	*Service
}

func NewCategoryService(s *Service) *CategoryService {
	return &CategoryService{Service: s}
}

// BudgetService 预算相关业务逻辑。
type BudgetService struct {
	*Service
}

func NewBudgetService(s *Service) *BudgetService {
	return &BudgetService{Service: s}
}

// RecurringService 周期规则相关业务逻辑。
type RecurringService struct {
	*Service
}

func NewRecurringService(s *Service) *RecurringService {
	return &RecurringService{Service: s}
}

// ExchangeRateService 汇率相关业务逻辑。
type ExchangeRateService struct {
	*Service
}

func NewExchangeRateService(s *Service) *ExchangeRateService {
	return &ExchangeRateService{Service: s}
}

// ReportService 报表相关业务逻辑。
type ReportService struct {
	*Service
}

func NewReportService(s *Service) *ReportService {
	return &ReportService{Service: s}
}

// OC RService OCR 相关业务逻辑。
type OCRService struct {
	*Service
}

func NewOCRService(s *Service) *OCRService {
	return &OCRService{Service: s}
}
