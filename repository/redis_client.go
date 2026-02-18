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

	// 원자적 감소 연산
	val, err := r.Client.Decr(ctx, key).Result()
	if err != nil {
		return -1, err
	}

	// 로그 추가: 터미널에서 숫자가 줄어드는지 확인 가능
	// fmt.Printf("[Redis 확인] %s 잔여: %d\n", key, val)

	return int(val), nil
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
