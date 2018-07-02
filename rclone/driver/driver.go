package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/spf13/viper"
)

var (
	//MountTimeout timeout before killing a mount try in seconds
	MountTimeout = 30
	//CfgVersion current config version compat
	CfgVersion = 1
	//CfgFolder config folder
	CfgFolder = "/etc/docker-volumes/rclone/"
)

type rcloneMountpoint struct {
	Path        string `json:"path"`
	Connections int    `json:"connections"`
}

type rcloneVolume struct {
	Config      string `json:"config"`
	Args        string `json:"args"`
	Remote      string `json:"remote"`
	Mount       string `json:"mount"`
	Connections int    `json:"connections"`
}

//RcloneDriver the global driver responding to call
type RcloneDriver struct {
	sync.RWMutex
	root       string
	persitence *viper.Viper
	volumes    map[string]*rcloneVolume
	mounts     map[string]*rcloneMountpoint
}

//Init start all needed deps and serve response to API call
func Init(root string) *RcloneDriver {
	d := &RcloneDriver{
		root:       root,
		persitence: viper.New(),
		volumes:    make(map[string]*rcloneVolume),
		mounts:     make(map[string]*rcloneMountpoint),
	}

	d.persitence.SetDefault("volumes", map[string]*rcloneVolume{})
	d.persitence.SetConfigName("persistence")
	d.persitence.SetConfigType("json")
	d.persitence.AddConfigPath(CfgFolder)
	if err := d.persitence.ReadInConfig(); err != nil { // Handle errors reading the config file
		log.Warn("No persistence file found, I will start with a empty list of volume. ", err)
	} else {
		log.Debug("Retrieving volume list from persistence file.")

		var version int
		err := d.persitence.UnmarshalKey("version", &version)
		if err != nil || version != CfgVersion {
			log.Warn("Unable to decode version of persistence, %s", err.Error())
			d.volumes = make(map[string]*rcloneVolume)
			d.mounts = make(map[string]*rcloneMountpoint)
		} else { //We have the same version
			err := d.persitence.UnmarshalKey("volumes", &d.volumes)
			if err != nil {
				log.Warn("Unable to decode into struct -> start with empty list, %s", err.Error())
				d.volumes = make(map[string]*rcloneVolume)
			}
			err = d.persitence.UnmarshalKey("mounts", &d.mounts)
			if err != nil {
				log.Warn("Unable to decode into struct -> start with empty list, %s", err.Error())
				d.mounts = make(map[string]*rcloneMountpoint)
			}
		}
	}
	return d
}

//Create create and init the requested volume
func (d *RcloneDriver) Create(r *volume.CreateRequest) error {
	log.Debugf("Entering Create: name: %s, options %v", r.Name, r.Options)
	d.Lock()
	defer d.Unlock()

	if r.Options == nil || r.Options["config"] == "" || r.Options["remote"] == "" {
		return fmt.Errorf("config and remote option required")
	}

	v := &rcloneVolume{
		Config:      r.Options["config"],
		Remote:      r.Options["remote"],
		Args:        r.Options["args"],
		Mount:       getMountName(d, r),
		Connections: 0,
	}

	if _, ok := d.mounts[v.Mount]; !ok { //This mountpoint doesn't allready exist -> create it
		m := &rcloneMountpoint{
			Path:        filepath.Join(d.root, v.Mount),
			Connections: 0,
		}

		_, err := os.Lstat(m.Path) //Create folder if not exist. This will also failed if already exist
		if os.IsNotExist(err) {
			if err = os.MkdirAll(m.Path, 0700); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		isempty, err := isEmpty(m.Path)
		if err != nil {
			return err
		}
		if !isempty {
			return fmt.Errorf("%v already exist and is not empty", m.Path)
		}
		d.mounts[v.Mount] = m
	}

	d.volumes[r.Name] = v
	log.Debugf("Volume Created: %v", v)
	return d.saveConfig()
}

//List volumes handled by the driver
func (d *RcloneDriver) List() (*volume.ListResponse, error) {
	log.Debugf("Entering List")
	d.Lock()
	defer d.Unlock()

	var vols []*volume.Volume
	for name, v := range d.volumes {
		log.Debugf("Volume found: %s", v)
		m, ok := d.mounts[v.Mount]
		if !ok {
			return nil, fmt.Errorf("volume mount %s not found for %s", v.Mount, v.Remote)
		}
		log.Debugf("Mount found: %v", m)
		vols = append(vols, &volume.Volume{Name: name, Mountpoint: m.Path})
	}
	return &volume.ListResponse{Volumes: vols}, nil
}

//Get get info on the requested volume
func (d *RcloneDriver) Get(r *volume.GetRequest) (*volume.GetResponse, error) {
	log.Debugf("Entering Get: name: %s", r.Name)
	d.Lock()
	defer d.Unlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return nil, fmt.Errorf("volume %s not found", r.Name)
	}
	log.Debugf("Volume found: %v", v)

	m, ok := d.mounts[v.Mount]
	if !ok {
		return nil, fmt.Errorf("volume mount %s not found for %s", v.Mount, r.Name)
	}
	log.Debugf("Mount found: %s", m)

	return &volume.GetResponse{Volume: &volume.Volume{Name: r.Name, Mountpoint: m.Path}}, nil
}

