package versions

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Versions struct {
	HardwareVersion string `json:"hardware_version"`
	SoftwareName    string `json:"software_name"`
	SoftwareVersion string `json:"software_version"`
	SoftwareBuildID string `json:"software_build_id"`
}

func (v *Versions) UserAgent() string {
	return fmt.Sprintf("[%s,%s,%s,%s]", v.SoftwareName, v.SoftwareVersion, v.SoftwareBuildID, v.HardwareVersion)
}

func Get() Versions {
	// Alternative methods:
	// Hardware version: `system_profiler SPHardwareDataType | awk '/Model Identifier/ { print $3 }'`
	// Software version: `sw_vers -productVersion`
	// Software build ID: `sw_vers -buildVersion`
	output, err := exec.Command("sysctl", "-n", "hw.model", "kern.osversion", "kern.osproductversion").Output()
	if err != nil {
		panic(fmt.Errorf("error running sysctl: %w", err))
	}
	softwareName, err := exec.Command("sw_vers", "-productName").Output()
	if err != nil {
		panic(fmt.Errorf("error running sw_vers: %w", err))
	}
	outParts := bytes.Split(output, []byte("\n"))
	if len(outParts) != 4 || len(outParts[3]) != 0 {
		panic(fmt.Errorf("unexpected output from sysctl: %q", string(output)))
	}
	return Versions{
		HardwareVersion: string(outParts[0]),
		SoftwareName:    strings.TrimSpace(string(softwareName)),
		SoftwareVersion: string(outParts[1]),
		SoftwareBuildID: string(outParts[2]),
	}
}

var Current = Get()
