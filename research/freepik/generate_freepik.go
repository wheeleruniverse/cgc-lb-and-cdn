package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type FreepikRequest struct {
	Prompt      string `json:"prompt"`
	NumImages   int    `json:"num_images"`
	AspectRatio string `json:"aspect_ratio"`
}

type FreepikResponse struct {
	Data []struct {
		Base64 string `json:"base64"`
	} `json:"data"`
}

func main() {
	apiKey := os.Getenv("FREEPIK_API_KEY")
	if apiKey == "" {
		log.Fatal("FREEPIK_API_KEY environment variable is required")
	}

	// Create generation request
	genRequest := FreepikRequest{
		Prompt:      "Robot holding a red skateboard",
		NumImages:   4,
		AspectRatio: "square_1_1",
	}

	jsonData, err := json.Marshal(genRequest)
	if err != nil {
		log.Fatal(err)
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	log.Printf("Starting image generation with FreePik...")

	// Create generation request
	req, err := http.NewRequest("POST", "https://api.freepik.com/v1/ai/text-to-image", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-freepik-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Generation request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var freepikResp FreepikResponse
	err = json.Unmarshal(body, &freepikResp)
	if err != nil {
		log.Fatal(err)
	}

	images := freepikResp.Data
	log.Printf("Generated %d images successfully", len(images))

	// Create images directory if it doesn't exist
	err = os.MkdirAll("images", 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Save images from base64
	for i, image := range images {
		imageData, err := base64.StdEncoding.DecodeString(image.Base64)
		if err != nil {
			log.Printf("Error decoding image %d: %v", i+1, err)
			continue
		}

		filename := fmt.Sprintf("images/freepik-%s.png", uuid.New().String())
		err = os.WriteFile(filename, imageData, 0644)
		if err != nil {
			log.Printf("Error saving image %d: %v", i+1, err)
		} else {
			log.Printf("Saved image %d to: %s", i+1, filename)
		}
	}
}
