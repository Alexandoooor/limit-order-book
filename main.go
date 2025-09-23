package main

import (
	"database/sql"
	"flag"
	"limit-order-book/engine"
	"limit-order-book/server"
	"log"
	"os"
	"strconv"
	_ "github.com/mattn/go-sqlite3"
)

var (
	port      = flag.Int("port", 3000, "HTTP port")
)

func main() {
	flag.Parse()
	addr := ":" + strconv.Itoa(*port)

	output := os.Stdout
	if logFile := os.Getenv("LOGFILE"); logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		output = f
		log.Printf("LimitOrderBook running on http://localhost%s\n", addr)
		log.Printf("Logging to file: %s\n", logFile)
	} else {
		log.Println("Logging to Stdout")
	}
	logger := log.New(output, "", log.LstdFlags|log.Lshortfile)
	engine.Logger = logger
	server.Logger = logger


	db, err := sql.Open("sqlite3", "orderbook.db")
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	err = InitDB(db)
	if err != nil {
		logger.Fatal(err)
	}
	ob := engine.NewOrderBook(db)
	defer ob.SaveToDB()

	logger.Printf("LimitOrderBook running on http://localhost%s\n", addr)
	if err := server.Serve(addr, ob); err != nil {
		logger.Fatal(err)
	}

}

func InitDB(db *sql.DB) error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS levels (
	    side INTEGER,
	    price INTEGER,
	    volume INTEGER,
	    count INTEGER,
	    next_level_price INTEGER,
	    PRIMARY KEY (side, price)
	);

	CREATE TABLE IF NOT EXISTS orders (
	    id TEXT PRIMARY KEY,
	    side INTEGER,
	    size INTEGER,
	    remaining INTEGER,
	    price INTEGER,
	    time DATETIME,
	    level_price INTEGER,
	    level_side INTEGER,
	    prev_order_id TEXT,
	    next_order_id TEXT
	);

	CREATE TABLE IF NOT EXISTS trades (
	    price INTEGER,
	    size INTEGER,
	    time DATETIME,
	    buyerID TEXT,
	    sellerID TEXT
	);`
	_, err := db.Exec(sqlStmt)
	return err
}
