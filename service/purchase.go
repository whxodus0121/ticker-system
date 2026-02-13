package service

import (
	"context"
	"ticket-system/repository"
	"time"
)

type TicketService struct {
	// [변경] 특정 구조체가 아닌 인터페이스 타입을 사용
	LockRepo   repository.LockRepository
	TicketRepo repository.TicketRepository
}

// [추가] 외부에서 부품(Repo)을 받아 서비스를 조립하는 생성자 함수
func NewTicketService(lr repository.LockRepository, tr repository.TicketRepository) *TicketService {
	return &TicketService{
		LockRepo:   lr,
		TicketRepo: tr,
	}
}

// [수정] 인자에 userID string 추가
func (s *TicketService) BuyTicket(userID string) (bool, int) {
	ctx := context.Background()
	lockKey := "ticket_lock"
	ticketName := "concert_2026"

	// 1. 사전 재고 확인
	stock, _ := s.TicketRepo.GetStock(ticketName)
	if stock <= 0 {
		return false, 0
	}

	// 2. 분산 락 로직 (재시도 포함)
	for i := 0; i < 10; i++ {
		ok, _ := s.LockRepo.Lock(ctx, lockKey, time.Second*2)
		if ok {
			// 함수가 끝날 때(성공이든 실패든) 락을 해제합니다.
			defer s.LockRepo.Unlock(ctx, lockKey)

			// 락 획득 후 다시 한번 정확한 재고 확인
			currentStock, _ := s.TicketRepo.GetStock(ticketName)
			if currentStock > 0 {
				// A. 재고 감소 시도
				err := s.TicketRepo.DecreaseStock(ticketName)
				if err == nil {
					// B. [추가] 재고 감소 성공 시 구매 기록 저장
					// 여기서 에러가 나더라도 이미 재고는 줄었으므로 로직상 성공으로 보거나,
					// 엄격하게 하려면 여기서 에러 시 재고를 다시 늘리는(Rollback) 처리를 합니다.
					_ = s.TicketRepo.SavePurchase(userID, ticketName)

					return true, currentStock - 1
				}
			}
			return false, 0
		}
		time.Sleep(time.Millisecond * 50)
	}

	lastStock, _ := s.TicketRepo.GetStock(ticketName)
	return false, lastStock
}
