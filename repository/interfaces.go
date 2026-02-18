package repository

import (
	"context"
	"time"
)

// LockRepository는 '분산 락' 기능을 수행하기 위한 규격입니다.
type LockRepository interface {
	Lock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
	DecreaseStock(ctx context.Context, ticketName string) (int, error)
	AddPurchasedUser(ctx context.Context, ticketName string, userID string) error
	IsUserPurchased(ctx context.Context, ticketName string, userID string) (bool, error)
	IncreaseStock(ctx context.Context, ticketName string) (int, error)               // 재고 +1
	RemovePurchasedUser(ctx context.Context, ticketName string, userID string) error // 명단 삭제
}

// TicketRepository는 '티켓 데이터'에 접근하기 위한 규격입니다.
type TicketRepository interface {
	GetStock(name string) (int, error)
	DecreaseStock(name string) error
	SavePurchase(userID string, ticketName string) error           // 구매 목록 저장
	ExistsPurchase(userID string, ticketName string) (bool, error) //구매 여부 확인
	DeletePurchase(userID string, ticketName string) error
}
