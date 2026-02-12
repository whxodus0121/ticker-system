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
			for {
				resp, err := http.Get("http://localhost:8080/ticket")
				if err != nil {
					fmt.Printf("사용자 %d: 연결 에러 - %v\n", user, err)
					return
				}

				// StatusCode로 상태 확인
				switch resp.StatusCode {
				case http.StatusOK: // 200 OK
					fmt.Printf("사용자 %d: ★ 예매 성공!\n", user)
					resp.Body.Close()
					return // 성공했으므로 해당 고루틴 종료

				case http.StatusGone: // 410 Gone (매진)
					fmt.Printf("사용자 %d: [품절] 매진되었습니다. 시도를 중단합니다.\n", user)
					resp.Body.Close()
					return // 매진되었으므로 해당 고루틴 종료

				case http.StatusConflict: // 409 Conflict (락 획득 실패)
					// 계속 for 문을 돌며 다시 시도합니다.
					resp.Body.Close()

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
