package main

import (
	"context"
	"encoding/json"
	"log"

	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	listModelsResponse, err := client.Models.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	imageGenModels := 0
	for _, item := range listModelsResponse.Items {
		modelName := item.Name
		if containsSubstring(modelName, "imagen") {
			imageGenModels++
			jsonData, err := json.MarshalIndent(item, "", "  ")
			if err != nil {
				log.Printf("Error marshaling item: %v", err)
			} else {
				log.Printf("Model: %s", string(jsonData))
			}
			log.Printf("---")
		}
	}

	log.Printf("Found %d Imagen models that support image generation", imageGenModels)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
