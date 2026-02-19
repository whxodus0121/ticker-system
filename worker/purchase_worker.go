package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"ticket-system/repository"

	"github.com/go-sql-driver/mysql" // MySQL 에러 번호를 확인하기 위해 필요
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
	fmt.Println("🚀 Kafka Consumer Worker 시작... MySQL 저장 대기 중")

	for {
		// 1. Kafka 메시지 읽기
		m, err := w.Reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("❌ 메시지 읽기 에러: %v", err)
			continue
		}

		userID := string(m.Key)
		ticketName := string(m.Value)

		// 2. MySQL에 실제 저장
		saved, err := w.TicketRepo.SavePurchase(userID, ticketName)

		if err != nil {
			// MySQL의 Duplicate Entry (1062) 에러인지 확인
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
				log.Printf("⚠️ [중복 스킵] 유저 %s는 이미 처리된 당첨자입니다.", userID)
			} else {
				log.Printf("🚨 [실패] 유저 %s의 티켓 저장 중 서버 에러: %v", userID, err)
			}
		} else if !saved {
			// 에러는 없지만 중복(INSERT IGNORE 등)으로 인해 저장되지 않은 경우
			log.Printf("⚠️ [중복 스킵] 유저 %s는 이미 처리된 당첨자입니다.", userID)
		} else {
			// 성공한 경우
			fmt.Printf("✅ [성공] 유저 %s의 티켓 정보 MySQL 저장 완료\n", userID)
		}
	}
}
