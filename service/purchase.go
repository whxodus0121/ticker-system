package service

import (
	"context"
	"log"
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

func (s *TicketService) BuyTicket(userID string) (bool, int) {
	ctx := context.Background()
	lockKey := "ticket_lock"
	ticketName := "concert_2026"

	// 1. [v3 추가] 락을 잡기 전 1차 확인 (이미 산 사람은 바로 거절해서 서버 부하 감소)
	exists, _ := s.TicketRepo.ExistsPurchase(userID, ticketName)
	if exists {
		log.Printf("[중복차단] 사용자 %s는 이미 구매했습니다.", userID)
		return false, -1 // 중복 구매를 나타내는 특수한 값 -1 반환
	}

	// 2. 재고 확인
	stock, _ := s.TicketRepo.GetStock(ticketName)
	if stock <= 0 {
		return false, 0
	}

	// 3. 분산 락 시도
	for i := 0; i < 10; i++ {
		ok, _ := s.LockRepo.Lock(ctx, lockKey, time.Second*2)
		if ok {
			defer s.LockRepo.Unlock(ctx, lockKey)

			// 4. [v3 추가] 락을 얻은 후 2차 확인 (동시성 방어)
			// 아주 짧은 찰나에 두 번 클릭했을 경우, 락 안에서 한 번 더 걸러줍니다.
			exists, _ := s.TicketRepo.ExistsPurchase(userID, ticketName)
			if exists {
				return false, -1
			}

			currentStock, _ := s.TicketRepo.GetStock(ticketName)
			if currentStock > 0 {
				err := s.TicketRepo.DecreaseStock(ticketName)
				if err == nil {
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
