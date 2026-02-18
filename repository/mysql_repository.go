package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DB 테이블 구조와 매핑될 구조체
type Ticket struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Stock int
}

type Purchase struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	UserID     string    `gorm:"column:user_id;not null"`
	TicketName string    `gorm:"column:ticket_name;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

// Gorm에게 이 구조체가 사용할 테이블 이름을 명시적으로 알려줍니다
func (Purchase) TableName() string {
	return "purchases"
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

func (r *MySQLRepository) ExistsPurchase(userID string, ticketName string) (bool, error) {
	var count int64
	// purchases 테이블에서 해당 유저와 티켓이 있는지 COUNT를 세어봅니다.
	err := r.DB.Table("purchases").
		Where("user_id = ? AND ticket_name = ?", userID, ticketName).
		Count(&count).Error

	return count > 0, err
}

func (r *MySQLRepository) DeletePurchase(userID string, ticketName string) error {
	// GORM을 사용하여 조건에 맞는 데이터를 삭제합니다.
	// Unscoped()를 붙이지 않으면 Soft Delete가 설정된 경우 실제 삭제가 안 될 수 있으므로 확실히 지우기 위해 사용합니다.
	result := r.DB.Unscoped().Where("user_id = ? AND ticket_name = ?", userID, ticketName).Delete(&Purchase{})

	if result.Error != nil {
		return result.Error
	}

	// 실제로 삭제된 행이 0개라면 취소할 내역이 없는 것입니다.
	if result.RowsAffected == 0 {
		return fmt.Errorf("취소할 내역이 없습니다 (유저: %s)", userID)
	}

	return nil
}
