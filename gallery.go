package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Data types
type GalleryItem struct {
	ID        int      `json:"id"`
	Images    []string `json:"images"`
	Caption   string   `json:"caption"`
	CreatedAt string   `json:"created_at"`
}

type GalleryInput struct {
	Images  []string `json:"images" binding:"required"`
	Caption string   `json:"caption"`
}

// Handlers

// GET /api/galeri
func listGalleryHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	query := "SELECT id, images, COALESCE(caption,''), created_at FROM galeri ORDER BY created_at DESC"
	rows, err := DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	items := []GalleryItem{}
	for rows.Next() {
		var item GalleryItem
		var imagesJSON []byte
		if err := rows.Scan(&item.ID, &imagesJSON, &item.Caption, &item.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// Unmarshal JSONB images
		if err := json.Unmarshal(imagesJSON, &item.Images); err != nil {
			item.Images = []string{}
		}
		
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"galeri": items})
}

// GET /api/galeri/:id
func getGalleryHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var item GalleryItem
	var imagesJSON []byte
	err = DB.QueryRow(
		"SELECT id, images, COALESCE(caption,''), created_at FROM galeri WHERE id = $1", id,
	).Scan(&item.ID, &imagesJSON, &item.Caption, &item.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gallery item not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Unmarshal JSONB images
	if err := json.Unmarshal(imagesJSON, &item.Images); err != nil {
		item.Images = []string{}
	}

	c.JSON(http.StatusOK, item)
}

// POST /api/galeri (protected)
func createGalleryHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input GalleryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imagesJSON, _ := json.Marshal(input.Images)

	var item GalleryItem
	var resImagesJSON []byte
	err := DB.QueryRow(
		`INSERT INTO galeri (images, caption)
		 VALUES ($1, $2)
		 RETURNING id, images, COALESCE(caption,''), created_at`,
		imagesJSON, input.Caption,
	).Scan(&item.ID, &resImagesJSON, &item.Caption, &item.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	json.Unmarshal(resImagesJSON, &item.Images)
	c.JSON(http.StatusCreated, item)
}

// PUT /api/galeri/:id (protected)
func updateGalleryHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input GalleryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imagesJSON, _ := json.Marshal(input.Images)

	var item GalleryItem
	var resImagesJSON []byte
	err = DB.QueryRow(
		`UPDATE galeri SET images=$1, caption=$2
		 WHERE id=$3
		 RETURNING id, images, COALESCE(caption,''), created_at`,
		imagesJSON, input.Caption, id,
	).Scan(&item.ID, &resImagesJSON, &item.Caption, &item.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gallery item not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	json.Unmarshal(resImagesJSON, &item.Images)
	c.JSON(http.StatusOK, item)
}

// DELETE /api/galeri/:id (protected)
func deleteGalleryHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result, err := DB.Exec("DELETE FROM galeri WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gallery item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gallery item deleted"})
}
