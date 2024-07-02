package sn_test

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	sn "github.com/ekzyis/snappy"
)

var (
	c = testClient()
)

func TestQueryItems(t *testing.T) {
	var (
		cursor *sn.ItemsCursor
		err    error
	)

	if cursor, err = c.Items(nil); err != nil {
		t.Error(err)
		return
	}

	if len(cursor.Items) == 0 {
		t.Error("items cursor empty")
		return
	}
}

func TestMutationCreateComment(t *testing.T) {
	var (
		parentId = 349
		text     = "test comment"
		err      error
	)

	// TODO: return result, invoice, paymentMethod from CreateComment and run assertions on that
	if _, err = c.CreateComment(parentId, text); err != nil {
		t.Error(err)
		return
	}
}

func TestMutationPostDiscussion(t *testing.T) {
	var (
		title = "test discussion"
		text  = "test discussion text"
		sub   = "bitcoin"
		err   error
	)

	// TODO: return result, invoice, paymentMethod from CreateComment and run assertions on that
	if _, err = c.PostDiscussion(title, text, sub); err != nil {
		t.Error(err)
		return
	}
}

func TestMutationPostLink(t *testing.T) {
	var (
		url   = "https://stacker.news"
		title = "test discussion"
		text  = "test discussion text"
		sub   = "bitcoin"
		err   error
	)

	// TODO: return result, invoice, paymentMethod from CreateComment and run assertions on that
	if _, err = c.PostLink(url, title, text, sub); err != nil {
		t.Error(err)
		return
	}
}

func testClient() *sn.Client {
	loadEnv()

	baseUrl, set := os.LookupEnv("TEST_SN_BASE_URL")
	if !set {
		baseUrl = "http://localhost:3000"
	}
	log.Printf("baseUrl=%s\n", baseUrl)

	apiKey, set := os.LookupEnv("TEST_SN_API_KEY")
	if !set {
		log.Fatalf("TEST_SN_API_KEY is not set")
	}
	log.Printf("apiKey=%s\n", apiKey)

	return sn.NewClient(
		sn.WithBaseUrl(baseUrl),
		sn.WithApiKey(apiKey),
	)
}

func loadEnv() {
	var (
		f   *os.File
		s   *bufio.Scanner
		err error
	)

	if f, err = os.Open(".env"); err != nil {
		log.Fatalf("error opening .env: %v", err)
	}
	defer f.Close()

	s = bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		line := s.Text()
		parts := strings.SplitN(line, "=", 2)

		// Check if we have exactly 2 parts (key and value)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		} else {
			log.Fatalf(".env: invalid line: %s\n", line)
		}
	}

	// Check for errors during scanning
	if err = s.Err(); err != nil {
		fmt.Println("error scanning .env:", err)
	}

}
