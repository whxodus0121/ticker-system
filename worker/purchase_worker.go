package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings" // 메시지 접두사 확인을 위해 추가
	"ticket-system/repository"

	"github.com/go-sql-driver/mysql"
	"github.com/segmentio/kafka-go"
)

type PurchaseWorker struct {
	Reader     *kafka.Reader
	TicketRepo repository.TicketRepository
}

func NewPurchaseWorker(brokers []string, topic string, groupID string, tr repository.TicketRepository) *PurchaseWorker {
	return &PurchaseWorker{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		TicketRepo: tr,
	}
}

func (w *PurchaseWorker) Start() {
	fmt.Println("🚀 Kafka Consumer Worker 시작... [예매 저장/취소 처리 대기 중]")

	for {
		// 1. Kafka 메시지 읽기
		m, err := w.Reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("❌ 메시지 읽기 에러: %v", err)
			continue
		}

		userID := string(m.Key)
		messageVal := string(m.Value)

		// 2. 메시지 타입에 따른 분기 처리 (취소 vs 저장)
		if strings.HasPrefix(messageVal, "CANCEL:") {
			// [취소 로직] CANCEL: 접두사를 떼고 실제 티켓 이름을 가져옵니다.
			ticketName := strings.TrimPrefix(messageVal, "CANCEL:")
			w.handleCancel(userID, ticketName)
		} else {
			// [저장 로직] 기존 예매 저장 방식 유지
			ticketName := messageVal
			w.handleSave(userID, ticketName)
		}
	}
}

// handleSave는 기존의 저장 로직을 담당합니다.
func (w *PurchaseWorker) handleSave(userID string, ticketName string) {
	saved, err := w.TicketRepo.SavePurchase(userID, ticketName)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			log.Printf("⚠️ [중복 저장 스킵] 유저 %s는 이미 처리되었습니다.", userID)
		} else {
			log.Printf("🚨 [저장 실패] 유저 %s 서버 에러: %v", userID, err)
		}
	} else if !saved {
		log.Printf("⚠️ [중복 저장 스킵] 유저 %s는 이미 처리되었습니다.", userID)
	} else {
		fmt.Printf("✅ [저장 성공] 유저 %s의 티켓 정보 MySQL 저장 완료\n", userID)
	}
}

// handleCancel은 DB에서 구매 내역을 삭제하는 로직을 담당합니다.
func (w *PurchaseWorker) handleCancel(userID string, ticketName string) {
	err := w.TicketRepo.DeletePurchase(userID, ticketName)
	if err != nil {
		log.Printf("🚨 [취소 실패] 유저 %s의 데이터 삭제 중 에러: %v", userID, err)
	} else {
		fmt.Printf("🗑️ [취소 성공] 유저 %s의 구매 내역 DB 삭제 완료\n", userID)
	}
}
