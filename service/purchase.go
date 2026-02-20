package service

import (
	"context"
	"fmt"
	"ticket-system/repository"
)

type TicketService struct {
	LockRepo   repository.LockRepository
	TicketRepo repository.TicketRepository
	KafkaRepo  *repository.KafkaRepository
}

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

	// 1. [중요] 중복 구매 확인 (이미 샀는지 Redis 명단 확인)
	isPurchased, err := s.LockRepo.IsUserPurchased(ctx, ticketName, userID)
	if err != nil || isPurchased {
		// 이미 구매했다면 여기서 컷!
		return false, 0
	}

	// 2. Redis에서 차감 (선착순 검증)
	remaining, err := s.LockRepo.DecreaseStock(ctx, ticketName)
	if err != nil || remaining < 0 {
		return false, 0
	}

	// 3. [중요] Redis 구매자 명단에 추가 (이게 있어야 나중에 취소 가능)
	err = s.LockRepo.AddPurchasedUser(ctx, ticketName, userID)
	if err != nil {
		fmt.Printf("Redis 명단 추가 실패: %v\n", err)
		// 명단 추가 실패 시 깎았던 재고 다시 복구 (정합성)
		s.LockRepo.IncreaseStock(ctx, ticketName)
		return false, 0
	}

	// 4. Kafka에 "예매 성공 이벤트" 발행
	err = s.KafkaRepo.PublishPurchase(userID, ticketName)
	if err != nil {
		fmt.Printf("Kafka 메시지 전송 실패: %v\n", err)
		// Kafka 전송 실패 시 롤백 (재고 복구 및 명단 삭제)
		s.LockRepo.IncreaseStock(ctx, ticketName)
		s.LockRepo.RemovePurchasedUser(ctx, ticketName, userID)
		return false, remaining
	}

	return true, remaining
}

func (s *TicketService) CancelTicket(userID string) (bool, string) {
	ctx := context.Background()
	ticketName := "concert_2026"

	// 1. Redis에서 구매한 유저인지 확인 (빠른 검증)
	isPurchased, err := s.LockRepo.IsUserPurchased(ctx, ticketName, userID)
	if err != nil || !isPurchased {
		return false, "구매 내역이 없거나 이미 취소되었습니다."
	}

	// 2. Redis 재고 복구 (+1)
	_, err = s.LockRepo.IncreaseStock(ctx, ticketName)
	if err != nil {
		fmt.Printf("[오류] 재고 복구 실패: %v\n", err)
		return false, "재고 복구 중 오류가 발생했습니다."
	}

	// 3. Redis 구매자 명단에서 삭제 (이래야 나중에 다시 살 수 있음)
	err = s.LockRepo.RemovePurchasedUser(ctx, ticketName, userID)
	if err != nil {
		fmt.Printf("[오류] 명단 삭제 실패: %v\n", err)
	}

	// 4. Kafka에 "취소 이벤트" 발행
	err = s.KafkaRepo.PublishCancel(userID, ticketName)
	if err != nil {
		fmt.Printf("[위험] 취소 메시지 전송 실패: %v\n", err)
	}

	return true, "취소 요청이 접수되었습니다. (재고 즉시 복구됨)"
}
