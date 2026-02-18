package main

import (
	"fmt"
	"log"
	"net/http"
	"ticket-system/handler"
	"ticket-system/repository"
	"ticket-system/service"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. Redis 연결 설정
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 2. MySQL 연결 설정
	dsn := "root:password123@tcp(127.0.0.1:3306)/ticket_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB 연결 실패: ", err)
	}

	// [추가] DB 커넥션 풀 설정
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

	// 4. Service 조립
	svc := service.NewTicketService(redisRepo, mysqlRepo)

	// 5. Handler 조립
	h := handler.NewTicketHandler(svc)

	// 6. 서버 설정 및 경로 등록
	mux := http.NewServeMux()

	// [기존] 예매 핸들러 (h는 ServeHTTP가 구현되어 있어야 함)
	mux.Handle("/ticket", h)

	// [추가] 취소 핸들러 등록
	// 핸들러 패키지에 별도의 CancelHandler를 만들지 않았다면,
	// 아래와 같이 즉석에서 핸들러 함수를 등록할 수 있습니다.
	mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "user_id가 필요합니다"}`)
			return
		}

		// 서비스 레이어의 CancelTicket 호출
		success, message := svc.CancelTicket(userID)

		if !success {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error": "%s"}`, message)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "%s"}`, message)
	})

	// 7. 서버 실행 설정
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("서버가 :8080 포트에서 시작되었습니다 (예매: /ticket, 취소: /cancel)...")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("서버 시작 실패: ", err)
	}
}
