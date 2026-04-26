package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pengaduan struct {
	ID        int    `json:"id"`
	Pengaduan string `json:"pengaduan"`
	FotoURL   string `json:"foto_url"`
	Status    string `json:"status"`
	Kategori  string `json:"kategori"`
	Nama      string `json:"nama"`
	NomorTelp string `json:"nomor_telp"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type PengaduanInput struct {
	ID        int    `json:"id"`
	Pengaduan string `json:"pengaduan"`
	FotoURL   string `json:"foto_url"`
	Kategori  string `json:"kategori"`
	Nama      string `json:"nama"`
	NomorTelp string `json:"nomor_telp"`
	Email     string `json:"email"`
}

// Handlers
// listPengaduanHandler godoc
// @Summary      List pengaduan
// @Description  Get a list of submitted complaints. Use ?status=submitted to filter. Leave empty to get all complaints.
// @Tags         pengaduan
// @Accept       json
// @Produce      json
// @Param        status  query     string  false  "Filter by status (submitted/draft)"
// @Success      200     {object}  map[string][]Pengaduan
// @Failure      503     {object}  map[string]string
// @Router       /pengaduan [get]

func listPengaduanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	status := c.Query("status") // optional filter
	query := "SELECT id, pengaduan, foto_url, kategori, nama, nomor_telp, email FROM pengaduan"
	args := []interface{}{}

	if status != "" && (status == "submitted") {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	pengaduan := []Pengaduan{}
	for rows.Next() {
		var p Pengaduan
		if err := rows.Scan(&p.ID, &p.Pengaduan, &p.FotoURL, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		pengaduan = append(pengaduan, p)
	}

	c.JSON(http.StatusOK, gin.H{"pengaduan": pengaduan})
}

// getPengaduanHandler godoc
// @Summary      Get a complaint
// @Description  Get a single complaint by its ID
// @Tags         pengaduan
// @Produce      json
// @Param        id   path      int  true  "Complaint ID"
// @Success      200  {object}  Pengaduan
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /pengaduan/{id} [get]

func getPengaduanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var p Pengaduan
	err = DB.QueryRow(
		"SELECT id, pengaduan, foto_url, kategori, nama, nomor_telp, email FROM pengaduan WHERE id = $1", id,
	).Scan(&p.ID, &p.Pengaduan, &p.FotoURL, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Complaint not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

// createPengaduanHandler godoc
// @Summary      Create complaint
// @Description  Create a new complaint
// @Tags         pengaduan
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        pengaduan  body      PengaduanInput  true  "Complaint Data"
// @Success      201      {object}  Pengaduan
// @Failure      400      {object}  map[string]string
// @Router       /pengaduan [post]

func createPengaduanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input PengaduanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Kategori == "" {
		input.Kategori = "Umum"
	}

	var p Pengaduan
	err := DB.QueryRow(
		`INSERT INTO pengaduan (pengaduan, foto_url, kategori, nama, nomor_telp, email)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, pengaduan, foto_url, kategori, nama, nomor_telp, email`,
		input.Pengaduan, input.FotoURL, input.Kategori, input.Nama, input.NomorTelp, input.Email,
	).Scan(&p.ID, &p.Pengaduan, &p.FotoURL, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}
