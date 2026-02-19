package service

import (
	"context"
	"fmt"
	"ticket-system/repository"
)

type TicketService struct {
	// [변경] 특정 구조체가 아닌 인터페이스 타입을 사용
	LockRepo   repository.LockRepository
	TicketRepo repository.TicketRepository
	KafkaRepo  *repository.KafkaRepository
}

// [추가] 외부에서 부품(Repo)을 받아 서비스를 조립하는 생성자 함수
func NewTicketService(lr repository.LockRepository, tr repository.TicketRepository, kr *repository.KafkaRepository) *TicketService {
	return &TicketService{
		LockRepo:   lr,
		TicketRepo: tr,
		KafkaRepo:  kr,
	}
}

func (s *TicketService) BuyTicket(userID string) (bool, int) {
	ctx := context.Background()
	ticketName := "concert_2026"

	// 1. Redis에서 차감 (선착순 검증)
	remaining, err := s.LockRepo.DecreaseStock(ctx, ticketName)
	if err != nil || remaining < 0 {
		return false, 0
	}

	// 2. [변경] MySQL 저장 대신 Kafka에 "예매 성공 이벤트"를 던집니다.
	// 이제 더 이상 여기서 DB가 느려지는 것을 기다리지 않습니다.
	err = s.KafkaRepo.PublishPurchase(userID, ticketName)
	if err != nil {
		fmt.Printf("Kafka 메시지 전송 실패: %v\n", err)
		return false, remaining
	}

	// 3. 사용자에게는 즉시 성공 응답을 보냅니다.
	return true, remaining
}

func (s *TicketService) CancelTicket(userID string) (bool, string) {
	ctx := context.Background()
	ticketName := "concert_2026"

	// 1. DB에서 구매 내역이 있는지 확인 (영수증 확인)
	// (이 작업을 위해 TicketRepo에 DeletePurchase 메서드가 필요합니다)
	err := s.TicketRepo.DeletePurchase(userID, ticketName)
	if err != nil {
		return false, "구매 내역을 찾을 수 없거나 취소에 실패했습니다."
	}

	// 2. DB 삭제 성공 후, Redis 재고 복구
	_, err = s.LockRepo.IncreaseStock(ctx, ticketName)
	if err != nil {
		// 이 경우 DB는 지워졌는데 Redis 복구가 실패한 상황 (중요 로그 필요)
		fmt.Printf("[위험] Redis 재고 복구 실패: %v\n", err)
	}

	// 3. Redis 구매자 명단에서 삭제 (다시 살 수 있게 해줌)
	_ = s.LockRepo.RemovePurchasedUser(ctx, ticketName, userID)

	return true, "취소가 완료되었습니다."
}
