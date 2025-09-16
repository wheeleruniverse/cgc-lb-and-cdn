package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type GenerationRequest struct {
	Height        int     `json:"height"`
	ModelID       string  `json:"modelId"`
	Prompt        string  `json:"prompt"`
	Width         int     `json:"width"`
	NumImages     int     `json:"num_images"`
	GuidanceScale float64 `json:"guidance_scale"`
}

type GenerationResponse struct {
	SDGenerationJob struct {
		GenerationID string `json:"generationId"`
	} `json:"sdGenerationJob"`
}

type GenerationStatusResponse struct {
	GenerationsByPK struct {
		GeneratedImages []struct {
			URL string `json:"url"`
			ID  string `json:"id"`
		} `json:"generated_images"`
		Status string `json:"status"`
	} `json:"generations_by_pk"`
}

func main() {
	apiKey := os.Getenv("LEONARDO_API_KEY")
	if apiKey == "" {
		log.Fatal("LEONARDO_API_KEY environment variable is required")
	}

	// Create generation request
	genRequest := GenerationRequest{
		ModelID:       "6bef9f1b-29cb-40c7-b9df-32b51c1f67d3", // Leonardo Creative model
		Prompt:        "Robot holding a red skateboard",
		Height:        1024,
		Width:         1024,
		NumImages:     4,
		GuidanceScale: 7, // recommended
	}

	jsonData, err := json.Marshal(genRequest)
	if err != nil {
		log.Fatal(err)
	}

	// Create generation
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://cloud.leonardo.ai/api/rest/v1/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", "Bearer "+apiKey)
	req.Header.Add("content-type", "application/json")

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

	var genResponse GenerationResponse
	err = json.Unmarshal(body, &genResponse)
	if err != nil {
		log.Fatal(err)
	}

	generationID := genResponse.SDGenerationJob.GenerationID
	log.Printf("Started generation with ID: %s", generationID)
	log.Printf("Waiting for generation to complete...")

	// Poll for completion
	for {
		time.Sleep(5 * time.Second)

		statusReq, err := http.NewRequest("GET", fmt.Sprintf("https://cloud.leonardo.ai/api/rest/v1/generations/%s", generationID), nil)
		if err != nil {
			log.Fatal(err)
		}

		statusReq.Header.Add("accept", "application/json")
		statusReq.Header.Add("authorization", "Bearer "+apiKey)

		statusResp, err := client.Do(statusReq)
		if err != nil {
			log.Fatal(err)
		}

		if statusResp.StatusCode != http.StatusOK {
			statusResp.Body.Close()
			log.Printf("Status check failed, retrying...")
			continue
		}

		statusBody, err := io.ReadAll(statusResp.Body)
		statusResp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		var statusResponse GenerationStatusResponse
		err = json.Unmarshal(statusBody, &statusResponse)
		if err != nil {
			log.Fatal(err)
		}

		status := statusResponse.GenerationsByPK.Status
		log.Printf("Generation status: %s", status)

		if status == "COMPLETE" {
			images := statusResponse.GenerationsByPK.GeneratedImages
			log.Printf("Generated %d images successfully", len(images))

			// Create images directory if it doesn't exist
			err = os.MkdirAll("images", 0755)
			if err != nil {
				log.Fatal(err)
			}

			// Download and save images
			for i, image := range images {
				err := downloadImage(image.URL, fmt.Sprintf("images/leonardo-ai-%s.png", uuid.New().String()))
				if err != nil {
					log.Printf("Error downloading image %d: %v", i+1, err)
				} else {
					log.Printf("Saved image %d to: images/leonardo-ai-%s.png", i+1, uuid.New().String())
				}
			}
			break
		} else if status == "FAILED" {
			log.Fatal("Generation failed")
		}
	}
}

func downloadImage(url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
