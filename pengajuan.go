package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pengajuan struct {
	ID         int    `json:"id"`
	Judul      string `json:"judul"`
	Isi        string `json:"isi"`
	DokumenURL string `json:"dokumen_url"`
	Status     string `json:"status"`
	Kategori   string `json:"kategori"`
	Nama       string `json:"nama"`
	NomorTelp  string `json:"nomor_telp"`
	Email      string `json:"email"`
	Lokasi     string `json:"lokasi"`
	Tanggal    string `json:"tanggal"`
	Keterangan string `json:"keterangan"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type PengajuanInput struct {
	Judul      string `json:"judul"`
	Isi        string `json:"isi"`
	DokumenURL string `json:"dokumen_url"`
	Kategori   string `json:"kategori"`
	Nama       string `json:"nama"`
	NomorTelp  string `json:"nomor_telp"`
	Email      string `json:"email"`
	Lokasi     string `json:"lokasi"`
	Tanggal    string `json:"tanggal"`
}

// Handlers
// listPengajuanHandler godoc
// @Summary      List pengajuan
// @Description  Get a list of submitted requests
// @Tags         pengajuan
// @Accept       json
// @Produce      json
// @Param        status  query     string  false  "Filter by status"
// @Success      200     {object}  map[string][]Pengajuan
// @Failure      503     {object}  map[string]string
// @Router       /pengajuan [get]

func listPengajuanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	status := c.Query("status")
	query := "SELECT id, judul, isi, dokumen_url, status, kategori, nama, nomor_telp, email, lokasi, tanggal, COALESCE(keterangan,''), created_at, updated_at FROM pengajuan"
	args := []interface{}{}

	if status != "" && (status == "Baru" || status == "Ditinjau" || status == "Disetujui" || status == "Ditolak") {
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

	pengajuan := []Pengajuan{}
	for rows.Next() {
		var p Pengajuan
		if err := rows.Scan(&p.ID, &p.Judul, &p.Isi, &p.DokumenURL, &p.Status, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email, &p.Lokasi, &p.Tanggal, &p.Keterangan, &p.CreatedAt, &p.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		pengajuan = append(pengajuan, p)
	}

	c.JSON(http.StatusOK, gin.H{"pengajuan": pengajuan})
}

// getPengajuanHandler godoc
// @Summary      Get a request
// @Description  Get a single request by its ID
// @Tags         pengajuan
// @Produce      json
// @Param        id   path      int  true  "Request ID"
// @Success      200  {object}  Pengajuan
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /pengajuan/{id} [get]

func getPengajuanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var p Pengajuan
	err = DB.QueryRow(
		"SELECT id, judul, isi, dokumen_url, status, kategori, nama, nomor_telp, email, lokasi, tanggal, COALESCE(keterangan,''), created_at, updated_at FROM pengajuan WHERE id = $1", id,
	).Scan(&p.ID, &p.Judul, &p.Isi, &p.DokumenURL, &p.Status, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email, &p.Lokasi, &p.Tanggal, &p.Keterangan, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

// createPengajuanHandler godoc
// @Summary      Create request
// @Description  Create a new request
// @Tags         pengajuan
// @Accept       json
// @Produce      json
// @Param        pengajuan  body      PengajuanInput  true  "Request Data"
// @Success      201      {object}  Pengajuan
// @Failure      400      {object}  map[string]string
// @Router       /pengajuan [post]

func createPengajuanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input PengajuanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Kategori == "" {
		input.Kategori = "Umum"
	}

	var p Pengajuan
	err := DB.QueryRow(
		`INSERT INTO pengajuan (judul, isi, dokumen_url, status, kategori, nama, nomor_telp, email, lokasi, tanggal)
		 VALUES ($1, $2, $3, 'Baru', $4, $5, $6, $7, $8, $9)
		 RETURNING id, judul, isi, dokumen_url, status, kategori, nama, nomor_telp, email, lokasi, tanggal, COALESCE(keterangan,''), created_at, updated_at`,
		input.Judul, input.Isi, input.DokumenURL, input.Kategori, input.Nama, input.NomorTelp, input.Email, input.Lokasi, input.Tanggal,
	).Scan(&p.ID, &p.Judul, &p.Isi, &p.DokumenURL, &p.Status, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email, &p.Lokasi, &p.Tanggal, &p.Keterangan, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

// updatePengajuanStatusHandler godoc
// @Summary      Update Pengajuan Status
// @Description  Update status and keterangan of a pengajuan
// @Tags         pengajuan
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id        path      int                         true  "Pengajuan ID"
// @Param        status    body      map[string]string           true  "Status and optional Keterangan"
// @Success      200       {object}  Pengajuan
// @Failure      400       {object}  map[string]string
// @Failure      404       {object}  map[string]string
// @Router       /pengajuan/{id}/status [patch]

func updatePengajuanStatusHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input map[string]string
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, ok := input["status"]
	if !ok || status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	keterangan := input["keterangan"]

	var p Pengajuan
	err = DB.QueryRow(
		`UPDATE pengajuan SET status = $1, keterangan = $2 WHERE id = $3
		 RETURNING id, judul, isi, dokumen_url, status, kategori, nama, nomor_telp, email, lokasi, tanggal, COALESCE(keterangan,''), created_at, updated_at`,
		status, keterangan, id,
	).Scan(&p.ID, &p.Judul, &p.Isi, &p.DokumenURL, &p.Status, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email, &p.Lokasi, &p.Tanggal, &p.Keterangan, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengajuan not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}
