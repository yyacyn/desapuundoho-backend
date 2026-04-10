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
// listProdukDesaHandler godoc
// @Summary      List Produk Desa
// @Description  Get a list of village products
// @Tags         produk
// @Produce      json
// @Success      200  {object}  map[string][]ProdukDesa
// @Failure      503  {object}  map[string]string
// @Router       /produk-desa [get]
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
// createProdukDesaHandler godoc
// @Summary      Create Produk Desa
// @Description  Create a new village product (Bendahara only)
// @Tags         produk
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        produk  body      ProdukDesaInput  true  "Produk Data"
// @Success      201     {object}  ProdukDesa
// @Failure      400     {object}  map[string]string
// @Router       /produk-desa [post]
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
// updateProdukDesaHandler godoc
// @Summary      Update Produk Desa
// @Description  Update an existing village product (Bendahara only)
// @Tags         produk
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id      path      int              true  "Produk ID"
// @Param        produk  body      ProdukDesaInput  true  "Update Data"
// @Success      200     {object}  ProdukDesa
// @Failure      404     {object}  map[string]string
// @Router       /produk-desa/{id} [put]
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
// deleteProdukDesaHandler godoc
// @Summary      Delete Produk Desa
// @Description  Delete a village product by ID (Bendahara only)
// @Tags         produk
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "Produk ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /produk-desa/{id} [delete]
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
