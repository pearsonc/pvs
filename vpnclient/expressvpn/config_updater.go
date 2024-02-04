package expressvpn

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ValidateConfig(configFile string) error {
	// Read the current config file into memory
	lines, err := readConfigFile(configFile)
	if err != nil {
		return err
	}

	// Define the changes and checks as key-value pairs or commands
	changes := map[string]string{
		"up /etc/openvpn/update-resolv-conf": "script-security 2\nup /etc/openvpn/update-resolv-conf\ndown /etc/openvpn/update-resolv-conf",
		"cipher AES-256-CBC":                 "data-ciphers AES-256-GCM",
		"ns-cert-type server":                "remote-cert-tls server",
		"redirect-gateway def1":              "redirect-gateway def1",
		"keepalive 60 120":                   "keepalive 60 120",
	}
	removeLines := []string{"keysize 256", "redirect-gateway"}

	// Update the config based on the defined changes
	updatedLines := updateConfig(lines, changes, removeLines)

	// Write the updated configuration back to the file
	if err := writeConfigFile(configFile, updatedLines); err != nil {
		return err
	}

	return nil
}

func readConfigFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return lines, nil
}

func updateConfig(lines []string, changes map[string]string, removeLines []string) []string {
	var updatedLines []string
	changesApplied := make(map[string]bool)

	for _, line := range lines {
		updated := false
		for check, change := range changes {
			if strings.Contains(line, check) && !changesApplied[check] {
				for _, newLine := range strings.Split(change, "\n") {
					updatedLines = append(updatedLines, newLine)
				}
				changesApplied[check] = true
				updated = true
				break // Apply only one change per line to avoid duplicating replacements
			}
		}
		if !updated && !shouldRemoveLine(line, removeLines) {
			updatedLines = append(updatedLines, line)
		}
	}

	// Append any changes that weren't applied because the check string wasn't found
	for check, change := range changes {
		if !changesApplied[check] {
			for _, newLine := range strings.Split(change, "\n") {
				updatedLines = append(updatedLines, newLine)
			}
		}
	}

	return updatedLines
}

func shouldRemoveLine(line string, removeLines []string) bool {
	for _, remove := range removeLines {
		if strings.Contains(line, remove) {
			return true
		}
	}
	return false
}

func writeConfigFile(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}
	return writer.Flush()
}
