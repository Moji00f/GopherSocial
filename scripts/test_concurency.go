package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type UpdatePayload struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

func main() {

	var wg sync.WaitGroup
	wg.Add(2)
	postId := 3
	content := "NEW CONTENT FROM USER B"
	title := "NEW TITLE FROM USER A"
	go updatePost(postId, UpdatePayload{Title: &title}, &wg)
	go updatePost(postId, UpdatePayload{Content: &content}, &wg)
	wg.Wait()

}

func updatePost(postId int, p UpdatePayload, wg *sync.WaitGroup) {
	defer wg.Done()

	url := fmt.Sprintf("http://localhost:8080/v1/posts/%d", postId)
	b, _ := json.Marshal(p)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err != nil {
		log.Println("Error Creating request: ", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request: ", err)
		return
	}

	defer res.Body.Close()

	fmt.Println("Update response status: ", res.Status)

}
