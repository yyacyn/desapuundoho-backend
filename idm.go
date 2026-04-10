package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetLiveIDM handles fetching live IDM data directly from Kemendesa API
// GetLiveIDM godoc
// @Summary      Get Live IDM Data
// @Description  Fetch live Index Desa Membangun (IDM) data directly from Kemendesa API
// @Tags         external
// @Produce      json
// @Param        tahun  query     string  false  "Year of data (default: 2023)"
// @Success      200    {object}  object
// @Failure      502    {object}  map[string]string
// @Router       /idm [get]
func GetLiveIDM(c *gin.Context) {
	// Izinkan parameter query tahun, contoh: /api/idm?tahun=2024
	year := c.Query("tahun")
	if year == "" {
		year = "2023" // Default ke 2023 jika tidak dispesifikasikan (karena data 2023 dijamin ada oleh user)
	}

	// Kode spesifik untuk Puundoho, Pakue Utara pada sistem IDM
	locationCode := "7408112001"
	kemendesaURL := fmt.Sprintf("https://idm.kemendesa.go.id/open/api/desa/rumusan/%s/%s", locationCode, year)

	// Create client
	client := &http.Client{}
	req, err := http.NewRequest("GET", kemendesaURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Add common headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", "https://idm.kemendesa.go.id/")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch from Kemendesa IDM API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Kemendesa IDM API returned non-200 status"})
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read API response"})
		return
	}

	// Meneruskan (proxy) response asli Kemendesa secara mentah tanpa modifikasi langsung ke Frontend
	c.Data(http.StatusOK, "application/json", bodyBytes)
}
