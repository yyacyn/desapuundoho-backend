package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Data types
type Listing struct {
	ID        int    `json:"id"`
	Nama      string `json:"nama"`
	Koordinat string `json:"koordinat"`
	ImageURL  string `json:"image_url"`
	CreatedAt string `json:"created_at"`
}

type ListingInput struct {
	Nama      string `json:"nama" binding:"required"`
	Koordinat string `json:"koordinat" binding:"required"`
	ImageURL  string `json:"image_url"`
}

// Handlers

// GET /api/listings
func listListingsHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	query := "SELECT id, nama, koordinat, COALESCE(image_url,''), created_at FROM listings ORDER BY created_at DESC"
	rows, err := DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	listings := []Listing{}
	for rows.Next() {
		var l Listing
		if err := rows.Scan(&l.ID, &l.Nama, &l.Koordinat, &l.ImageURL, &l.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		listings = append(listings, l)
	}

	c.JSON(http.StatusOK, gin.H{"listings": listings})
}

// GET /api/listings/:id
func getListingHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var l Listing
	err = DB.QueryRow(
		"SELECT id, nama, koordinat, COALESCE(image_url,''), created_at FROM listings WHERE id = $1", id,
	).Scan(&l.ID, &l.Nama, &l.Koordinat, &l.ImageURL, &l.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Listing not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, l)
}

// POST /api/listings (protected)
func createListingHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input ListingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var l Listing
	err := DB.QueryRow(
		`INSERT INTO listings (nama, koordinat, image_url)
		 VALUES ($1, $2, $3)
		 RETURNING id, nama, koordinat, COALESCE(image_url,''), created_at`,
		input.Nama, input.Koordinat, input.ImageURL,
	).Scan(&l.ID, &l.Nama, &l.Koordinat, &l.ImageURL, &l.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, l)
}

// PUT /api/listings/:id (protected)
func updateListingHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input ListingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var l Listing
	err = DB.QueryRow(
		`UPDATE listings SET nama=$1, koordinat=$2, image_url=$3
		 WHERE id=$4
		 RETURNING id, nama, koordinat, COALESCE(image_url,''), created_at`,
		input.Nama, input.Koordinat, input.ImageURL, id,
	).Scan(&l.ID, &l.Nama, &l.Koordinat, &l.ImageURL, &l.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Listing not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, l)
}

// DELETE /api/listings/:id (protected)
func deleteListingHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result, err := DB.Exec("DELETE FROM listings WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Listing not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Listing deleted"})
}
