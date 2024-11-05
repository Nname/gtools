package docker

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"io"
	"os"
)

type Actions struct {
	Options []client.Opt
}

type Response struct {
	Stream string `json:"stream,omitempty"`
	Error  string `json:"error,omitempty"`
}

func (a Actions) Cli() (*client.Client, error) {
	return client.NewClientWithOpts(a.Options...)
}

func (a Actions) Ping() (types.Ping, error) {
	cli, err := a.Cli()
	if err != nil {
		return types.Ping{}, err
	}
	return cli.Ping(context.Background())
}

func readResponse(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var response Response
		line := scanner.Text()

		err := json.Unmarshal([]byte(line), &response)
		if err != nil {
			return err
		}
		if response.Stream != "" {
			_, _ = fmt.Fprint(os.Stdout, response.Stream)
		}
		if response.Error != "" {
			return errors.New(response.Error)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (a Actions) ImageBuild(buildContextDir, dockerfile string, tags []string) error {
	cli, err := a.Cli()
	if err != nil {
		return err
	}

	tarContext, err := archive.TarWithOptions(buildContextDir, &archive.TarOptions{})
	if err != nil {
		return err
	}

	buildResponse, err := cli.ImageBuild(
		context.Background(),
		tarContext,
		types.ImageBuildOptions{
			Tags:           tags,
			SuppressOutput: false,
			Dockerfile:     dockerfile,
			Remove:         true,
		},
	)
	if err != nil {
		return err
	}
	defer buildResponse.Body.Close()
	return readResponse(buildResponse.Body)
}

func (a Actions) ImagePush(images, user, auth string, clean bool) error {
	cli, err := a.Cli()
	if err != nil {
		return err
	}
	authConfig := registry.AuthConfig{
		Username: user,
		Password: auth,
	}
	authConfigBytes, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}
	encodedAuth := base64.URLEncoding.EncodeToString(authConfigBytes)
	pushOptions := image.PushOptions{
		RegistryAuth: encodedAuth,
	}
	pushResponse, err := cli.ImagePush(context.Background(), images, pushOptions)
	if err != nil {
		return err
	}
	defer pushResponse.Close()

	// clean image
	if clean {
		removeOptions := image.RemoveOptions{
			Force:         true,
			PruneChildren: true,
		}
		_, err := cli.ImageRemove(context.Background(), images, removeOptions)
		if err != nil {
			return err
		}
	}

	return readResponse(pushResponse)
}

func (a Actions) BuildPush(buildContextDir, dockerfile string, tags []string, user, auth string, clean bool) error {
	err := a.ImageBuild(buildContextDir, dockerfile, tags)
	if err != nil {
		return err
	}
	for _, item := range tags {
		err := a.ImagePush(item, user, auth, clean)
		if err != nil {
			return err
		}
	}
	return nil
}
