package driver_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-plugins-helpers/volume"

	"github.com/sapk/docker-volume-rclone/rclone"
	"github.com/sapk/docker-volume-rclone/rclone/driver"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	d := driver.Init("/tmp/test-root")
	if d == nil {
		t.Error("Expected to be not null, got ", d)
	}
	/*
		  if _, err := os.Stat(cfgFolder + "gluster-persistence.json"); err != nil {
				t.Error("Expected file to exist, got ", err)
			}
	*/
}

func TestMountName(t *testing.T) {
	name := driver.GetMountName(&driver.RcloneDriver{}, &volume.CreateRequest{
		Name: "test",
		Options: map[string]string{
			"remote": "some-remote:bucket/",
		},
	})

	if name != "test" {
		t.Error("Expected to be test, got ", name)
	}
}

//Inspired from https://github.com/docker/go-plugins-helpers/blob/master/volume/api_test.go
const (
	createPath       = "/VolumeDriver.Create"
	getPath          = "/VolumeDriver.Get"
	listPath         = "/VolumeDriver.List"
	removePath       = "/VolumeDriver.Remove"
	hostVirtualPath  = "/VolumeDriver.Path"
	mountPath        = "/VolumeDriver.Mount"
	unmountPath      = "/VolumeDriver.Unmount"
	capabilitiesPath = "/VolumeDriver.Capabilities"
)

func TestHandler(t *testing.T) {
	//Setup
	driver.CfgFolder = filepath.Join(t.TempDir(), "config")
	volumePath := filepath.Join(t.TempDir(), "volume")
	dataPath := filepath.Join(t.TempDir(), "data")
	assert.NoError(t, os.MkdirAll(dataPath, 0700))
	testFilePath := filepath.Join(dataPath, "test.file")
	testData := make([]byte, 42)
	rand.Read(testData)
	ioutil.WriteFile(testFilePath, testData, 0666)

	//Start in-memory handler
	d := driver.Init(filepath.Join(volumePath, rclone.PluginAlias))
	h := volume.NewHandler(d)
	l := sockets.NewInmemSocket("test", 0)
	go h.Serve(l)
	defer l.Close()

	client := &http.Client{Transport: &http.Transport{
		Dial: l.Dial,
	}}

	// Create No Option
	resp, err := pluginRequest(client, createPath, &volume.CreateRequest{Name: "foo"})
	assert.NoError(t, err)
	var vResp volume.ErrorResponse
	assert.NoError(t, json.NewDecoder(resp).Decode(&vResp))
	assert.Equal(t, "config and remote option required", vResp.Err)

	// Create No Remote
	resp, err = pluginRequest(client, createPath, &volume.CreateRequest{Name: "foo", Options: map[string]string{
		"config": "TODO",
	}})
	assert.NoError(t, err)
	assert.NoError(t, json.NewDecoder(resp).Decode(&vResp))
	assert.Equal(t, "config and remote option required", vResp.Err)
	// Create No Config
	resp, err = pluginRequest(client, createPath, &volume.CreateRequest{Name: "foo", Options: map[string]string{
		"remote": "TODO",
	}})
	assert.NoError(t, err)
	assert.NoError(t, json.NewDecoder(resp).Decode(&vResp))
	assert.Equal(t, "config and remote option required", vResp.Err)

	// Create
	resp, err = pluginRequest(client, createPath, &volume.CreateRequest{Name: "foo", Options: map[string]string{
		"config": "W3Rlc3RpbmddCnR5cGUgPSBsb2NhbAoK",
		"remote": "testing:" + dataPath,
		"args":   "",
	}})
	assert.NoError(t, err)
	assert.NoError(t, json.NewDecoder(resp).Decode(&vResp))
	assert.Equal(t, "config and remote option required", vResp.Err)

	//TODO test args

	// Get
	resp, err = pluginRequest(client, getPath, &volume.GetRequest{Name: "foo"})
	assert.NoError(t, err)
	var gResp *volume.GetResponse
	assert.NoError(t, json.NewDecoder(resp).Decode(&gResp))
	assert.Equal(t, "foo", gResp.Volume.Name)

	// List
	resp, err = pluginRequest(client, listPath, nil)
	assert.NoError(t, err)
	var lResp *volume.ListResponse
	assert.NoError(t, json.NewDecoder(resp).Decode(&lResp))
	assert.Equal(t, 1, len(lResp.Volumes))
	assert.Equal(t, "foo", lResp.Volumes[0].Name)

	// Path
	resp, err = pluginRequest(client, hostVirtualPath, &volume.PathRequest{Name: "foo"})
	assert.NoError(t, err)
	var pResp *volume.PathResponse
	assert.NoError(t, json.NewDecoder(resp).Decode(&pResp))
	assert.Equal(t, filepath.Join(volumePath, "rclone", "foo"), pResp.Mountpoint)

	// Mount
	resp, err = pluginRequest(client, mountPath, &volume.MountRequest{Name: "foo"})
	assert.NoError(t, err)
	var mResp *volume.PathResponse
	assert.NoError(t, json.NewDecoder(resp).Decode(&mResp))
	assert.Equal(t, filepath.Join(volumePath, "rclone", "foo"), mResp.Mountpoint)

	//Check content
	filePathInVol := filepath.Join(mResp.Mountpoint, "test.file")
	dataDetected, err := ioutil.ReadFile(filePathInVol)
	assert.NoError(t, err)
	assert.Equal(t, testData, dataDetected)

	// Unmount
	resp, err = pluginRequest(client, unmountPath, &volume.UnmountRequest{Name: "foo"})
	assert.NoError(t, err)
	var uResp volume.ErrorResponse
	assert.NoError(t, json.NewDecoder(resp).Decode(&uResp))
	assert.Equal(t, "", uResp.Err)

	// Remove
	resp, err = pluginRequest(client, removePath, &volume.RemoveRequest{Name: "foo"})
	assert.NoError(t, err)
	var rmResp volume.ErrorResponse
	assert.NoError(t, json.NewDecoder(resp).Decode(&rmResp))
	assert.Equal(t, "", rmResp.Err)
	//Re-List
	resp, err = pluginRequest(client, listPath, nil)
	assert.NoError(t, err)
	assert.NoError(t, json.NewDecoder(resp).Decode(&lResp))
	assert.Equal(t, 0, len(lResp.Volumes))

	// Capabilities
	resp, err = pluginRequest(client, capabilitiesPath, nil)
	assert.NoError(t, err)
	var cResp *volume.CapabilitiesResponse
	assert.NoError(t, json.NewDecoder(resp).Decode(&cResp))
	assert.Equal(t, "local", cResp.Capabilities.Scope)
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
