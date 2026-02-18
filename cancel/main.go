package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	// 취소를 시도할 유저 범위 (예: user_0부터 user_19까지 20명 취소)
	cancelStart := 0
	cancelEnd := 19

	fmt.Printf("--- %d번부터 %d번 유저까지 취소 테스트 시작 ---\n", cancelStart, cancelEnd)

	for i := cancelStart; i <= cancelEnd; i++ {
		wg.Add(1)
		go func(user int) {
			defer wg.Done()

			userID := fmt.Sprintf("user_%d", user)
			// 취소 API 호출 (URL에 user_id를 실어서 보냄)
			url := fmt.Sprintf("http://localhost:8080/cancel?user_id=%s", userID)

			// DELETE 또는 POST 등 서버에서 설정한 메서드에 맞춰 http.Get 혹은 http.NewRequest 사용
			// 여기서는 테스트 편의상 Get으로 작성합니다.
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("사용자 %d: 연결 에러 - %v\n", user, err)
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK: // 200 OK
				fmt.Printf("사용자 %d: [성공] 취소가 완료되었습니다. (ID: %s)\n", user, userID)

			case http.StatusBadRequest: // 400 (내역 없음 등)
				var result map[string]string
				json.NewDecoder(resp.Body).Decode(&result)
				fmt.Printf("사용자 %d: [실패] %s (ID: %s)\n", user, result["error"], userID)

			default:
				fmt.Printf("사용자 %d: 기타 응답 (%d) (ID: %s)\n", user, resp.StatusCode, userID)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("------------------------------------")
	fmt.Println("취소 테스트가 종료되었습니다.")
	fmt.Println("이제 Redis 재고와 MySQL purchases 테이블을 확인해보세요.")
	fmt.Println("------------------------------------")
}
