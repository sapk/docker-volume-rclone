package rclone

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/sapk/docker-volume-rclone/rclone/driver"
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
		log.Fatal().Err(err)
	}
}

//DaemonStart Start the deamon
func DaemonStart(cmd *cobra.Command, args []string) {
	d := driver.Init(baseDir)
	log.Debug().Msgf("driver: %v", d)
	h := volume.NewHandler(d)
	log.Debug().Msgf("handler: %v", h)
	err := h.ServeUnix(PluginAlias, 0)
	if err != nil {
		log.Debug().Err(err)
	}
}

func setupFlags() {
	rootCmd.PersistentFlags().BoolP(VerboseFlag, "v", os.Getenv("DEBUG") == "1", "Turns on verbose logging")
	rootCmd.PersistentFlags().StringVarP(&baseDir, BasedirFlag, "b", filepath.Join(volume.DefaultDockerRootDirectory, PluginAlias), "Mounted volume base directory")
}

func setupLogger(cmd *cobra.Command, args []string) {
	if verbose, _ := cmd.Flags().GetBool(VerboseFlag); verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		//Activate log to file in debug mode
		f, err := os.OpenFile("/var/log/docker-volume-rclone.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal().Err(err)
		}
		//log.SetOutput(f)
		//logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
		log.Logger = zerolog.New(f).With().Timestamp().Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
