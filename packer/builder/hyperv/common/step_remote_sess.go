package common

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/MSOpenTech/packer-hyperv/packer/communicator/powershell"
)

type StepRemoteSession struct {
	comm packer.Communicator
}

func (s *StepRemoteSession) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error StepRemoteSession: %s"
	vmName := state.Get("vmName").(string)
	ip := state.Get("ip").(string)

	ui.Say("Adding to TrustedHosts (require elevated mode)")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock { Set-Item -path WSMan:\\localhost\\Client\\TrustedHosts '")
	blockBuffer.WriteString(ip)
	blockBuffer.WriteString("' -Force }")

	var err error
	err = driver.HypervManage(blockBuffer.String())

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	comm, err := powershell.New(
		&powershell.Config{
			Username: "vagrant",
			Password: "vagrant",
			RemoteHostIP: ip,
			VmName: vmName,
			Ui: ui,
		})

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.comm = comm
	state.Put("communicator", comm)

	return multistep.ActionContinue
}

func (s *StepRemoteSession) Cleanup(state multistep.StateBag) {
	// do nothing
}
