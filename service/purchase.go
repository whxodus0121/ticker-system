package service

import (
	"context"
	"ticket-system/repository"
	"time"
)

type TicketService struct {
	RedisRepo *repository.RedisRepository
	MySQLRepo *repository.MySQLRepository
}

// 리턴 타입을 (bool)에서 (bool, int)로 변경합니다.
func (s *TicketService) BuyTicket() (bool, int) {
	ctx := context.Background()
	lockKey := "ticket_lock"
	ticketName := "concert_2026"

	// 1. 재고 먼저 확인 (락 잡기 전에 재고 없으면 바로 컷)
	stock, _ := s.MySQLRepo.GetStock(ticketName)
	if stock <= 0 {
		return false, 0
	}

	// 2. Redis 락 획득 시도 (최대 10번 재시도)
	for i := 0; i < 10; i++ {
		ok, _ := s.RedisRepo.Lock(ctx, lockKey, time.Second*2)
		if ok {
			defer s.RedisRepo.Unlock(ctx, lockKey)

			// 3. 락 획득 후 다시 한번 재고 확인 (그사이 남이 가져갔을 수 있음)
			currentStock, _ := s.MySQLRepo.GetStock(ticketName)
			if currentStock > 0 {
				err := s.MySQLRepo.DecreaseStock(ticketName)
				if err == nil {
					return true, currentStock - 1
				}
			}
			return false, 0 // 재고 없음
		}
		// 락 실패 시 잠깐 대기
		time.Sleep(time.Millisecond * 50)
	}

	// 락 획득 최종 실패 (하지만 재고는 남아있을 수 있음)
	lastStock, _ := s.MySQLRepo.GetStock(ticketName)
	return false, lastStock
}
