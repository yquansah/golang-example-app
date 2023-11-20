package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"dagger.io/dagger"
	"github.com/docker/docker/client"
)

func main() {
	if err := build(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func build(ctx context.Context) error {
	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}

	daggerClient, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer daggerClient.Close()

	dir := daggerClient.Host().Directory("/Users/yoofiquansah/Documents/twoheads/platform-engineeering/argocd-sample-project/golang-example-app")

	golang := daggerClient.Container().From("golang:1.21.3-alpine3.18")

	golang = golang.
		WithDirectory("/app", dir, dagger.ContainerWithDirectoryOpts{
			Exclude: []string{".gitignore", "build/", "go.work"},
		}).
		WithWorkdir("/app").
		WithExposedPort(8000).
		WithExec([]string{"go", "build", "-o", "bin/sample-app", "src/main.go"}).
		WithEntrypoint([]string{"./bin/sample-app"})

	_, err = golang.Export(ctx, "build-image.tar")
	if err != nil {
		return err
	}

	fi, err := os.Open("build-image.tar")
	if err != nil {
		return err
	}
	defer fi.Close()

	resp, err := dockerClient.ImageLoad(ctx, fi, false)
	if err != nil {
		return err
	}

	if resp.JSON {
		decoder := json.NewDecoder(resp.Body)
		var last map[string]string
		for {
			err := decoder.Decode(&last)
			if errors.Is(err, io.EOF) {
				break
			}
		}

		stream, ok := last["stream"]
		if !ok {
			return errors.New("local: parsing response: stream not found")
		}

		id := strings.TrimSpace(stream[strings.Index(stream, "sha256:"):])
		if err := dockerClient.ImageTag(ctx, id, "hello-world:latest"); err != nil {
			return fmt.Errorf("local: tag image: %w", err)
		}
	} else {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		fmt.Println("Load Response:", string(data))
	}

	return nil
}
