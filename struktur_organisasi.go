package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Deprecated: previous ImageUpload types removed in favor of StrukturOrganisasi

// Handlers
// StrukturOrganisasi represents a Struktur Organisasi image record.
// swagger:model StrukturOrganisasi
type StrukturOrganisasi struct {
	ID        int    `json:"id"`
	ImageURL  string `json:"image_url"`
	Caption   string `json:"caption"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// StrukturOrganisasiInput represents payload for creating/updating Struktur Organisasi.
// swagger:model StrukturOrganisasiInput
type StrukturOrganisasiInput struct {
	ImageURL string `json:"image_url" binding:"required"`
	Caption  string `json:"caption"`
}

// listStrukturOrganisasiHandler godoc
// @Summary      List Struktur Organisasi
// @Description  Get a list of struktur organisasi images
// @Tags         struktur-organisasi
// @Produce      json
// @Success      200  {object}  map[string][]StrukturOrganisasi
// @Failure      503  {object}  map[string]string
// @Router       /struktur-organisasi [get]
func listStrukturOrganisasiHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	rows, err := DB.Query(`SELECT id, image_url, COALESCE(caption,''), created_at, updated_at FROM struktur_organisasi ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	items := []StrukturOrganisasi{}
	for rows.Next() {
		var item StrukturOrganisasi
		if err := rows.Scan(&item.ID, &item.ImageURL, &item.Caption, &item.CreatedAt, &item.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"struktur_organisasi": items})
}

// getStrukturOrganisasiHandler godoc
// @Summary      Get Struktur Organisasi
// @Description  Get a single struktur organisasi image by ID
// @Tags         struktur-organisasi
// @Produce      json
// @Param        id   path      int  true  "ID"
// @Success      200  {object}  StrukturOrganisasi
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      503  {object}  map[string]string
// @Router       /struktur-organisasi/{id} [get]
func getStrukturOrganisasiHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var item StrukturOrganisasi
	err = DB.QueryRow(`SELECT id, image_url, COALESCE(caption,''), created_at, updated_at FROM struktur_organisasi WHERE id = $1`, id).
		Scan(&item.ID, &item.ImageURL, &item.Caption, &item.CreatedAt, &item.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Struktur organisasi not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// createStrukturOrganisasiHandler godoc
// @Summary      Create Struktur Organisasi
// @Description  Create a new struktur organisasi image record (Admin only)
// @Tags         struktur-organisasi
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        upload  body      StrukturOrganisasiInput  true  "Upload Data"
// @Success      201     {object}  StrukturOrganisasi
// @Failure      400     {object}  map[string]string
// @Failure      503     {object}  map[string]string
// @Router       /struktur-organisasi [post]
func createStrukturOrganisasiHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input StrukturOrganisasiInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var item StrukturOrganisasi
	err := DB.QueryRow(
		`INSERT INTO struktur_organisasi (image_url, caption)
		 VALUES ($1, $2)
		 RETURNING id, image_url, COALESCE(caption,''), created_at, updated_at`,
		input.ImageURL, input.Caption,
	).Scan(&item.ID, &item.ImageURL, &item.Caption, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// updateStrukturOrganisasiHandler godoc
// @Summary      Update Struktur Organisasi
// @Description  Update an existing struktur organisasi image record (Admin only)
// @Tags         struktur-organisasi
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id      path      int               true  "ID"
// @Param        upload  body      StrukturOrganisasiInput   true  "Update Data"
// @Success      200     {object}  StrukturOrganisasi
// @Failure      400     {object}  map[string]string
// @Failure      404     {object}  map[string]string
// @Failure      503     {object}  map[string]string
// @Router       /struktur-organisasi/{id} [put]
func updateStrukturOrganisasiHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input StrukturOrganisasiInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var item StrukturOrganisasi
	err = DB.QueryRow(
		`UPDATE struktur_organisasi
		 SET image_url = $1,
			 caption = $2,
			 updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, image_url, COALESCE(caption,''), created_at, updated_at`,
		input.ImageURL, input.Caption, id,
	).Scan(&item.ID, &item.ImageURL, &item.Caption, &item.CreatedAt, &item.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Struktur organisasi not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// deleteStrukturOrganisasiHandler godoc
// @Summary      Delete Struktur Organisasi
// @Description  Delete a struktur organisasi image record by ID (Admin only)
// @Tags         struktur-organisasi
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      503  {object}  map[string]string
// @Router       /struktur-organisasi/{id} [delete]
func deleteStrukturOrganisasiHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result, err := DB.Exec("DELETE FROM struktur_organisasi WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Struktur organisasi not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Struktur organisasi deleted"})
}
