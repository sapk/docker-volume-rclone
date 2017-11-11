package rclone

import (
	"fmt"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/sapk/docker-volume-rclone/rclone/driver"
	"github.com/spf13/cobra"
)

const (
	//VerboseFlag flag to set more verbose level
	VerboseFlag = "verbose"
	//BasedirFlag flag to set the basedir of mounted volumes
	BasedirFlag = "basedir"
	longHelp    = `
docker-volume-rclone (Rclone Volume Driver Plugin)
Provides docker volume support for Rclone.
== Version: %s - Branch: %s - Commit: %s - BuildTime: %s ==
`
)

var (
	//Version version of running code
	Version string
	//Branch branch of running code
	Branch string
	//Commit commit of running code
	Commit string
	//BuildTime build time of running code
	BuildTime string
	//PluginAlias plugin alias name in docker
	PluginAlias = "rclone"
	baseDir     = ""
	rootCmd     = &cobra.Command{
		Use:              "docker-volume-rclone",
		Short:            "Rclone - Docker volume driver plugin",
		Long:             longHelp,
		PersistentPreRun: setupLogger,
	}
	daemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "Run listening volume drive deamon to listen for mount request",
		Run:   DaemonStart,
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display current version and build date",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("\nVersion: %s - Branch: %s - Commit: %s - BuildTime: %s\n\n", Version, Branch, Commit, BuildTime)
		},
	}
)

//Start start the program
func Start() {
	setupFlags()
	rootCmd.Long = fmt.Sprintf(longHelp, Version, Branch, Commit, BuildTime)
	rootCmd.AddCommand(versionCmd, daemonCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

//DaemonStart Start the deamon
func DaemonStart(cmd *cobra.Command, args []string) {
	d := driver.Init(baseDir)
	log.Debug(d)
	h := volume.NewHandler(d)
	log.Debug(h)
	err := h.ServeUnix(PluginAlias, 0)
	if err != nil {
		log.Debug(err)
	}
}

func setupFlags() {
	rootCmd.PersistentFlags().BoolP(VerboseFlag, "v", false, "Turns on verbose logging")
	rootCmd.PersistentFlags().StringVarP(&baseDir, BasedirFlag, "b", filepath.Join(volume.DefaultDockerRootDirectory, PluginAlias), "Mounted volume base directory")
}

func setupLogger(cmd *cobra.Command, args []string) {
	if verbose, _ := cmd.Flags().GetBool(VerboseFlag); verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
