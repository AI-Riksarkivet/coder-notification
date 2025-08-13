package main

import (
	"context"
	"dagger/slack/internal/dagger"
	"fmt"
)

type Slack struct{}

// Build the application container
func (m *Slack) Build(
	// +defaultPath="/"
	source *dagger.Directory,
) *dagger.Container {
	nodeCache := dag.CacheVolume("node")
	return dag.Container().
		From("node:21-slim").
		WithDirectory("/app", source).
		WithMountedCache("/root/.npm", nodeCache).
		WithWorkdir("/app/slack").
		WithExec([]string{"npm", "install"}).
		WithExposedPort(6000).
		WithDefaultArgs([]string{"node", "app.js"})
}

// Run the application locally
func (m *Slack) Run(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// Slack bot token
	// +optional
	slackBotToken string,
	// Slack signing secret
	// +optional
	slackSigningSecret string,
) (*dagger.Service, error) {
	container := m.Build(source)

	if slackBotToken != "" {
		container = container.WithEnvVariable("SLACK_BOT_TOKEN", slackBotToken)
	}
	if slackSigningSecret != "" {
		container = container.WithEnvVariable("SLACK_SIGNING_SECRET", slackSigningSecret)
	}

	return container.
		WithEnvVariable("PORT", "6000").
		WithExec([]string{"node", "app.js"}).
		AsService(), nil
}

// Publish the application container to Docker Hub
func (m *Slack) Publish(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// Registry username
	username string,
	// Registry password/token
	password string,
	// Docker Hub repository (e.g., "username/slack-notification")
	repository string,
	// Tag for the image
	// +optional
	// +default="latest"

	tag string,
) (string, error) {
	if tag == "" {
		tag = "latest"
	}

	imageRef := repository + ":" + tag

	container := m.Build(source)

	addr, err := container.WithRegistryAuth(registry, username, dag.SetSecret("registry-password", password)).Publish(ctx, destination)
	if err != nil {
		return "", fmt.Errorf("failed to publish image: %w", err)
	}
	return fmt.Sprintf("Successfully built and pushed image: %s", addr), nil

}
