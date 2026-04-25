package excel

import (
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ValidateStrictColumns checks if the row contains all expected columns exactly (case-insensitive trim).
func ValidateStrictColumns(row []string, expectedHeaders []string) error {
	// Create a map for fast lookup of found columns in the row
	foundMap := make(map[string]bool)
	for _, col := range row {
		cleanCol := strings.TrimSpace(strings.ToUpper(col))
		foundMap[cleanCol] = true
	}

	var missing []string
	for _, expected := range expectedHeaders {
		cleanExpected := strings.TrimSpace(strings.ToUpper(expected))
		if !foundMap[cleanExpected] {
			missing = append(missing, expected)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required columns: %s", strings.Join(missing, ", "))
	}
	return nil
}

// findTargetSheetStrict searches for a sheet that passes the strict validator.
// It returns the sheet name, the header row index, and any validation error.
func findTargetSheetStrict(f *excelize.File, expectedHeaders []string) (string, int, error) {
	var lastErr error
	for _, name := range f.GetSheetList() {
		rows, err := f.GetRows(name, excelize.Options{RawCellValue: true})
		if err != nil || len(rows) == 0 {
			continue
		}

		// Check first 5 rows for the header
		for i := 0; i < 5 && i < len(rows); i++ {
			err := ValidateStrictColumns(rows[i], expectedHeaders)
			if err == nil {
				fmt.Printf("[DEBUG] Valid Strict Header Found in Sheet: %s at row %d\n", name, i)
				return name, i, nil
			}
			lastErr = err // Store the last validation error to show the user
		}
	}
	if lastErr != nil {
		return "", -1, fmt.Errorf("invalid file format: %v", lastErr)
	}
	return "", -1, fmt.Errorf("invalid file format: could not find matching headers")
}

// ParseExcelToJSONStrict parses the file, enforces exact column match, and returns rows starting from the header.
func ParseExcelToJSONStrict(fileHeader *multipart.FileHeader, expectedHeaders []string) ([][]string, error) {
	fmt.Printf("[DEBUG] Parsing Excel File (Strict Mode): %s\n", fileHeader.Filename)
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	// defer src.Close()
	defer func() { _ = src.Close() }()

	f, err := excelize.OpenReader(src)
	if err != nil {
		fmt.Printf("[DEBUG] OpenReader failed: %v\n", err)
		return nil, err
	}
	// defer f.Close()
	defer func() { _ = f.Close() }()
	sheetName, headerRowIdx, err := findTargetSheetStrict(f, expectedHeaders)
	if err != nil {
		fmt.Printf("[DEBUG] Strict Validation Failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] Found Target Sheet: %s at Row %d\n", sheetName, headerRowIdx)

	allRows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	
	// Slice the rows to start exactly at the header row
	if headerRowIdx >= len(allRows) {
		return nil, fmt.Errorf("no data found after header")
	}
	
	validRows := allRows[headerRowIdx:]
	fmt.Printf("[DEBUG] Read %d strict rows from sheet\n", len(validRows))
	return validRows, nil
}
