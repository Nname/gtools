package docker

import (
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestActions_Ping(t *testing.T) {
	actions := Actions{
		Options: []client.Opt{
			// client.WithHost("tcp://127.0.0.1:2375"),
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		},
	}
	_, err := actions.Ping()
	assert.NoError(t, err, "should not return an error")
}
