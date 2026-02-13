package repository

import (
	"gorm.io/gorm"
)

// DB 테이블 구조와 매핑될 구조체
type Ticket struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Stock int
}

type MySQLRepository struct {
	DB *gorm.DB
}

// 재고 차감 (UPDATE 쿼리 실행)
func (r *MySQLRepository) DecreaseStock(name string) error {
	// stock > 0 일 때만 1을 깎는 안전한 쿼리
	return r.DB.Model(&Ticket{}).
		Where("name = ? AND stock > 0", name).
		Update("stock", gorm.Expr("stock - 1")).Error
}

// 현재 재고 확인 (SELECT 쿼리 실행)
func (r *MySQLRepository) GetStock(name string) (int, error) {
	var ticket Ticket
	err := r.DB.Where("name = ?", name).First(&ticket).Error
	return ticket.Stock, err
}

func (r *MySQLRepository) SavePurchase(userID string, ticketName string) error {
	// purchases 테이블에 사용자 ID와 티켓명을 삽입합니다.
	return r.DB.Exec("INSERT INTO purchases (user_id, ticket_name) VALUES (?, ?)", userID, ticketName).Error
}
