package commands

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/SAP/cloud-mta-build-tool/internal/artifacts"
	"github.com/SAP/cloud-mta-build-tool/internal/exec"
	"github.com/SAP/cloud-mta-build-tool/internal/tpl"
)

const (
	makefile = "Makefile.mta"
)

// flags of init command
var initCmdSrc string
var initCmdTrg string
var initCmdExtensions []string
var initCmdMode string

// flags of build command
var buildProjectCmdSrc string
var buildProjectCmdTrg string
var buildProjectCmdExtensions []string
var buildProjectCmdMtar = "*"
var buildProjectCmdPlatform string
var buildProjectCmdStrict bool

func init() {
	// set flags for init command
	initCmd.Flags().StringVarP(&initCmdSrc, "source", "s", "", "The path to the MTA project; the current path is set as the default")
	initCmd.Flags().StringVarP(&initCmdTrg, "target", "t", "", "The path to the generated Makefile folder; the current path is set as the default")
	initCmd.Flags().StringSliceVarP(&initCmdExtensions, "extensions", "e", nil, "The MTA extension descriptors")
	initCmd.Flags().StringVarP(&initCmdMode, "mode", "m", "", `The mode of the Makefile generation; supported values: "default" and "verbose"`)
	_ = initCmd.Flags().MarkHidden("mode")
	initCmd.Flags().BoolP("help", "h", false, `Displays detailed information about the "init" command`)

	// set flags of build command
	buildCmd.Flags().StringVarP(&buildProjectCmdSrc, "source", "s", "", "The path to the MTA project; the current path is set as the default")
	buildCmd.Flags().StringVarP(&buildProjectCmdTrg, "target", "t", "", "The path to the results folder; the current path is set as the default")
	buildCmd.Flags().StringSliceVarP(&buildProjectCmdExtensions, "extensions", "e", nil, "The MTA extension descriptors")
	buildCmd.Flags().StringVarP(&buildProjectCmdMtar, "mtar", "", "", "The file name of the generated archive file")
	buildCmd.Flags().StringVarP(&buildProjectCmdPlatform, "platform", "p", "cf", `The deployment platform; supported platforms: "cf", "xsa", "neo"`)
	buildCmd.Flags().BoolVarP(&buildProjectCmdStrict, "strict", "", true, `If set to true, duplicated fields and fields not defined in the "mta.yaml" schema are reported as errors; if set to false, they are reported as warnings`)
	buildCmd.Flags().BoolP("help", "h", false, `Displays detailed information about the "build" command`)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generates a GNU Make manifest file that describes the build process of the MTA project",
	Long:  "Generates a GNU Make manifest file that describes the build process of the MTA project",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Generate build script
		err := tpl.ExecuteMake(initCmdSrc, initCmdTrg, initCmdExtensions, makefile, initCmdMode, os.Getwd, true)
		logError(err)
	},
}

// Execute MTA project build
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the project modules and generates an MTA archive according to the MTA development descriptor (mta.yaml)",
	Long:  "Builds the project modules and generates an MTA archive according to the MTA development descriptor (mta.yaml)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Generate temp Makefile with unique id
		makefileTmp := "Makefile_" + time.Now().Format("20060102150405") + ".mta"
		// Generate build script
		// Note: we can only use the non-default mbt (i.e. the current executable name) from inside the command itself because if this function runs from other places like tests it won't point to the MBT
		err := artifacts.ExecBuild(makefileTmp, buildProjectCmdSrc, buildProjectCmdTrg, buildProjectCmdExtensions, "", buildProjectCmdMtar, buildProjectCmdPlatform, buildProjectCmdStrict, os.Getwd, exec.Execute, false)
		return err
	},
	SilenceUsage: true,
}
