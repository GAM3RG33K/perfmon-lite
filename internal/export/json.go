package export

import (
	"encoding/json"
	"os"
)

// ExportJSON writes the export data as a formatted JSON file.
func ExportJSON(data ExportData, opts Options) (string, error) {
	outPath := outputFilename(opts.OutputPath, FormatJSON)

	f, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(data); err != nil {
		return "", err
	}

	return outPath, nil
}
