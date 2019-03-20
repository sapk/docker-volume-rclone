package driver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/rs/zerolog/log"
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
		log.Warn().Err(err).Msg("Unable to encode persistence struct")
	}
	//log.Debug("Writing persistence struct, %v", b, d.volumes)
	err = ioutil.WriteFile(CfgFolder+"/persistence.json", b, 0600)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to write persistence struct, %s")
	}
	//TODO display error messages
	return err
}

// run deamon in context of this gvfs drive with custome env
func (d *RcloneDriver) runCmd(cmd string) error {
	log.Debug().Msg(cmd)
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
