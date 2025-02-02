package artifacts

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/SAP/cloud-mta/mta"

	"github.com/SAP/cloud-mta-build-tool/internal/archive"
	"github.com/SAP/cloud-mta-build-tool/internal/buildops"
	"github.com/SAP/cloud-mta-build-tool/internal/logs"
)

type mtadLoc struct {
	path string
}

func (loc *mtadLoc) GetMtadPath() string {
	return filepath.Join(loc.path, dir.Mtad)
}

func (loc *mtadLoc) GetMetaPath() string {
	return filepath.Clean(loc.path)
}

func (loc *mtadLoc) GetManifestPath() string {
	return ""
}

func (loc *mtadLoc) GetMtarDir(targetProvided bool) string {
	return ""
}

// ExecuteGenMtad - generates MTAD from MTA
func ExecuteGenMtad(source, target string, extensions []string, platform string, wdGetter func() (string, error)) error {
	logs.Logger.Info("generating the MTAD file...")
	loc, err := dir.Location(source, target, dir.Dev, extensions, wdGetter)
	if err != nil {
		return errors.Wrap(err, "generation of the MTAD file failed when initializing the location")
	}

	return executeGenMetaByLocation(loc, &mtadLoc{target}, platform, false)
}

func validatePlatform(platform string) (string, error) {
	result := strings.ToLower(platform)
	if result != "xsa" && result != "cf" && result != "neo" {
		return "", fmt.Errorf(invalidPlatformMsg, platform)
	}
	return result, nil
}

// genMtad generates an mtad.yaml file from a mta.yaml file and a platform configuration file.
func genMtad(mtaStr *mta.MTA, ep dir.ITargetArtifacts, deploymentDesc bool, platform string,
	marshal func(interface{}) (out []byte, err error)) error {

	if !deploymentDesc {
		// convert modules types according to platform
		err := ConvertTypes(*mtaStr, platform)
		if err != nil {
			return errors.Wrapf(err, genMTADTypeTypeCnvMsg, platform)
		}
	}

	err := adjustSchemaVersion(mtaStr)
	if err != nil {
		return errors.Wrap(err, genMetaMTADMsg)
	}

	// Create readable Yaml before writing to file
	mtad, err := marshal(mtaStr)
	if err != nil {
		return errors.Wrap(err, genMTADMarshMsg)
	}
	mtadPath := ep.GetMtadPath()
	// Write back the MTAD to the META-INF folder
	err = ioutil.WriteFile(mtadPath, mtad, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, genMTADWriteMsg)
	}
	return nil
}

// removeUndeployedModules - remove elements from MTA that are not relevant for MTAD
// Function is used in process of deployment artifacts preparation
// SupportedPlatforms of module's build parameters indicate if module has to be deployed
// if SupportedPlatforms property defined with empty list of properties
// module will not be packed, not listed in MTAD yaml and in manifest
func removeUndeployedModules(mtaStr *mta.MTA, platform string) {

	// remove modules with no platforms defined
	for doCleaning := true; doCleaning; {
		doCleaning = false
		for i, m := range mtaStr.Modules {
			if !buildops.PlatformDefined(m, platform) {
				// join slices before and after removed module
				mtaStr.Modules = mtaStr.Modules[:i+copy(mtaStr.Modules[i:], mtaStr.Modules[i+1:])]
				doCleaning = true
				break
			}
		}
	}

	//TODO move to configuration
	if platform == "neo" {
		if mtaStr.Parameters == nil {
			mtaStr.Parameters = make(map[string]interface{})
		}
		if mtaStr.Parameters["hcp-deployer-version"] == nil {
			mtaStr.Parameters["hcp-deployer-version"] = "1.1.0"
		}
	}
}

// if module has to be deployed we clean build parameters from module,
// as this section is not used in MTAD yaml
func removeBuildParamsFromMta(loc dir.ITargetPath, mtaStr *mta.MTA) error {
	for _, m := range mtaStr.Modules {
		// remove build parameters from modules with defined platforms
		m.BuildParams = map[string]interface{}{}
		err := adaptModulePath(loc, m)
		if err != nil {
			return errors.Wrapf(err, adaptationMsg, m.Name)
		}
	}
	return nil
}

func adaptModulePath(loc dir.ITargetPath, module *mta.Module) error {
	modulePath := filepath.Join(loc.GetTargetTmpDir(), module.Name)
	_, e := os.Stat(modulePath)
	if e != nil {
		return e
	}
	module.Path = module.Name
	return nil
}

func adjustSchemaVersion(mtaStr *mta.MTA) error {
	schemaVersionSlice := strings.Split(*mtaStr.SchemaVersion, ".")
	schemaVersion, err := strconv.Atoi(schemaVersionSlice[0])
	if err != nil {
		return err
	}
	if schemaVersion < 3 {
		*mtaStr.SchemaVersion = "3.1"
	}
	return nil
}
