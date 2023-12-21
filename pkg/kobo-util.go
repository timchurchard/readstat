package pkg

import (
	"os"
	"path/filepath"
	"strings"
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

/*// cleanContentFilename takes a
func cleanContentFilename(BookID string) (string, string) {

}*/
