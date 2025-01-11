package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

func extractMagic(htmlContent string) string {
	re := regexp.MustCompile(`<input[^>]*name=["']magic["'][^>]*value=["']([^"']+)["']`)
	matches := re.FindStringSubmatch(htmlContent)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractLocation(responseBody string) string {
	re := regexp.MustCompile(`window\.location="([^"]+)"`)

	matches := re.FindStringSubmatch(responseBody)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func main() {
	username := flag.String("username", "", "Your username")
	password := flag.String("password", "", "Your password")

	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Println("Usage: go run main.go -username=yourusername -password=yourpassword")
		return
	}

	serverURL := "https://fw.bits-pilani.ac.in:8090/login?="

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var resp *http.Response
	var err error
	for i := 0; i < 5; i++ {
		resp, err = client.Get(serverURL)
		if err == nil {
			break
		}
		fmt.Println("Connection failed, retrying...")
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		fmt.Println("Failed to connect after retries.")
		return
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	htmlContent := buf.String()

	magic := extractMagic(htmlContent)
	if magic == "" {
		fmt.Println("Couldn't find the magic value.")
		return
	}

	formData := url.Values{}
	formData.Set("username", *username)
	formData.Set("password", *password)
	formData.Set("magic", magic)
	formData.Set("4Tredir", "http://example.com/")

	encodedForm := formData.Encode()

	postURL := "https://fw.bits-pilani.ac.in:8090/"

	var respPost *http.Response
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("POST", postURL, bytes.NewBufferString(encodedForm))
		if err != nil {
			fmt.Println("Error creating POST request:", err)
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		respPost, err = client.Do(req)
		if err == nil {
			break
		}
		fmt.Println("Error posting data, retrying...")
		time.Sleep(3 * time.Second)
	}

	if respPost == nil {
		fmt.Println("Failed to post data after retries.")
		return
	}
	defer respPost.Body.Close()

	fmt.Println("POST Response Status:", respPost.Status)

	bufPost := new(bytes.Buffer)
	bufPost.ReadFrom(respPost.Body)
	responseBody := bufPost.String()
	// fmt.Println("POST Response Body:", responseBody)

	location := extractLocation(responseBody)

	if location != "" && regexp.MustCompile("keepalive").MatchString(location) {
		fmt.Println("logged in")
	} else {
		fmt.Println("invalid username/password")
	}
}
