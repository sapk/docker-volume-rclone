package driver

import (
	"testing"

	"github.com/docker/go-plugins-helpers/volume"
)

func TestInit(t *testing.T) {
	d := Init("/tmp/test-root", "", false)
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
	name := getMountName(&RcloneDriver{}, volume.Request{
		Name: "test",
		Options: map[string]string{
			"remote": "some-remote:bucket/",
		},
	})

	if name != "test" {
		t.Error("Expected to be test, got ", name)
	}
}
