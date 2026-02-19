package main

import (
	"log"
	"ticket-system/repository"
	"ticket-system/worker"

	"gorm.io/driver/mysql" // GORM용 드라이버
	"gorm.io/gorm"
)

func main() {
	// 1. GORM으로 MySQL 연결
	dsn := "root:password123@tcp(127.0.0.1:3306)/ticket_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB 연결 실패: %v", err)
	}

	// 2. 레포지토리 초기화
	ticketRepo := repository.NewMySQLRepository(db)

	// 3. 워커 생성 및 시작
	pWorker := worker.NewPurchaseWorker(
		[]string{"localhost:9092"},
		"ticket-topic",
		"ticket-group",
		ticketRepo,
	)

	pWorker.Start()
}
