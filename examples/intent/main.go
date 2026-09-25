package main

import "neurogo/pkg/runtime"

import (
	"fmt"
)

/* [NeuroGo] train "intent_model.gow" extracted */

func handleQuery(query string) {
	fmt.Printf("
[User Input] %s
", query)

	{
	_ngoMatch := runtime.Match("intent_model.gow", query)
	switch {

	case _ngoMatch.Label == "Refund" && _ngoMatch.Score >= 0.70:
		fmt.Println(">> [Routing] Connecting to Refund & Billing Specialist...")
	case _ngoMatch.Label == "Delivery" && _ngoMatch.Score >= 0.70:
		fmt.Println(">> [Routing] Launching Real-time Shipment Tracking...")
	case _ngoMatch.Label == "Inquiry" && _ngoMatch.Score >= 0.70:
		fmt.Println(">> [Routing] Directing to Product FAQ & Support Bot...")
	default:
		fmt.Println(">> [Routing] Low confidence query. Routing to General Helpdesk.")
	
	}
}
}

func main() {
	fmt.Println("=== NeuroGo Intelligent Branching Demo ===")

	testQueries := []string{
		"Please cancel my order and issue a full refund",
		"Where is my package? Track shipment please",
		"Do you have this jacket in size medium?",
		"What should I have for lunch today?",
	}

	for _, q := range testQueries {
		handleQuery(q)
	}
}
