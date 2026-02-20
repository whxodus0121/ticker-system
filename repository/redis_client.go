package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Client *redis.Client
}

// Lock: SetNX를 이용해 열쇠를 획득 시도
func (r *RedisRepository) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, "locked", expiration).Result()
}

// Unlock: 열쇠 반납
func (r *RedisRepository) Unlock(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

func (r *RedisRepository) DecreaseStock(ctx context.Context, ticketName string) (int, error) {
	key := "ticket_stock:" + ticketName

	// Lua 스크립트 작성
	// 1. 현재 재고(GET)를 가져와서 숫자로 변환합니다.
	// 2. 재고가 존재하고(stock) 0보다 크면(stock > 0) 1을 뺍니다(DECR).
	// 3. 재고가 없으면 깎지 않고 -1을 반환합니다.
	script := `
		local stock = redis.call("GET", KEYS[1])
		if stock and tonumber(stock) > 0 then
			return redis.call("DECR", KEYS[1])
		else
			return -1
		end
	`

	// Eval 명령어로 스크립트 실행
	val, err := r.Client.Eval(ctx, script, []string{key}).Int()
	if err != nil {
		return -1, err
	}

	// 재고 부족 상황 처리
	if val == -1 {
		return -1, nil // 서비스 계층에서 "매진"으로 판단할 수 있게 -1 반환
	}

	return val, nil
}

func (r *RedisRepository) AddPurchasedUser(ctx context.Context, ticketName string, userID string) error {
	// 구매자 명단 Key 예시: "purchased_users:concert_2026"
	key := "purchased_users:" + ticketName

	// Redis Set에 userID 추가
	err := r.Client.SAdd(ctx, key, userID).Err()
	return err
}

// (참고) IsUserPurchased도 아래와 같이 구현되어 있어야 합니다.
func (r *RedisRepository) IsUserPurchased(ctx context.Context, ticketName string, userID string) (bool, error) {
	key := "purchased_users:" + ticketName
	// Set에 해당 유저가 있는지 확인 (SIsMember)
	exists, err := r.Client.SIsMember(ctx, key, userID).Result()
	return exists, err
}

func (r *RedisRepository) IncreaseStock(ctx context.Context, ticketName string) (int, error) {
	key := "ticket_stock:" + ticketName
	val, err := r.Client.Incr(ctx, key).Result() // 재고 1 증가
	return int(val), err
}

func (r *RedisRepository) RemovePurchasedUser(ctx context.Context, ticketName string, userID string) error {
	key := "purchased_users:" + ticketName
	return r.Client.SRem(ctx, key, userID).Err() // 구매 명단에서 유저 삭제
}
