package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const url = "http://localhost:8080/cotacao"
const filePath = "./client/cotacao.txt"

type Quotation struct {
	Bid string `json:"bid"`
}

func logConfig() *os.File {
	file, err := os.OpenFile("./client/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
	log.SetPrefix("CLIENT: ")

	return file
}

func getQuotation(quotation *Quotation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error:", err)
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, quotation)
	if err != nil {
		return err
	}
	return nil
}

func saveData(quotation *Quotation) error {
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fileData := fmt.Sprint("Dólar: ", quotation.Bid)

	_, err = file.Write([]byte(fileData))
	if err != nil {
		return err
	}
	return nil
}

func main() {
	file := logConfig()
	defer file.Close()

	var quotation Quotation

	err := getQuotation(&quotation)
	if err != nil {
		panic(err)
	}

	err = saveData(&quotation)
	if err != nil {
		panic(err)
	}
}
