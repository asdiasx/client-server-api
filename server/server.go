package main

import (
	"context"
	sql "database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const url = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

var db *sql.DB

type CurrencyRate struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type ApiResponse struct {
	USDBRL CurrencyRate `json:"USDBRL"`
}

func logConfig() *os.File {
	file, err := os.OpenFile("./server/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
	log.SetPrefix("SERVER: ")

	return file
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./server/quotation.db")
	if err != nil {
		panic(err)
	}

	stmt := `
							create table if not exists currency_rates (
								id integer not null primary key autoincrement,
								code text,
								codein text,
								name text,
								high text,
								low text,
								varBid text,
								pctChange text,
								bid text,
								ask text,
								timestamp text,
								create_date text
							);
	`
	_, err = db.Exec(stmt)
	if err != nil {
		panic(err)
	}
	return db
}

func startServer() {
	http.HandleFunc("/cotacao", quotationHandler())

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func quotationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiResCh := make(chan string)
		apiErrCh := make(chan error)

		go func() {
			defer close(apiResCh)
			defer close(apiErrCh)

			quotation, err := getQuotation()
			if err != nil {
				apiErrCh <- err
				return
			}
			apiResCh <- quotation.USDBRL.Bid
		}()

		ctx := r.Context()

		select {
		case <-ctx.Done():
			log.Println("INFO: Request canceled by client")
			http.Error(w, "Request canceledby client", http.StatusRequestTimeout)
			return
		case apiErr := <-apiErrCh:
			log.Println("ERROR:", apiErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		case apiResult := <-apiResCh:
			result := map[string]string{
				"bid": apiResult,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			err := json.NewEncoder(w).Encode(result)
			if err != nil {
				log.Println("ERROR:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}
	}
}

func getQuotation() (*ApiResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	var quotation ApiResponse

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &quotation)
	if err != nil {
		return nil, err
	}

	go saveData(quotation)

	return &quotation, nil
}

func saveData(quotation ApiResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.Prepare(
		"INSERT INTO currency_rates(code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		log.Println("Error:", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		quotation.USDBRL.Code,
		quotation.USDBRL.Codein,
		quotation.USDBRL.Name,
		quotation.USDBRL.High,
		quotation.USDBRL.Low,
		quotation.USDBRL.VarBid,
		quotation.USDBRL.PctChange,
		quotation.USDBRL.Bid,
		quotation.USDBRL.Ask,
		quotation.USDBRL.Timestamp,
		quotation.USDBRL.CreateDate,
	)
	if err != nil {
		log.Println("Error:", err)
		return err
	}
	return nil
}

func main() {
	file := logConfig()
	defer file.Close()

	db = initDB()
	defer db.Close()

	startServer()
}
