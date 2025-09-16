package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	config := &genai.GenerateImagesConfig{
		NumberOfImages: 4,
	}

	generateImagesResponse, err := client.Models.GenerateImages(
		ctx,
		"imagen-3.0-generate-002",
		"Robot holding a red skateboard",
		config,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Generated %d images successfully", len(generateImagesResponse.GeneratedImages))

	// Create images directory if it doesn't exist
	err = os.MkdirAll("images", 0755)
	if err != nil {
		log.Fatal(err)
	}

	for i, image := range generateImagesResponse.GeneratedImages {
		name := fmt.Sprintf("images/google-imagen-%s.png", uuid.New().String())
		err = os.WriteFile(name, image.Image.ImageBytes, 0644)
		if err != nil {
			log.Printf("Error saving image %d: %v", i+1, err)
		} else {
			log.Printf("Saved image %d to: %s", i+1, name)
		}
	}
}
