package iso

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mitchellh/multistep"
	hypervcommon "github.com/MSOpenTech/packer-hyperv/packer/builder/hyperv/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
//	"debug/elf"
	"io/ioutil"
	"regexp"
)

const BuilderId = "MSOpenTech.hyperv.iso"

// EVALUATION EDITIONS
const (
	WS2012R2DC string 	= "WindowsServer2012R2Datacenter"
	PRODUCT_DATACENTER_EVALUATION_SERVER int64 = 80
	PRODUCT_DATACENTER_SERVER int64 = 8
)

// Builder implements packer.Builder and builds the actual Hyperv
// images.
type Builder struct {
	config iso_config
	runner multistep.Runner
}

type iso_config struct {
	DiskSizeGB            uint     			`mapstructure:"disk_size_gb"`
	RamSizeMB             uint     			`mapstructure:"ram_size_mb"`
	GuestOSType         string   			`mapstructure:"guest_os_type"`
	VMName              string   			`mapstructure:"vm_name"`
	SwitchName          string  			`mapstructure:"switch_name"`
	RawSingleISOUrl 	string 				`mapstructure:"iso_url"`
	SleepTimeMinutes 	time.Duration		`mapstructure:"wait_time_minutes"`
	ProductKey 			string				`mapstructure:"product_key"`

	common.PackerConfig           			`mapstructure:",squash"`
	hypervcommon.OutputConfig     			`mapstructure:",squash"`

	tpl *packer.ConfigTemplate
}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {

	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("%s: %v", "PackerUserVars", b.config.PackerUserVars))

	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors and warnings
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.OutputConfig.Prepare(b.config.tpl, &b.config.PackerConfig)...)
	warnings := make([]string, 0)

	if b.config.DiskSizeGB == 0 {
		b.config.DiskSizeGB = 40
	}
	log.Println(fmt.Sprintf("%s: %v", "DiskSize", b.config.DiskSizeGB))

	if(b.config.DiskSizeGB < 10 ){
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Windows server requires disk space no less than 10GB, but defined: %v", b.config.DiskSizeGB))
	}

	if b.config.RamSizeMB == 0 {
		b.config.RamSizeMB = 1024
	}

	log.Println(fmt.Sprintf("%s: %v", "RamSize", b.config.RamSizeMB))

	if(b.config.RamSizeMB < 512 ){
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Windows server requires memory size no less than 512MB, but defined: %v", b.config.RamSizeMB))
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}

	if b.config.SwitchName == "" {
		b.config.SwitchName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}

	if b.config.SleepTimeMinutes == 0 {
		b.config.SleepTimeMinutes = 10
	}
	log.Println(fmt.Sprintf("%s: %v", "SleepTimeMinutes", uint(b.config.SleepTimeMinutes)))

	if len(b.config.ProductKey) != 0 {
		pattern := "^[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}$"
		value := b.config.ProductKey

		match, _ := regexp.MatchString(pattern, value)
		if !match {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Make sure the product_key follows the pattern: XXXXX-XXXXX-XXXXX-XXXXX-XXXXX"))
		}
	}

	// Errors
	templates := map[string]*string{
		"vm_name":                &b.config.VMName,
		"switch_name":            &b.config.SwitchName,
		"product_key":            &b.config.ProductKey,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	log.Println(fmt.Sprintf("%s: %v","VMName", b.config.VMName))
	log.Println(fmt.Sprintf("%s: %v","SwitchName", b.config.SwitchName))
	log.Println(fmt.Sprintf("%s: %v","ProductKey", b.config.ProductKey))


	log.Println(fmt.Sprintf("%s: %v","RawSingleISOUrl", b.config.RawSingleISOUrl))

	if b.config.RawSingleISOUrl == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("iso_url must be specified."))
	}else if _, err := os.Stat(b.config.RawSingleISOUrl); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Check iso_url is correct"))
	}


	guestOSTypesIsValid := false
	guestOSTypes := []string{
		WS2012R2DC,
//		WS2012R2St,
	}

	log.Println(fmt.Sprintf("%s: %v","GuestOSType", b.config.GuestOSType))

	for _, guestType := range guestOSTypes {
		if b.config.GuestOSType == guestType {
			guestOSTypesIsValid = true
			break
		}
	}

	if !guestOSTypesIsValid {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("guest_os_type is invalid. Must be one of: %v", guestOSTypes))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

// Run executes a Packer build and returns a packer.Artifact representing
// a Hyperv appliance.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with Hyperv
	driver, err := hypervcommon.NewHypervPS4Driver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating Hyper-V driver: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	if len(b.config.ProductKey) > 0{
		ui.Say("Product key found in the configuration file. To complete Windows activation with the product key Packer will need an Internet connection.")
	}

	tempDir := os.TempDir()
	packerTempDir, err := ioutil.TempDir(tempDir, "packer")
	if err != nil {
		return nil, fmt.Errorf("Failed creating temporary directory: %s", err)
	}

	state.Put("packerTempDir", packerTempDir)
	log.Println("packerTempDir = .'" + packerTempDir + "'")

// TODO: remove me
//	state.Put("vmName", "PackerFull")

	steps := []multistep.Step{

		&hypervcommon.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},

		&hypervcommon.StepCreateSwitch{
			SwitchName: b.config.SwitchName,
		},
		new(StepCreateVM),
		new(hypervcommon.StepEnableIntegrationService),
		new(StepMountDvdDrive),
		new(StepMountFloppydrive),
//		new(hypervcommon.StepConfigureVlan),
		new(hypervcommon.StepStartVm),
		&hypervcommon.StepSleep{ Minutes: b.config.SleepTimeMinutes, ActionName: "Installing" },

		new(hypervcommon.StepConfigureIp),
//		new(hypervcommon.StepPollingInstalation),
//		&hypervcommon.StepSleep{ Minutes: 1, ActionName: "Booting" },
		new(hypervcommon.StepRemoteSession),
		new(common.StepProvision),
		new(StepInstallProductKey),

//		new(hypervcommon.StepStopVm),
	}

	// Run the steps.
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}
	b.runner.Run(state)

	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	return hypervcommon.NewArtifact(b.config.OutputDir)
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
