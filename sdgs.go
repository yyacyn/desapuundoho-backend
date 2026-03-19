package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SDGResponse mirrors the JSON response from Kemendesa
type SDGResponse struct {
	Average string `json:"average"`
	Data    []struct {
		Goals int     `json:"goals"`
		Title string  `json:"title"`
		Image string  `json:"image"`
		Score float64 `json:"score"`
	} `json:"data"`
	TotalDesa int `json:"total_desa"`
}

// GetLiveSDGS handles fetching live SDG data directly from Kemendesa API
func GetLiveSDGS(c *gin.Context) {
	// The specific location code for Puundoho: 7408051004
	kemendesaURL := "https://sid.kemendesa.go.id/sdgs/searching/score-sdgs?location_code=7408051004"

	// Create client
	client := &http.Client{}
	req, err := http.NewRequest("GET", kemendesaURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Add common headers to mimic a normal browser request (prevent blocking)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", "https://sid.kemendesa.go.id/sdgs")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch from Kemendesa API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Kemendesa API returned non-200 status"})
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read API response"})
		return
	}

	var sdgData SDGResponse
	if err := json.Unmarshal(bodyBytes, &sdgData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse API JSON"})
		return
	}

	c.JSON(http.StatusOK, sdgData)
}
