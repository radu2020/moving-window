package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

func sendRequest() int {
	resp, _ := http.Get("http://localhost:3000/count")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data HTTPResponse
	json.Unmarshal(body, &data)
	return data.Count
}
func TestMainProgram(t *testing.T) {
	go main()

	// Wait for DB setup
	time.Sleep(3 * time.Second)

	t.Run("should return correct request amount for each request", func(t *testing.T) {
		var result int
		for i := 0; i < 3; i++ {
			result = sendRequest()
			assertCorrectMessage(t, result, i+1)
		}
	})

	t.Run("should return correct request amount after 60 seconds have passed", func(t *testing.T) {
		log.Printf("Waiting 61 seconds before making new request...")
		time.Sleep(61 * time.Second)
		response := sendRequest()
		assertCorrectMessage(t, response, 1)
	})
}

func assertCorrectMessage(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %d want %d", got, want)
	}
}
