package main

import (
	"github.com/sapk/docker-volume-rclone/rclone"
)

var (
	//Version version of app set by build flag
	Version string
	//Branch git branch of app set by build flag
	Branch string
	//Commit git commit of app set by build flag
	Commit string
	//BuildTime build time of app set by build flag
	BuildTime string
)

func main() {
	rclone.Version = Version
	rclone.Commit = Commit
	rclone.Branch = Branch
	rclone.BuildTime = BuildTime
	rclone.Start()
}
