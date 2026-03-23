package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParsedPDFTable represents a single extracted table from a PDF
type ParsedPDFTable struct {
	Title   string     `json:"title"`
	Rows    [][]string `json:"rows"`
	PageNum int        `json:"page_number"`
}

// ODLElement represents a single element in OpenDataLoader JSON output
type ODLElement struct {
	Type       string      `json:"type"`
	Content    interface{} `json:"content"`
	PageNumber int         `json:"page number"`
	ID         int         `json:"id"`
}

// parsePDFHandler accepts a PDF or JSON file upload to extract tables
func parsePDFHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File diperlukan"})
		return
	}
	defer file.Close()

	// Validate it's a PDF or JSON
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" && ext != ".json" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya file PDF atau JSON yang diterima"})
		return
	}

	// Save to temp directory
	tmpDir, err := os.MkdirTemp("", "apbdes-parser-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat direktori sementara"})
		return
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, header.Filename)
	outputDir := filepath.Join(tmpDir, "output")
	os.MkdirAll(outputDir, 0755)

	out, err := os.Create(inputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
		return
	}
	io.Copy(out, file)
	out.Close()

	var pdfText string

	if ext == ".pdf" {
		// Use pdf-parse for raw text extraction instead of OpenDataLoader
		// Create script and package.json in temp dir
		script := `
const fs = require('fs');
const pdfParse = require('pdf-parse');
let dataBuffer = fs.readFileSync(process.argv[2]);
pdfParse(dataBuffer).then(function(data) {
    fs.writeFileSync(process.argv[3], data.text);
}).catch(err => {
    console.error(err);
    process.exit(1);
});
`
		scriptPath := filepath.Join(tmpDir, "extract.js")
		os.WriteFile(scriptPath, []byte(script), 0644)

		pkgJson := `{"dependencies": {"pdf-parse": "^1.1.1"}}`
		os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJson), 0644)

		// Install pdf-parse
		cmdInstall := exec.Command("npm", "install", "--silent")
		cmdInstall.Dir = tmpDir
		cmdInstall.Run()

		// Run extraction
		txtPath := filepath.Join(tmpDir, "out.txt")
		cmdRun := exec.Command("node", "extract.js", inputPath, txtPath)
		cmdRun.Dir = tmpDir
		cmdOutput, err := cmdRun.CombinedOutput()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal mengekstrak teks PDF menggunakan pdf-parse.",
				"details": string(cmdOutput),
			})
			return
		}

		extractedBytes, _ := os.ReadFile(txtPath)
		pdfText = string(extractedBytes)
	} else {
		// File is JSON, read it directly
		jsonData, err := os.ReadFile(inputPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file JSON"})
			return
		}
		pdfText = string(jsonData)
	}

	if strings.TrimSpace(pdfText) == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tidak ada teks yang dapat diekstrak dari dokumen"})
		return
	}

	var tables []ParsedPDFTable

	// Try semantic APBDes JSON if it was a JSON upload
	if ext == ".json" {
		tables = parseSemanticAPBDesJSON([]byte(pdfText))
	}

	// Try extracting APBDes codes via RegEx from raw text (Works for PDF text AND generic JSON)
	if len(tables) == 0 {
		tables = extractAPBDesFromTextStr(pdfText)
	}

	// Raw text fallback (last resort)
	if len(tables) == 0 {
		var textRows [][]string
		lines := strings.Split(pdfText, "\n")
		// take only first 50 lines to not overwhelm if it's huge unformatted garbage
		maxLines := len(lines)
		if maxLines > 50 {
			maxLines = 50
		}
		for i := 0; i < maxLines; i++ {
			line := strings.TrimSpace(lines[i])
			if line != "" {
				textRows = append(textRows, []string{"Teks", line})
			}
		}
		if len(textRows) > 0 {
			tables = []ParsedPDFTable{{
				Title: "Konten Teks Diekstrak (Raw / Unformatted)",
				Rows:  append([][]string{{"Tipe", "Konten"}}, textRows...),
			}}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tables":   tables,
		"filename": header.Filename,
		"message":  fmt.Sprintf("Berhasil mengekstrak %d tabel dari dokumen", len(tables)),
	})
}

