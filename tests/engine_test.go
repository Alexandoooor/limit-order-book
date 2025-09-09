package tests

import (
	"testing"
	"limit-order-book/engine"
)

func TestHighestBid(t *testing.T) {

	ob := engine.NewOrderBook()

	ob.HighestBid

	for x := 88; x <= 92; x++ {
		ob.AddOrder(engine.Buy, x, 1)
	}

	// input := `
	// 	let five = 5;
	// `
	//
	// tests := []struct {
	// 	expectedType    token.TokenType
	// 	expectedLiteral string
	// }{
	// 	{token.Let, "let"},
	// 	{token.Identifier, "five"},
	// 	{token.Assign, "="},
	// 	{token.Int, "5"},
	// 	{token.Semicolon, ";"},
	// }
	//
	// l := New(input)
	//
	// for i, tt := range tests {
	// 	tok := l.NextToken()
	// 	if tok.Type != tt.expectedType {
	// 		t.Fatalf("tests[%d] - tokenType wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
	// 	}
	//
	// 	if tok.Literal != tt.expectedLiteral {
	// 		t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
	// 	}
	// }
}
