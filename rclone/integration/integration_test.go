package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os/user"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-plugins-helpers/volume"

	"github.com/sapk/docker-volume-rclone/rclone"

	"github.com/stretchr/testify/assert"
)

//Inspired from https://github.com/docker/go-plugins-helpers/blob/master/volume/api_test.go
const (
	manifest         = `{"Implements": ["VolumeDriver"]}`
	createPath       = "/VolumeDriver.Create"
	getPath          = "/VolumeDriver.Get"
	listPath         = "/VolumeDriver.List"
	removePath       = "/VolumeDriver.Remove"
	hostVirtualPath  = "/VolumeDriver.Path"
	mountPath        = "/VolumeDriver.Mount"
	unmountPath      = "/VolumeDriver.Unmount"
	capabilitiesPath = "/VolumeDriver.Capabilities"
)

func startDaemon(t *testing.T) {
	//Launch
	cmd := rclone.NewRootCmd()
	cmd.SetArgs([]string{"daemon"})
	go cmd.Execute()

	time.Sleep(10 * time.Millisecond)
}

func TestIntergation(t *testing.T) {
	//Start one at a time
	t.Run("cmd/version", testCmdVersion)
	t.Run("cmd/daemon", testCmdDeamon)
}

func testCmdDeamon(t *testing.T) {
	u, err := user.Current()
	assert.NoError(t, err)
	if u.Uid != "0" {
		t.Skipf("Skipping daemon tests since you are not root")
	}

	startDaemon(t)

	t.Run("Capabilities", testCapabilities)
	//TODO add more

	time.Sleep(10 * time.Millisecond)
}

func testCapabilities(t *testing.T) {
	dial, err := net.Dial("unix", filepath.Join("/run/docker/plugins", "rclone.sock"))
	assert.NoError(t, err)

	client := &http.Client{Transport: &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return dial, nil
		},
	}}

	// Capabilities
	resp, err := pluginRequest(client, capabilitiesPath, nil)
	assert.NoError(t, err)
	var cResp *volume.CapabilitiesResponse
	assert.NoError(t, json.NewDecoder(resp).Decode(&cResp))
	assert.Equal(t, "local", cResp.Capabilities.Scope)
}

func testCmdVersion(t *testing.T) {
	rclone.Version = "TESTING"

	cmd := rclone.NewRootCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"version"})
	cmd.Execute()
	out, err := ioutil.ReadAll(b)
	assert.NoError(t, err)
	assert.Equal(t, "\nVersion: TESTING - Branch:  - Commit:  - BuildTime: \n\n", string(out), "The version returned by CLI is invalid")
}

func pluginRequest(client *http.Client, method string, req interface{}) (io.Reader, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if req == nil {
		b = []byte{}
	}
	resp, err := client.Post("http://localhost"+method, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
