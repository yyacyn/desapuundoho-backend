package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DusunBoundary struct {
	ID          int    `json:"id"`
	NamaDusun   string `json:"nama_dusun"`
	Warna       string `json:"warna"`
	GeojsonData string `json:"geojson_data"`
	CreatedAt   string `json:"created_at"`
}

func listDusunHandler(c *gin.Context) {
	rows, err := DB.Query("SELECT id, nama_dusun, warna, geojson_data::text, created_at FROM dusun_boundaries ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var dusuns []DusunBoundary
	for rows.Next() {
		var d DusunBoundary
		if err := rows.Scan(&d.ID, &d.NamaDusun, &d.Warna, &d.GeojsonData, &d.CreatedAt); err != nil {
			continue
		}
		dusuns = append(dusuns, d)
	}

	// Make sure we don't return null slice
	if dusuns == nil {
		dusuns = []DusunBoundary{}
	}

	c.JSON(http.StatusOK, gin.H{"dusun": dusuns})
}

func createDusunHandler(c *gin.Context) {
	var input struct {
		NamaDusun   string `json:"nama_dusun" binding:"required"`
		Warna       string `json:"warna" binding:"required"`
		GeojsonData string `json:"geojson_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var newID int
	var createdAt string
	err := DB.QueryRow(
		"INSERT INTO dusun_boundaries (nama_dusun, warna, geojson_data) VALUES ($1, $2, $3::jsonb) RETURNING id, created_at",
		input.NamaDusun, input.Warna, input.GeojsonData,
	).Scan(&newID, &createdAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dusun: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, DusunBoundary{
		ID:          newID,
		NamaDusun:   input.NamaDusun,
		Warna:       input.Warna,
		GeojsonData: input.GeojsonData,
		CreatedAt:   createdAt,
	})
}

func updateDusunHandler(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		NamaDusun   string `json:"nama_dusun" binding:"required"`
		Warna       string `json:"warna" binding:"required"`
		GeojsonData string `json:"geojson_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var updatedId int
	var createdAt string
	err := DB.QueryRow(
		"UPDATE dusun_boundaries SET nama_dusun = $1, warna = $2, geojson_data = $3::jsonb, updated_at = NOW() WHERE id = $4 RETURNING id, created_at",
		input.NamaDusun, input.Warna, input.GeojsonData, id,
	).Scan(&updatedId, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Dusun not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update dusun"})
		}
		return
	}

	c.JSON(http.StatusOK, DusunBoundary{
		ID:          updatedId,
		NamaDusun:   input.NamaDusun,
		Warna:       input.Warna,
		GeojsonData: input.GeojsonData,
		CreatedAt:   createdAt,
	})
}

func deleteDusunHandler(c *gin.Context) {
	id := c.Param("id")
	res, err := DB.Exec("DELETE FROM dusun_boundaries WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dusun"})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dusun not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}
