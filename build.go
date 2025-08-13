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

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	imageTag := os.Getenv("IMAGE_TAG")
	if imageTag == "" {
		imageTag = "slack-bolt-coder:latest"
	}

	registry := os.Getenv("REGISTRY")
	if registry != "" {
		imageTag = fmt.Sprintf("%s/%s", registry, imageTag)
	}

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

	_, err = container.Export(ctx, imageTag)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully built image: %s\n", imageTag)

	if os.Getenv("PUSH_TO_REGISTRY") == "true" && registry != "" {
		fmt.Printf("Pushing image to registry: %s\n", imageTag)
		_, err = container.Publish(ctx, imageTag)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Successfully pushed image: %s\n", imageTag)
	}
}
