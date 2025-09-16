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

type CraiyonRequest struct {
	Prompt string `json:"prompt"`
}

type CraiyonResponse struct {
	Images []string `json:"images"` // Base64 encoded images
}

func main() {
	// Create generation request
	genRequest := CraiyonRequest{
		Prompt: "Robot holding a red skateboard",
	}

	jsonData, err := json.Marshal(genRequest)
	if err != nil {
		log.Fatal(err)
	}

	// Create HTTP client with longer timeout since Craiyon can be slow
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	log.Printf("Starting image generation with Craiyon...")
	log.Printf("This may take 30-60 seconds...")

	// Create generation request
	req, err := http.NewRequest("POST", "https://backend.craiyon.com/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Research-Bot/1.0")

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

	var craiyonResp CraiyonResponse
	err = json.Unmarshal(body, &craiyonResp)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Generated %d images successfully", len(craiyonResp.Images))

	// Create images directory if it doesn't exist
	err = os.MkdirAll("images", 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Save all images (Craiyon typically returns 9)
	imageCount := len(craiyonResp.Images)

	// Save images
	for i := 0; i < imageCount; i++ {
		imageData, err := base64.StdEncoding.DecodeString(craiyonResp.Images[i])
		if err != nil {
			log.Printf("Error decoding image %d: %v", i+1, err)
			continue
		}

		filename := fmt.Sprintf("images/craiyon-%s.png", uuid.New().String())
		err = os.WriteFile(filename, imageData, 0644)
		if err != nil {
			log.Printf("Error saving image %d: %v", i+1, err)
		} else {
			log.Printf("Saved image %d to: %s", i+1, filename)
		}
	}
}
