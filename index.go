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

	response, err := client.Models.GenerateImages(
		ctx,
		"imagen-4.0-generate-001",
		"Robot holding a red skateboard",
		config,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(response)

	for _, image := range response.GeneratedImages {
		name := fmt.Sprintf("images/imagen-%s.png", uuid.New().String())
		_ = os.WriteFile(name, image.Image.ImageBytes, 0644)
	}
}