//Remove remove the requested volume
func (d *RcloneDriver) Remove(r *volume.RemoveRequest) error {
	//TODO remove related mounts
	//TODO Error response from daemon: unable to remove volume: remove hubic-crypt: VolumeDriver.Remove: volume hubic-crypt is currently used by a container

	log.Debugf("Entering Remove: name: %s", r.Name)
	d.Lock()
	defer d.Unlock()
	v, ok := d.volumes[r.Name]
	if !ok {
		return fmt.Errorf("volume %s not found", r.Name)
	}
	log.Debugf("Volume found: %s", v)

	m, ok := d.mounts[v.Mount]
	if !ok {
		return fmt.Errorf("volume mount %s not found for %s", v.Mount, r.Name)
	}
	log.Debugf("Mount found: %s", m)

	//disable check as it seems to fail and in this plugin v.Mount = r.Name
	//if v.Connections == 0 {
	//	if m.Connections == 0 {
	//Unmount
	if err := d.runCmd(fmt.Sprintf("umount \"%s\"", m.Path)); err != nil {
		return err
	}
	//Remove mount point
	if err := os.Remove(m.Path); err != nil {
		return err
	}
	delete(d.mounts, v.Mount)
	//}
	delete(d.volumes, r.Name)
	return d.saveConfig()
	//}
	/*
		if err := d.saveConfig(); err != nil {
			return err
		}
		return fmt.Errorf("volume %s is currently used by a container", r.Name)
	*/
}

//Path get path of the requested volume
func (d *RcloneDriver) Path(r *volume.PathRequest) (*volume.PathResponse, error) {
	log.Debugf("Entering Path: name: %s, options %v", r.Name)
	d.RLock()
	defer d.RUnlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return nil, fmt.Errorf("volume %s not found", r.Name)
	}
	log.Debugf("Volume found: %s", v)

	m, ok := d.mounts[v.Mount]
	if !ok {
		return nil, fmt.Errorf("volume mount %s not found for %s", v.Mount, r.Name)
	}
	log.Debugf("Mount found: %s", m)

	return &volume.PathResponse{Mountpoint: m.Path}, nil
}

//Mount mount the requested volume
func (d *RcloneDriver) Mount(r *volume.MountRequest) (*volume.MountResponse, error) {
	log.Debugf("Entering Mount: %v", r)
	d.Lock()
	defer d.Unlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return nil, fmt.Errorf("volume %s not found", r.Name)
	}

	m, ok := d.mounts[v.Mount]
	if !ok {
		return nil, fmt.Errorf("volume mount %s not found for %s", v.Mount, r.Name)
	}

	if m.Connections > 0 {
		v.Connections++
		m.Connections++
		if err := d.saveConfig(); err != nil {
			return nil, err
		}
		return &volume.MountResponse{Mountpoint: m.Path}, nil
	}

	//TODO write temp file before dans don't use base64
	var cmd string
	if log.GetLevel() == log.DebugLevel {
		cmd = fmt.Sprintf("/usr/bin/rclone --log-file /var/log/rclone.log --config=<(echo \"%s\"| base64 -d) %s mount \"%s\" \"%s\" & sleep 5s", v.Config, v.Args, v.Remote, m.Path)
	} else {
		cmd = fmt.Sprintf("/usr/bin/rclone --config=<(echo \"%s\"| base64 -d) %s mount \"%s\" \"%s\" & sleep 5s", v.Config, v.Args, v.Remote, m.Path)
	}
	if err := d.runCmd(cmd); err != nil {
		return nil, err
	}

	/* TODO test more this before using it.
	cmdCheck := fmt.Sprintf("mount | grep %s > /dev/null", m.Path)
	folderMounted := false
	for !folderMounted {
		log.Debugf("Waiting for mount: %s", m.Path)
		time.Sleep(5 * time.Second)
		folderMounted = (nil == d.runCmd(cmdCheck))
	}
	*/
	//Temporary fix
	time.Sleep(15 * time.Second)

	v.Connections++
	m.Connections++
	if err := d.saveConfig(); err != nil {
		return nil, err
	}
	return &volume.MountResponse{Mountpoint: m.Path}, nil
}

//Unmount unmount the requested volume
func (d *RcloneDriver) Unmount(r *volume.UnmountRequest) error {
	log.Debugf("Entering Unmount: %v", r)
	d.Lock()
	defer d.Unlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return fmt.Errorf("volume %s not found", r.Name)
	}

	m, ok := d.mounts[v.Mount]
	if !ok {
		return fmt.Errorf("volume mount %s not found for %s", v.Mount, r.Name)
	}

	if m.Connections <= 1 {
		cmd := fmt.Sprintf("/usr/bin/umount %s", m.Path)
		if err := d.runCmd(cmd); err != nil {
			return err
		}
		m.Connections = 0
		v.Connections = 0
	} else {
		m.Connections--
		v.Connections--
	}

	if err := d.saveConfig(); err != nil {
		return err
	}
	return nil
}

//Capabilities Send capabilities of the local driver
func (d *RcloneDriver) Capabilities() *volume.CapabilitiesResponse {
	log.Debugf("Entering Capabilities")
	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{
			Scope: "local",
		},
	}
}
