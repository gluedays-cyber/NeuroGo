package main

import "neurogo/pkg/runtime"

import (
	"fmt"
)

/* [NeuroGo] train "intent_model.gow" extracted */

func handleQuery(query string) {
	fmt.Printf("\n[User Input] %s\n", query)

	{
	_ngoMatch := runtime.Match("intent_model.gow", query)
	switch {

	case _ngoMatch.Label == "Refund" && _ngoMatch.Score >= 0.70:
		fmt.Println(">> [Routing] 환불/결제취소 전문 상담사로 연결합니다.")
	case _ngoMatch.Label == "Delivery" && _ngoMatch.Score >= 0.70:
		fmt.Println(">> [Routing] 실시간 배송 추적 시스템을 실행합니다.")
	case _ngoMatch.Label == "Inquiry" && _ngoMatch.Score >= 0.70:
		fmt.Println(">> [Routing] 상품 상세 FAQ 및 Q&A 봇으로 안내합니다.")
	default:
		fmt.Println(">> [Routing] 명확하지 않은 문의입니다. 기본 고객센터로 연결합니다.")
	
	}
}
}

func main() {
	fmt.Println("=== NeuroGo 지능형 분기 데모 실행 ===")

	testQueries := []string{
		"주문 결제한 거 환불해주세요",
		"송장번호 배송 조회 좀 해주세요",
		"이거 사이즈 재고 있나요?",
		"오늘 점심 뭐 먹지?",
	}

	for _, q := range testQueries {
		handleQuery(q)
	}
}
