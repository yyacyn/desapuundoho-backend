package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pengaduan struct {
	ID         int    `json:"id"`
	Judul      string `json:"judul"`
	Isi        string `json:"isi"`
	FotoURL    string `json:"foto_url"`
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

type PengaduanInput struct {
	Judul     string `json:"judul"`
	Isi       string `json:"isi"`
	FotoURL   string `json:"foto_url"`
	Kategori  string `json:"kategori"`
	Nama      string `json:"nama"`
	NomorTelp string `json:"nomor_telp"`
	Email     string `json:"email"`
	Lokasi    string `json:"lokasi"`
	Tanggal   string `json:"tanggal"`
}

// Handlers
// listPengaduanHandler godoc
// @Summary      List pengaduan
// @Description  Get a list of complaints. Use ?status=Baru, Ditinjau, Diproses, Selesai, or Ditolak to filter. Leave empty to get all complaints.
// @Tags         pengaduan
// @Accept       json
// @Produce      json
// @Param        status  query     string  false  "Filter by complaint status"
// @Success      200     {object}  map[string][]Pengaduan
// @Failure      503     {object}  map[string]string
// @Router       /pengaduan [get]

func listPengaduanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	status := c.Query("status") // optional filter
	query := "SELECT id, judul, isi, foto_url, status, kategori, nama, nomor_telp, email, lokasi, tanggal, COALESCE(keterangan,''), created_at, updated_at FROM pengaduan"
	args := []interface{}{}

	if status != "" && (status == "Baru" || status == "Ditinjau" || status == "Diproses" || status == "Selesai" || status == "Ditolak") {
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
		if err := rows.Scan(&p.ID, &p.Judul, &p.Isi, &p.FotoURL, &p.Status, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email, &p.Lokasi, &p.Tanggal, &p.Keterangan, &p.CreatedAt, &p.UpdatedAt); err != nil {
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
// @Failure      503  {object}  map[string]string
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
		"SELECT id, judul, isi, foto_url, status, kategori, nama, nomor_telp, email, lokasi, tanggal, COALESCE(keterangan,''), created_at, updated_at FROM pengaduan WHERE id = $1", id,
	).Scan(&p.ID, &p.Judul, &p.Isi, &p.FotoURL, &p.Status, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email, &p.Lokasi, &p.Tanggal, &p.Keterangan, &p.CreatedAt, &p.UpdatedAt)

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
// @Description  Create a new complaint submission
// @Tags         pengaduan
// @Accept       json
// @Produce      json
// @Param        pengaduan  body      PengaduanInput  true  "Complaint Data"
// @Success      201      {object}  Pengaduan
// @Failure      400      {object}  map[string]string
// @Failure      503      {object}  map[string]string
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
		`INSERT INTO pengaduan (judul, isi, foto_url, status, kategori, nama, nomor_telp, email, lokasi, tanggal)
		 VALUES ($1, $2, $3, 'Baru', $4, $5, $6, $7, $8, $9)
		 RETURNING id, judul, isi, foto_url, status, kategori, nama, nomor_telp, email, lokasi, tanggal, COALESCE(keterangan,''), created_at, updated_at`,
		input.Judul, input.Isi, input.FotoURL, input.Kategori, input.Nama, input.NomorTelp, input.Email, input.Lokasi, input.Tanggal,
	).Scan(&p.ID, &p.Judul, &p.Isi, &p.FotoURL, &p.Status, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email, &p.Lokasi, &p.Tanggal, &p.Keterangan, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

// updatePengaduanStatusHandler godoc
// @Summary      Update Pengaduan Status
// @Description  Update status and keterangan of a pengaduan
// @Tags         pengaduan
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id        path      int                         true  "Pengaduan ID"
// @Param        status    body      map[string]string           true  "Status and optional Keterangan"
// @Success      200       {object}  Pengaduan
// @Failure      400       {object}  map[string]string
// @Failure      404       {object}  map[string]string
// @Failure      503       {object}  map[string]string
// @Router       /pengaduan/{id}/status [patch]

func updatePengaduanStatusHandler(c *gin.Context) {
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

	var p Pengaduan
	err = DB.QueryRow(
		`UPDATE pengaduan SET status = $1, keterangan = $2 WHERE id = $3
		 RETURNING id, judul, isi, foto_url, status, kategori, nama, nomor_telp, email, lokasi, tanggal, COALESCE(keterangan,''), created_at, updated_at`,
		status, keterangan, id,
	).Scan(&p.ID, &p.Judul, &p.Isi, &p.FotoURL, &p.Status, &p.Kategori, &p.Nama, &p.NomorTelp, &p.Email, &p.Lokasi, &p.Tanggal, &p.Keterangan, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengaduan not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}
