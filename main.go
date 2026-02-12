package main

import (
	"net/http"
	"ticket-system/handler"
	"ticket-system/repository"
	"ticket-system/service"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Redis 연결
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	// MySQL 연결 (비밀번호와 DB 이름 확인!)
	dsn := "root:password123@tcp(127.0.0.1:3306)/ticket_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("DB 연결 실패: " + err.Error())
	}

	// 레이어 조립
	redisRepo := &repository.RedisRepository{Client: rdb}
	mysqlRepo := &repository.MySQLRepository{DB: db}
	svc := &service.TicketService{
		RedisRepo: redisRepo,
		MySQLRepo: mysqlRepo,
	}
	h := &handler.TicketHandler{Service: svc}

	// 경로 설정 및 서버 시작
	http.Handle("/ticket", h)
	http.ListenAndServe(":8080", nil)
}
