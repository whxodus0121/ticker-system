package main

import (
	"fmt"
	"net/http"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	totalUsers := 1000 // 동시에 시도할 가상 사용자 수

	for i := 0; i < totalUsers; i++ {
		wg.Add(1)
		go func(user int) {
			defer wg.Done()

			// 1. 사용자별 고유 ID를 먼저 만듭니다. (예: user_1, user_2...)
			userID := fmt.Sprintf("user_%d", user)

			for {
				// 2. [중요] URL 뒤에 ?user_id= 값을 붙여서 서버에 알려줍니다.
				url := fmt.Sprintf("http://localhost:8080/ticket?user_id=%s", userID)

				resp, err := http.Get(url)
				if err != nil {
					fmt.Printf("사용자 %d: 연결 에러 - %v\n", user, err)
					return
				}

				// StatusCode로 상태 확인
				switch resp.StatusCode {
				case http.StatusOK: // 200 OK
					fmt.Printf("사용자 %d: ★ 예매 성공! (ID: %s)\n", user, userID)
					resp.Body.Close()
					return

				case http.StatusGone: // 410 Gone (매진)
					fmt.Printf("사용자 %d: [품절] 매진되었습니다. 시도를 중단합니다.\n", user)
					resp.Body.Close()
					return

				case http.StatusConflict: // 409 Conflict (락 획득 실패)
					resp.Body.Close()
					// 재시도

				default:
					fmt.Printf("사용자 %d: 기타 응답 (%d)\n", user, resp.StatusCode)
					resp.Body.Close()
					return
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("------------------------------------")
	fmt.Println("모든 테스트가 종료되었습니다.")
	fmt.Println("------------------------------------")
}
