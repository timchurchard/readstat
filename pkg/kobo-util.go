package pkg

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// getDevice read the .kobo/version file for device id
func getDevice(fn string) (string, error) {
	dir, _ := filepath.Split(fn)
	versionFn := filepath.Join(dir, "version")

	data, err := os.ReadFile(versionFn) // todo: read limited amount?
	if err != nil {
		return "", err
	}

	stringData := string(data)
	firstComma := strings.Index(stringData, ",")

	return stringData[:firstComma], nil
}

// getModel simple lookup for device to human readable model
func getModel(device string) string {
	var (
		// models from https://help.kobo.com/hc/en-us/articles/360019676973-Identify-your-Kobo-eReader-or-Kobo-tablet
		models = map[string]string{
			"N605": "Kobo Elipsa 2E",
			"N506": "Kobo Clara 2E",
			"N778": "Kobo Sage",
			"N418": "Kobo Libra 2",
			"N604": "Kobo Elipsa",
			"N306": "Kobo Nia",
			"N873": "Kobo Libra H2O",
			"N782": "Kobo Forma",
			"N249": "Kobo Clara HD",
			"N867": "Kobo Aura H2O Edition 2",
			"N709": "Kobo Aura ONE",
			"N236": "Kobo Aura Edition 2",
			"N587": "Kobo Touch 2",
			"N437": "Kobo Glu HD",
			"N250": "Kobo Aura H2O",
			"N514": "Kobo Aura",
			"N204": "Kobo Aura HD",
			"N613": "Kobo Glo",
			"N705": "Kobo Mini",
			"N905": "Kobo Touch",
			"N416": "Kobo Original",
			"N647": "Kobo Wireless",
			"N47B": "Kobo Wireless",
		}
	)

	if name, exists := models[device[:4]]; exists {
		return name
	}

	return ""
}

// cleanContentFilename takes a "file:///mnt/onboard/dir/file.epub" and returns "dir/file.epub"
func cleanContentFilename(fn string) string {
	startPos := strings.Index(fn, KoboFilenamePrefix)

	if startPos == -1 {
		// Pocket articles have an ID so no cleaning needed
		return fn
	}

	return fn[startPos+len(KoboFilenamePrefix):]
}

// splitContentFilename takes a "/mnt/onboard/dir/file.epub!!part.html" and returns "dir/file.epub" and "part.html"
func splitContentFilename(fn string) (string, string) {
	startPos := strings.Index(fn, KoboFilenamePrefix)
	if startPos == -1 {
		startPos = 0
	}

	partPos := 0
	partLen := 0
	fUp := strings.ToUpper(fn)

	switch {
	case strings.Contains(fUp, koboFilePartSeparator):
		partPos = strings.Index(fUp, koboFilePartSeparator)
		partLen = len(koboFilePartSeparator)

	case strings.Contains(fUp, koboFilePartSeparatorAlt):
		partPos = strings.Index(fUp, koboFilePartSeparatorAlt)
		partLen = len(koboFilePartSeparatorAlt)

	case strings.Contains(fUp, koboDLPartSeparator):
		partPos = strings.Index(fUp, koboDLPartSeparator)
		partLen = len(koboDLPartSeparator)

	case strings.Contains(fUp, koboDLPartSeparatorLegacy):
		partPos = strings.Index(fUp, koboDLPartSeparatorLegacy)
		partLen = len(koboDLPartSeparatorLegacy)

	case strings.Contains(fUp, koboDLPartEpub):
		partPos = strings.Index(fUp, koboDLPartEpub)
		partLen = len(koboDLPartEpub)
	}

	if partPos == 0 {
		return fn[startPos:], ""
	}

	return fn[startPos:partPos], fn[partPos+partLen:]
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil // todo: good enough?
}
