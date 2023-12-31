package versions

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tidwall/gjson"
)

type Versions struct {
	HardwareVersion string `json:"hardware_version"`
	SoftwareName    string `json:"software_name"`
	SoftwareVersion string `json:"software_version"`
	SoftwareBuildID string `json:"software_build_id"`

	SerialNumber   string `json:"serial_number"`
	UniqueDeviceID string `json:"unique_device_id,omitempty"`
	Hostname       string `json:"hostname"`
}

func (v *Versions) UserAgent() string {
	return fmt.Sprintf("[%s,%s,%s,%s]", v.SoftwareName, v.SoftwareVersion, v.SoftwareBuildID, v.HardwareVersion)
}

func getSoftwareName() string {
	softwareName, err := exec.Command("sw_vers", "-productName").Output()
	if err != nil {
		panic(fmt.Errorf("error running sw_vers: %w", err))
	}
	return strings.TrimSpace(string(softwareName))
}

func getSerialNumber() (serial, uuid string) {
	data, err := exec.Command("system_profiler", "SPHardwareDataType", "-json").Output()
	if err != nil {
		xmlData, err := exec.Command("system_profiler", "SPHardwareDataType", "-xml").Output()
		if err != nil {
			panic(fmt.Errorf("error running system_profiler: %w", err))
		}
		var result struct {
			Data struct {
				SPHardwareDataType []struct {
					SerialNumber string `xml:"serial_number"`
					PlatformUUID string `xml:"platform_UUID"`
				} `xml:"array>item"`
			} `xml:"array>dict"`
		}
		return result.Data.SPHardwareDataType[0].SerialNumber, result.Data.SPHardwareDataType[0].PlatformUUID
	}
	return gjson.GetBytes(data, "SPHardwareDataType.0.serial_number").Str,
		gjson.GetBytes(data, "SPHardwareDataType.0.platform_UUID").Str
}

func getHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func Get() Versions {
	// Alternative methods:
	// Hardware version: `system_profiler SPHardwareDataType | awk '/Model Identifier/ { print $3 }'`
	// Software version: `sw_vers -productVersion`
	// Software build ID: `sw_vers -buildVersion`
	// Serial number: `ioreg -c IOPlatformExpertDevice -d 2 | awk -F\" '/IOPlatformSerialNumber/{print $(NF-1)}'`
	output, err := exec.Command("sysctl", "-n", "hw.model", "kern.osversion", "kern.osproductversion").Output()
	if err != nil {
		panic(fmt.Errorf("error running sysctl: %w", err))
	}
	outParts := bytes.Split(output, []byte("\n"))
	if len(outParts) != 4 || len(outParts[3]) != 0 {
		panic(fmt.Errorf("unexpected output from sysctl: %q", string(output)))
	}
	serialNumber, deviceUUID := getSerialNumber()
	return Versions{
		HardwareVersion: string(outParts[0]),
		SoftwareName:    getSoftwareName(),
		SoftwareVersion: string(outParts[2]),
		SoftwareBuildID: string(outParts[1]),

		SerialNumber:   serialNumber,
		UniqueDeviceID: deviceUUID,
		Hostname:       getHostname(),
	}
}

var Current = Get()
