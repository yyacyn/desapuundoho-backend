package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProdukDesa struct {
	ID        int     `json:"id"`
	Nama      string  `json:"nama"`
	Deskripsi string  `json:"deskripsi"`
	Harga     int64   `json:"harga"`
	Rating    float64 `json:"rating"`
	Kontak    string  `json:"kontak"`
	ImageURL  string  `json:"image_url"`
	CreatedAt string  `json:"created_at"`
}

type ProdukDesaInput struct {
	Nama      string  `json:"nama" binding:"required"`
	Deskripsi string  `json:"deskripsi"`
	Harga     int64   `json:"harga" binding:"required"`
	Rating    float64 `json:"rating"`
	Kontak    string  `json:"kontak" binding:"required"`
	ImageURL  string  `json:"image_url"`
}

// GET /api/produk-desa
func listProdukDesaHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	rows, err := DB.Query(`SELECT id, nama, COALESCE(deskripsi,''), harga, rating, kontak, COALESCE(image_url,''), created_at FROM produk_desa ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	produk := []ProdukDesa{}
	for rows.Next() {
		var p ProdukDesa
		if err := rows.Scan(&p.ID, &p.Nama, &p.Deskripsi, &p.Harga, &p.Rating, &p.Kontak, &p.ImageURL, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		produk = append(produk, p)
	}

	c.JSON(http.StatusOK, gin.H{"produk": produk})
}

// POST /api/produk-desa
func createProdukDesaHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input ProdukDesaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var p ProdukDesa
	err := DB.QueryRow(
		`INSERT INTO produk_desa (nama, deskripsi, harga, rating, kontak, image_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, nama, COALESCE(deskripsi,''), harga, rating, kontak, COALESCE(image_url,''), created_at`,
		input.Nama, input.Deskripsi, input.Harga, input.Rating, input.Kontak, input.ImageURL,
	).Scan(&p.ID, &p.Nama, &p.Deskripsi, &p.Harga, &p.Rating, &p.Kontak, &p.ImageURL, &p.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

// PUT /api/produk-desa/:id
func updateProdukDesaHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input ProdukDesaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var p ProdukDesa
	err = DB.QueryRow(
		`UPDATE produk_desa SET nama=$1, deskripsi=$2, harga=$3, rating=$4, kontak=$5, image_url=$6
		 WHERE id=$7
		 RETURNING id, nama, COALESCE(deskripsi,''), harga, rating, kontak, COALESCE(image_url,''), created_at`,
		input.Nama, input.Deskripsi, input.Harga, input.Rating, input.Kontak, input.ImageURL, id,
	).Scan(&p.ID, &p.Nama, &p.Deskripsi, &p.Harga, &p.Rating, &p.Kontak, &p.ImageURL, &p.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

// DELETE /api/produk-desa/:id
func deleteProdukDesaHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result, err := DB.Exec("DELETE FROM produk_desa WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk dihapus"})
}
