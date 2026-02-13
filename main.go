package main

import (
	"log"
	"net/http"
	"ticket-system/handler"
	"ticket-system/repository"
	"ticket-system/service"
	"time" // 시간 설정을 위해 추가

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

	// [추가] DB 커넥션 풀 설정 (Too many connections 에러 방지)
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("커넥션 풀 설정 실패: ", err)
	}
	// 동시에 열 수 있는 최대 연결 수 (MySQL max_connections보다 작게 설정)
	sqlDB.SetMaxOpenConns(100)
	// 유휴 상태(대기 중)로 유지할 최대 연결 수
	sqlDB.SetMaxIdleConns(50)
	// 연결이 유지될 최대 시간
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 3. Repository 생성
	redisRepo := &repository.RedisRepository{Client: rdb}
	mysqlRepo := &repository.MySQLRepository{DB: db}

	// 4. Service 조립
	svc := service.NewTicketService(redisRepo, mysqlRepo)

	// 5. Handler 조립
	h := handler.NewTicketHandler(svc)

	// 6. 서버 설정 및 실행
	mux := http.NewServeMux()
	mux.Handle("/ticket", h)

	// [추가] 서버 자체에도 타임아웃을 설정
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("서버가 :8080 포트에서 시작되었습니다 (커넥션 풀 설정 완료)...")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("서버 시작 실패: ", err)
	}
}
