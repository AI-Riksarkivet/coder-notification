package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()

	// Initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Get build arguments from environment variables or use defaults
	imageTag := os.Getenv("IMAGE_TAG")
	if imageTag == "" {
		imageTag = "slack-bolt-coder:latest"
	}

	registry := os.Getenv("REGISTRY")
	if registry != "" {
		imageTag = fmt.Sprintf("%s/%s", registry, imageTag)
	}

	// Get the source directory
	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{
			"node_modules/",
			".git/",
			"*.go",
			"go.mod",
			"go.sum",
			".env",
			".dagger/",
		},
	})

	// Build the container
	container := client.Container().
		From("node:20-alpine").
		WithWorkdir("/app").
		WithFile("/app/package.json", src.File("package.json")).
		WithFile("/app/package-lock.json", src.File("package-lock.json")).
		WithExec([]string{"npm", "ci", "--only=production"}).
		WithFile("/app/app.js", src.File("app.js")).
		WithExec([]string{"addgroup", "-g", "1001", "-S", "nodejs"}).
		WithExec([]string{"adduser", "-S", "nodejs", "-u", "1001"}).
		WithUser("nodejs").
		WithExposedPort(6000).
		WithEntrypoint([]string{"node", "app.js"})

	// Export the container image
	_, err = container.Export(ctx, imageTag)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully built image: %s\n", imageTag)

	// Optional: Push to registry if PUSH_TO_REGISTRY is set
	if os.Getenv("PUSH_TO_REGISTRY") == "true" && registry != "" {
		fmt.Printf("Pushing image to registry: %s\n", imageTag)
		_, err = container.Publish(ctx, imageTag)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Successfully pushed image: %s\n", imageTag)
	}
}