// parseSemanticAPBDesJSON parses highly-structured JSON APBDes (e.g. LLM extracted)
func parseSemanticAPBDesJSON(data []byte) []ParsedPDFTable {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil
	}

	// Identify it's the APBDes semantic format
	if _, hasDok := doc["dokumen"]; !hasDok {
		return nil
	}
	if _, hasPendapatan := doc["pendapatan"]; !hasPendapatan {
		return nil
	}

	var tables []ParsedPDFTable

	if p, ok := doc["pendapatan"].(map[string]interface{}); ok {
		tables = append(tables, ParsedPDFTable{
			Title: "Anggaran Pendapatan",
			Rows:  flattenSemanticNode(p, true),
		})
	}

	if b, ok := doc["belanja"].(map[string]interface{}); ok {
		tables = append(tables, ParsedPDFTable{
			Title: "Anggaran Belanja",
			Rows:  flattenSemanticNode(b, true),
		})
	}

	if pb, ok := doc["pembiayaan"].(map[string]interface{}); ok {
		tables = append(tables, ParsedPDFTable{
			Title: "Pembiayaan",
			Rows:  flattenSemanticNode(pb, true),
		})
	}

	return tables
}

func flattenSemanticNode(node map[string]interface{}, isRoot bool) [][]string {
	var rows [][]string
	if isRoot {
		rows = append(rows, []string{"Kode", "Uraian", "Anggaran (Rp)", "Realisasi (Rp)"})
	}

	kode, _ := node["kode"].(string)
	uraian, _ := node["uraian"].(string)
	
	anggaranStr := ""
	if ang := getFloat(node["anggaran"]); ang > 0 {
		anggaranStr = fmt.Sprintf("%.0f", ang)
	}
	realisasiStr := ""
	if real := getFloat(node["realisasi"]); real > 0 {
		realisasiStr = fmt.Sprintf("%.0f", real)
	}

	if kode != "" || uraian != "" {
		rows = append(rows, []string{kode, uraian, anggaranStr, realisasiStr})
	}

	if rincian, ok := node["rincian"].([]interface{}); ok {
		for _, r := range rincian {
			if rMap, ok := r.(map[string]interface{}); ok {
				rows = append(rows, flattenSemanticNode(rMap, false)...)
			}
		}
	}

	if items, ok := node["item"].([]interface{}); ok {
		for _, i := range items {
			if iMap, ok := i.(map[string]interface{}); ok {
				rows = append(rows, flattenSemanticNode(iMap, false)...)
			}
		}
	}

	return rows
}

func getFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	}
	return 0
}

// extractTablesFromODLJSON parses OpenDataLoader generic elements json
// (Kept for backwards compatibility if needed, though now unused)
func extractTablesFromODLJSON(data []byte) []ParsedPDFTable {
	var tables []ParsedPDFTable

	// Try parsing as array of elements
	var elements []ODLElement
	if err := json.Unmarshal(data, &elements); err == nil {
		tableIndex := 0
		for _, elem := range elements {
			if elem.Type == "table" {
				tableIndex++
				contentStr := fmt.Sprintf("%v", elem.Content)
				rows := parseTableContent(contentStr)
				if len(rows) > 0 {
					tables = append(tables, ParsedPDFTable{
						Title:   fmt.Sprintf("Tabel %d (Halaman %d)", tableIndex, elem.PageNumber),
						Rows:    rows,
						PageNum: elem.PageNumber,
					})
				}
			}
		}
		return tables
	}

	// Try parsing as object with elements array
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err == nil {
		if elems, ok := doc["elements"].([]interface{}); ok {
			tableIndex := 0
			for _, e := range elems {
				if em, ok := e.(map[string]interface{}); ok {
					if t, _ := em["type"].(string); t == "table" {
						tableIndex++
						contentStr := fmt.Sprintf("%v", em["content"])
						pageNum, _ := em["page number"].(float64)
						rows := parseTableContent(contentStr)
						if len(rows) > 0 {
							tables = append(tables, ParsedPDFTable{
								Title:   fmt.Sprintf("Tabel %d (Halaman %d)", tableIndex, int(pageNum)),
								Rows:    rows,
								PageNum: int(pageNum),
							})
						}
					}
				}
			}
		}
	}

	return tables
}

