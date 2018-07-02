package driver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/sirupsen/logrus"
)

//RclonePersistence represent struct of persistence file
type RclonePersistence struct {
	Version int                          `json:"version"`
	Volumes map[string]*rcloneVolume     `json:"volumes"`
	Mounts  map[string]*rcloneMountpoint `json:"mounts"`
}

func (d *RcloneDriver) saveConfig() error {
	fi, err := os.Lstat(CfgFolder)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(CfgFolder, 0700); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if fi != nil && !fi.IsDir() {
		return fmt.Errorf("%v already exist and it's not a directory", d.root)
	}
	b, err := json.Marshal(RclonePersistence{Version: CfgVersion, Volumes: d.volumes, Mounts: d.mounts})
	if err != nil {
		log.Warn("Unable to encode persistence struct, %s", err.Error())
	}
	//log.Debug("Writing persistence struct, %v", b, d.volumes)
	err = ioutil.WriteFile(CfgFolder+"/persistence.json", b, 0600)
	if err != nil {
		log.Warn("Unable to write persistence struct, %s", err.Error())
	}
	//TODO display error messages
	return err
}

// run deamon in context of this gvfs drive with custome env
func (d *RcloneDriver) runCmd(cmd string) error {
	log.Debugf(cmd)
	/*
		cli := exec.Command("/bin/bash", "-c", cmd)
		stdoutStderr, err := cli.CombinedOutput()
		log.Debugf("%s", stdoutStderr)
		return err
	*/
	return exec.Command("/bin/bash", "-c", cmd).Run()
	//TODO output log
}

func getMountName(d *RcloneDriver, r *volume.CreateRequest) string {
	return r.Name
}

//based on: http://stackoverflow.com/questions/30697324/how-to-check-if-directory-on-path-is-empty
func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer func() {
		cerr := f.Close()
		if err == nil && cerr != nil {
			err = cerr
		}
	}()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
