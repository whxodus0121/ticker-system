package repository

import (
	"context"
	"time"
)

// LockRepository는 '분산 락' 기능을 수행하기 위한 규격입니다.
type LockRepository interface {
	Lock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
}

// TicketRepository는 '티켓 데이터'에 접근하기 위한 규격입니다.
type TicketRepository interface {
	GetStock(name string) (int, error)
	DecreaseStock(name string) error
	SavePurchase(userID string, ticketName string) error           // 구매 목록 저장
	ExistsPurchase(userID string, ticketName string) (bool, error) //구매 여부 확인
}