// extractOrderedTextFromODL deeply searches the OpenDataLoader JSON structure
// to extract all text content in document order.
// (Kept for backwards compatibility if needed, though now unused)
func extractOrderedTextFromODL(data []byte) string {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		
		// If it's a flat array instead
		var elements []ODLElement
		if err2 := json.Unmarshal(data, &elements); err2 == nil {
			var sb strings.Builder
			for _, e := range elements {
				if e.Content != nil {
					sb.WriteString(fmt.Sprintf("%v\n", e.Content))
				}
			}
			return sb.String()
		}
		return ""
	}

	var sb strings.Builder

	// Traverse pages -> elements -> content
	if pages, ok := doc["pages"].([]interface{}); ok {
		for _, p := range pages {
			if pm, ok := p.(map[string]interface{}); ok {
				if elems, ok := pm["elements"].([]interface{}); ok {
					for _, e := range elems {
						if em, ok := e.(map[string]interface{}); ok {
							if content, ok := em["content"]; ok && content != nil {
								sb.WriteString(fmt.Sprintf("%v\n", content))
							}
						}
					}
				}
			}
		}
	} else if elems, ok := doc["elements"].([]interface{}); ok {
		// Or flat elements array inside object
		for _, e := range elems {
			if em, ok := e.(map[string]interface{}); ok {
				if content, ok := em["content"]; ok && content != nil {
					sb.WriteString(fmt.Sprintf("%v\n", content))
				}
			}
		}
	}

	return sb.String()
}

// parseTableContent parses generic string table into cells
func parseTableContent(content string) [][]string {
	var rows [][]string

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Count(line, "-") > len(line)/2 && strings.Contains(line, "|") {
			continue
		}

		var cells []string
		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					cells = append(cells, p)
				}
			}
		} else if strings.Contains(line, "\t") {
			parts := strings.Split(line, "\t")
			for _, p := range parts {
				cells = append(cells, strings.TrimSpace(p))
			}
		} else {
			cells = splitByMultipleSpaces(line)
		}

		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	}

	return rows
}

func splitByMultipleSpaces(s string) []string {
	var result []string
	current := ""
	spaceCount := 0

	for _, ch := range s {
		if ch == ' ' {
			spaceCount++
		} else {
			if spaceCount >= 2 && current != "" {
				result = append(result, strings.TrimSpace(current))
				current = ""
			} else if spaceCount > 0 {
				current += strings.Repeat(" ", spaceCount)
			}
			spaceCount = 0
			current += string(ch)
		}
	}
	if strings.TrimSpace(current) != "" {
		result = append(result, strings.TrimSpace(current))
	}
	if len(result) <= 1 {
		return []string{s}
	}
	return result
}

// extractTablesFromMarkdown parses markdown-formatted tables
// (Kept for backwards compatibility if needed, though now unused)
func extractTablesFromMarkdown(md string) []ParsedPDFTable {
	var tables []ParsedPDFTable
	lines := strings.Split(md, "\n")
	var currentTable [][]string
	tableIndex := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
			trimmed := strings.Trim(line, "|")
			if strings.Count(trimmed, "-") > len(trimmed)/2 {
				continue
			}
			parts := strings.Split(line, "|")
			var cells []string
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					cells = append(cells, p)
				}
			}
			if len(cells) > 0 {
				currentTable = append(currentTable, cells)
			}
		} else {
			if len(currentTable) > 0 {
				tableIndex++
				tables = append(tables, ParsedPDFTable{
					Title: fmt.Sprintf("Tabel %d", tableIndex),
					Rows:  currentTable,
				})
				currentTable = nil
			}
		}
	}
	if len(currentTable) > 0 {
		tableIndex++
		tables = append(tables, ParsedPDFTable{
			Title: fmt.Sprintf("Tabel %d", tableIndex),
			Rows:  currentTable,
		})
	}
	return tables
}

