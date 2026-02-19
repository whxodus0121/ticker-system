package main

import (
	"fmt"
	"log"
	"net/http"
	"ticket-system/handler"
	"ticket-system/repository"
	"ticket-system/service"
	"ticket-system/worker" // [추가] 워커 패키지
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. Redis 연결 설정 (docker-compose의 ticket-redis 사용)
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 2. MySQL 연결 설정 (docker-compose의 ticket-mysql 사용)
	// 비밀번호와 DB명은 docker-compose.yml 설정과 동일하게 유지
	dsn := "root:password123@tcp(127.0.0.1:3306)/ticket_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB 연결 실패: ", err)
	}

	// DB 커넥션 풀 설정
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("커넥션 풀 설정 실패: ", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 3. Repository 생성
	redisRepo := &repository.RedisRepository{Client: rdb}
	mysqlRepo := &repository.MySQLRepository{DB: db}

	// [추가] Kafka Repository 생성 (Producer 역할)
	kafkaRepo := repository.NewKafkaRepository([]string{"localhost:9092"}, "ticket-topic")

	// 4. [수정] Service 조립 (오류 해결: kafkaRepo 추가)
	svc := service.NewTicketService(redisRepo, mysqlRepo, kafkaRepo)

	// 5. [추가] Kafka Consumer Worker 실행
	// 서버가 켜질 때 백그라운드에서 Kafka 메시지를 읽어 DB에 저장합니다.
	purchaseWorker := worker.NewPurchaseWorker(
		[]string{"localhost:9092"},
		"ticket-topic",
		"purchase-group",
		mysqlRepo,
	)
	go purchaseWorker.Start() // 고루틴으로 실행

	// 6. Handler 조립
	h := handler.NewTicketHandler(svc)

	// 7. 서버 설정 및 경로 등록
	mux := http.NewServeMux()
	mux.Handle("/ticket", h)

	// 취소 핸들러 등록
	mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "user_id가 필요합니다"}`)
			return
		}

		success, message := svc.CancelTicket(userID)
		if !success {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error": "%s"}`, message)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "%s"}`, message)
	})

	// 8. 서버 실행 설정
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("🚀 비동기 티켓 시스템 서버 시작 (:8080)...")
	log.Println("- 예매: /ticket")
	log.Println("- 취소: /cancel")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("서버 시작 실패: ", err)
	}
}
