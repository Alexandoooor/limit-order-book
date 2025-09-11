package main

import (
	"flag"
	"fmt"
	"limit-order-book/engine"
)

func main() {
	flag.Parse()

	ob := engine.NewOrderBook()

	// id := ob.AddOrder(engine.Buy, 88, 1)
	// ob.PrintOrder(id)

	for x := 88; x <= 92; x++ {
		ob.AddOrder(engine.Buy, x, 1)
		fmt.Println(ob)
	}

	// add existing level
	id := ob.AddOrder(engine.Buy, 90, 1)
	ob.PrintOrder(id)
	fmt.Println(ob)
	ob.PrintOrderBook()

	// ob.AddOrder(engine.Sell, 82, 834)
}

// cb = 92
// cb.nextlevel = 91
// cb.nextlevel.nextlevel = 89 -> 90
// newLevel = 90

// newLevel = 90
// 92 91 89
//    x


// newLevel = 90
// newLevel.nextLevel = 89