// extractAPBDesFromText tries to reconstruct APBDes tables from raw text
func extractAPBDesFromText(elements []ODLElement) []ParsedPDFTable {
	var pendapatanRows [][]string
	var belanjaRows [][]string
	var pembiayaanRows [][]string

	currentCategory := ""
	reCode := regexp.MustCompile(`^([456](?:\.\d+)*)\s+(.*)$`)
	// Enhanced RegEx: Number can be positive, negative (parentheses), and may have spaces around
	reNums := regexp.MustCompile(`((?:[\d\.]+(?:,\d+)?|\([\d\.]+(?:,\d+)?\))\s*)+$`)

	for _, elem := range elements {
		content := fmt.Sprintf("%v", elem.Content)
		lines := strings.Split(content, "\n")
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			matches := reCode.FindStringSubmatch(line)
			if len(matches) > 1 {
				code := strings.TrimSpace(matches[1])
				rest := strings.TrimSpace(matches[2])
				
				if strings.HasPrefix(code, "4") {
					currentCategory = "pendapatan"
				} else if strings.HasPrefix(code, "5") {
					currentCategory = "belanja"
				} else if strings.HasPrefix(code, "6") {
					currentCategory = "pembiayaan"
				}

				numMatches := reNums.FindStringSubmatch(rest)
				uraian := rest
				anggaran := ""
				realisasi := ""

				if len(numMatches) > 1 {
					numStr := strings.TrimSpace(numMatches[1])
					uraian = strings.TrimSpace(rest[:len(rest)-len(numMatches[1])])
					
					nums := strings.Fields(numStr)
					if len(nums) >= 1 {
						anggaran = cleanRupiah(nums[0])
					}
					if len(nums) >= 2 {
						realisasi = cleanRupiah(nums[1])
					}
				}

				row := []string{code, uraian, anggaran, realisasi}

				if currentCategory == "pendapatan" {
					pendapatanRows = append(pendapatanRows, row)
				} else if currentCategory == "belanja" {
					belanjaRows = append(belanjaRows, row)
				} else if currentCategory == "pembiayaan" {
					pembiayaanRows = append(pembiayaanRows, row)
				}
			}
		}
	}

	var tables []ParsedPDFTable
	if len(pendapatanRows) > 0 {
		tables = append(tables, ParsedPDFTable{
			Title: "Pendapatan (Ekstrak Teks)", 
			Rows: append([][]string{{"Kode", "Uraian", "Anggaran", "Realisasi"}}, pendapatanRows...)})
	}
	if len(belanjaRows) > 0 {
		tables = append(tables, ParsedPDFTable{
			Title: "Belanja (Ekstrak Teks)", 
			Rows: append([][]string{{"Kode", "Uraian", "Anggaran", "Realisasi"}}, belanjaRows...)})
	}
	if len(pembiayaanRows) > 0 {
		tables = append(tables, ParsedPDFTable{
			Title: "Pembiayaan (Ekstrak Teks)", 
			Rows: append([][]string{{"Kode", "Uraian", "Anggaran", "Realisasi"}}, pembiayaanRows...)})
	}

	return tables
}

func cleanRupiah(s string) string {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		s = "-" + s[1:len(s)-1]
	}
	parts := strings.Split(s, ".")
	return parts[0]
}

// extractAPBDesFromTextStr wraps plain text to reuse the element-based text extractor
func extractAPBDesFromTextStr(text string) []ParsedPDFTable {
	var elements []ODLElement
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		elements = append(elements, ODLElement{Content: line})
	}
	return extractAPBDesFromText(elements)
}

