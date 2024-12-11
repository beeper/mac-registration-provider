package find_offsets

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	FAT_MAGIC      = 0xcafebabe
	MACHO_MAGIC_32 = 0xfeedface
	MACHO_MAGIC_64 = 0xfeedfacf
)

type Architecture struct {
	Name             string
	CPUType          uint32
	CPUSubtype       uint32
	CPUSubtypeCaps   uint32
	Offset           uint32
	Size             uint32
	Align            uint32
	ValidMachOHeader bool
}

type HexStrings map[string]map[string]string

var (
	HexStringsModern = HexStrings{
		"x86_64": {
			"ReferenceAddress (_IDSProtoKeyTransparencyTrustedServiceReadFrom)": "554889e54157415641554154534883ec28..89..48897dd04c8b3d",
			"NACInitAddress":             "554889e541574156415541545350b87818",
			"NACKeyEstablishmentAddress": "554889e54157415641554154534881ec..010000488b05......00488b00488945d04885",
			"NACSignAddress":             "554889e54157415641554154534881ec..030000........................................................................................................................................................................................48....48..........................................................................................................89............................................................",
		},
		"arm64e": {
			"ReferenceAddress (_IDSProtoKeyTransparencyTrustedServiceReadFrom)": "7f2303d5ffc301d1fc6f01a9fa6702a9f85f03a9f65704a9f44f05a9fd7b06a9fd830191f30301aa....00........f9..0280b9..68..f8....00........f9....80b9..68..f8....00........f9..01..eb....0054f40300aa............................................................................................................................80b96d6a6df89f010deb....0054..0380b96d6a6df8................................................",
			"NACInitAddress":             "7f2303d5fc6fbaa9fa6701a9f85f02a9f65703a9f44f04a9fd7b05a9fd43019109..8352....00..10....f91f0a3fd6ff0740d1ff....d1....00..08....f9080140f9a8....f8......d2......f2......f2......f2e9",
			"NACKeyEstablishmentAddress": "7f2303d5ff....d1fc6f..a9fa67..a9f85f..a9f657..a9f44f..a9fd7b..a9fd..0591....00..08....f9080140f9a8....f8......52",
			"NACSignAddress":             "7f2303d5fc6fbaa9fa6701a9f85f02a9f65703a9f44f04a9fd7b05a9fd430191ff....d1................08....f9......................................................................................................................................f2......f2......................d2",
		},
	}
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test.go <path_to_binary>")
		return
	}

	filePath := os.Args[1]

	architectures, err := ScanMachOFATBinary(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	printArchInfo(architectures)

	symbol := "_IDSProtoKeyTransparencyTrustedServiceReadFrom"
	fmt.Println("\n-= Found Symbol Offsets =-")
	for _, arch := range architectures {
		offset := getSymbolOffset(filePath, symbol, arch.Name)
		if offset != "" {
			fmt.Printf("Offset of %s in architecture %s: %s\n", symbol, arch.Name, offset)
		} else {
			fmt.Printf("Symbol %s not found in architecture %s.\n", symbol, arch.Name)
		}
	}

	fmt.Println("")

	searchResults := SearchInArchitectures(filePath, architectures, HexStringsModern)
	printSearchResults(searchResults, architectures, " (with pure Go fixed sequence search + regex)")
}

func ScanMachOFATBinary(filePath string) (map[int]Architecture, error) {
	architectures := make(map[int]Architecture)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	magic := binary.BigEndian.Uint32(data[0:4])
	if magic != FAT_MAGIC {
		offset := uint32(0x0)
		isValidMachO := validateMachOHeader(data, offset)
		architectures[0] = Architecture{
			Name:             "x86_64",
			CPUType:          0x0,
			CPUSubtype:       0x0,
			CPUSubtypeCaps:   0x0,
			Offset:           offset,
			Size:             uint32(len(data)),
			Align:            0,
			ValidMachOHeader: isValidMachO,
		}
		return architectures, nil
	}

	numArchs := binary.BigEndian.Uint32(data[4:8])
	for i := 0; i < int(numArchs); i++ {
		archOffset := 8 + i*20
		archInfo := data[archOffset : archOffset+20]

		cpuType := binary.BigEndian.Uint32(archInfo[0:4])
		cpuSubtypeFull := binary.BigEndian.Uint32(archInfo[4:8])
		offset := binary.BigEndian.Uint32(archInfo[8:12])
		size := binary.BigEndian.Uint32(archInfo[12:16])
		align := binary.BigEndian.Uint32(archInfo[16:20])

		cpuSubtype := cpuSubtypeFull & 0x00FFFFFF
		cpuSubtypeCaps := (cpuSubtypeFull >> 24) & 0xFF

		archName := getArchName(cpuType, cpuSubtype, cpuSubtypeCaps)
		isValidMachO := validateMachOHeader(data, offset)

		architectures[i] = Architecture{
			Name:             archName,
			CPUType:          cpuType,
			CPUSubtype:       cpuSubtype,
			CPUSubtypeCaps:   cpuSubtypeCaps,
			Offset:           offset,
			Size:             size,
			Align:            align,
			ValidMachOHeader: isValidMachO,
		}
	}

	return architectures, nil
}

func getArchName(cpuType, cpuSubtype, cpuSubtypeCaps uint32) string {
	archNames := map[[3]uint32]string{
		{16777223, 3, 0}:   "x86_64",
		{16777228, 2, 128}: "arm64e",
	}
	key := [3]uint32{cpuType, cpuSubtype, cpuSubtypeCaps}
	if name, ok := archNames[key]; ok {
		return name
	}
	return fmt.Sprintf("Unknown (Type: %d, Subtype: %d, Subtype Capability: %d)", cpuType, cpuSubtype, cpuSubtypeCaps)
}

func validateMachOHeader(data []byte, offset uint32) bool {
	magic := binary.LittleEndian.Uint32(data[offset : offset+4])
	return magic == MACHO_MAGIC_32 || magic == MACHO_MAGIC_64
}

func printArchInfo(architectures map[int]Architecture) {
	fmt.Println("-= Universal Binary Sections =-")
	for i, arch := range architectures {
		fmt.Printf("Architecture %d (%s):\n", i, arch.Name)
		fmt.Printf("  CPU Type: %d (0x%x)\n", arch.CPUType, arch.CPUType)
		fmt.Printf("  CPU Subtype: %d (0x%x)\n", arch.CPUSubtype, arch.CPUSubtype)
		fmt.Printf("  CPU Subtype Capability: %d (0x%x)\n", arch.CPUSubtypeCaps, arch.CPUSubtypeCaps)
		fmt.Printf("  Offset: 0x%x (Valid Mach-O Header: %v)\n", arch.Offset, arch.ValidMachOHeader)
		fmt.Printf("  Size: %d\n", arch.Size)
		fmt.Printf("  Align: %d\n", arch.Align)
	}
}

func getSymbolOffset(binaryPath, symbol, arch string) string {
	archFlag := "--arch=x86_64"
	if arch == "arm64e" {
		archFlag = "--arch=arm64e"
	}
	cmd := exec.Command("/usr/bin/nm", "--defined-only", "--extern-only", archFlag, binaryPath)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error executing nm: %s\n", err)
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[2] == symbol {
			return fmt.Sprintf("0x%s", parts[0][len(parts[0])-6:])
		}
	}
	return ""
}

func SearchInArchitectures(filePath string, architectures map[int]Architecture, hexStrings HexStrings) map[int]map[string][]int {
	searchResults := make(map[int]map[string][]int)

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return searchResults
	}

	for i, arch := range architectures {
		if !arch.ValidMachOHeader {
			fmt.Printf("Warning: Skipping architecture %d (%s) due to invalid Mach-O header.\n", i, arch.Name)
			continue
		}

		binaryData := data[arch.Offset : arch.Offset+arch.Size]

		archHexStrings, ok := hexStrings[arch.Name]
		if !ok {
			continue
		}

		searchResults[i] = make(map[string][]int)
		for name, hexString := range archHexStrings {
			matches := searchWithFixedSequencesAndRegex(binaryData, hexString)
			if len(matches) > 0 {
				searchResults[i][name] = matches
			}
		}
	}
	return searchResults
}

func searchWithFixedSequencesAndRegex(binaryData []byte, hexString string) []int {
	var results []int

	hexString = strings.ReplaceAll(hexString, " ", "")
	hexString = strings.ToLower(hexString)

	longestFixed := findLongestFixedSequence(hexString)
	fixedBytes := hexStringToBytes(longestFixed)

	start := 0
	for start < len(binaryData) {
		index := bytes.Index(binaryData[start:], fixedBytes)
		if index == -1 {
			break
		}

		if matchHexPattern(binaryData[start+index:], hexString) {
			results = append(results, start+index)
		}
		start += index + len(fixedBytes) // Move start position past the found fixedBytes
	}
	return results
}

func findLongestFixedSequence(hexPattern string) string {
	wildcardIndex := strings.Index(hexPattern, "..")
	if wildcardIndex != -1 {
		return hexPattern[:wildcardIndex]
	}
	return hexPattern
}

func hexStringToBytes(hexString string) []byte {
	b, _ := hex.DecodeString(hexString)
	return b
}

func matchHexPattern(data []byte, hexPattern string) bool {
	for i := 0; i < len(hexPattern); i += 2 {
		bytePattern := hexPattern[i : i+2]
		if bytePattern == ".." {
			continue
		}
		byteValue, err := strconv.ParseUint(bytePattern, 16, 8)
		if err != nil {
			return false
		}
		if data[i/2] != byte(byteValue) {
			return false
		}
	}
	return true
}

func printSearchResults(searchResults map[int]map[string][]int, architectures map[int]Architecture, suffix string) {
	fmt.Printf("-= Found Hex Offsets%s =-\n", suffix)
	for archIndex, results := range searchResults {
		archName := architectures[archIndex].Name
		fmt.Printf("Architecture %d (%s):\n", archIndex, archName)
		for name, offsets := range results {
			offsetStrings := make([]string, len(offsets))
			for i, offset := range offsets {
				offsetStrings[i] = fmt.Sprintf("0x%x", offset)
			}
			fmt.Printf("  %s: %s\n", name, strings.Join(offsetStrings, "; "))
		}
	}
}
