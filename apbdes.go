package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// --- Types ---

type ApbdDesa struct {
	ID               int    `json:"id"`
	Tahun            int    `json:"tahun"`
	TotalPendapatan  int64  `json:"total_pendapatan"`
	TotalPengeluaran int64  `json:"total_pengeluaran"`
	CreatedAt        string `json:"created_at"`
}

type ApbdDesaInput struct {
	Tahun            int   `json:"tahun" binding:"required"`
	TotalPendapatan  int64 `json:"total_pendapatan"`
	TotalPengeluaran int64 `json:"total_pengeluaran"`
}

type ApbdPendapatan struct {
	ID       int    `json:"id"`
	IdApbd   int    `json:"id_apbd"`
	Kategori string `json:"kategori"`
	Jumlah   int64  `json:"jumlah"`
}

type ApbdPendapatanInput struct {
	IdApbd   int    `json:"id_apbd" binding:"required"`
	Kategori string `json:"kategori" binding:"required"`
	Jumlah   int64  `json:"jumlah" binding:"required"`
}

type ApbdPengeluaran struct {
	ID     int    `json:"id"`
	IdApbd int    `json:"id_apbd"`
	Bidang string `json:"bidang"`
	Jumlah int64  `json:"jumlah"`
}

type ApbdPengeluaranInput struct {
	IdApbd int    `json:"id_apbd" binding:"required"`
	Bidang string `json:"bidang" binding:"required"`
	Jumlah int64  `json:"jumlah" binding:"required"`
}

// --- APBDes Handlers ---

// GET /api/apbdes — list all APBD years
func listApbdesHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	rows, err := DB.Query(`SELECT id, tahun, total_pendapatan, total_pengeluaran, created_at FROM apbd_desa ORDER BY tahun DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []ApbdDesa{}
	for rows.Next() {
		var a ApbdDesa
		if err := rows.Scan(&a.ID, &a.Tahun, &a.TotalPendapatan, &a.TotalPengeluaran, &a.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, a)
	}

	c.JSON(http.StatusOK, gin.H{"apbdes": list})
}

// POST /api/apbdes
func createApbdesHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input ApbdDesaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var a ApbdDesa
	err := DB.QueryRow(
		`INSERT INTO apbd_desa (tahun, total_pendapatan, total_pengeluaran)
		 VALUES ($1, $2, $3)
		 RETURNING id, tahun, total_pendapatan, total_pengeluaran, created_at`,
		input.Tahun, input.TotalPendapatan, input.TotalPengeluaran,
	).Scan(&a.ID, &a.Tahun, &a.TotalPendapatan, &a.TotalPengeluaran, &a.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, a)
}

// DELETE /api/apbdes/:id
func deleteApbdesHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result, err := DB.Exec("DELETE FROM apbd_desa WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "APBD tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "APBD dihapus"})
}

// --- Pendapatan Handlers ---

// GET /api/apbdes/:id/pendapatan
func listPendapatanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	apbdId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid APBD ID"})
		return
	}

	rows, err := DB.Query(`SELECT id, id_apbd, kategori, jumlah FROM apbd_pendapatan WHERE id_apbd = $1 ORDER BY id`, apbdId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []ApbdPendapatan{}
	for rows.Next() {
		var p ApbdPendapatan
		if err := rows.Scan(&p.ID, &p.IdApbd, &p.Kategori, &p.Jumlah); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, p)
	}

	c.JSON(http.StatusOK, gin.H{"pendapatan": list})
}

// POST /api/apbdes/pendapatan
func createPendapatanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input ApbdPendapatanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var p ApbdPendapatan
	err := DB.QueryRow(
		`INSERT INTO apbd_pendapatan (id_apbd, kategori, jumlah)
		 VALUES ($1, $2, $3)
		 RETURNING id, id_apbd, kategori, jumlah`,
		input.IdApbd, input.Kategori, input.Jumlah,
	).Scan(&p.ID, &p.IdApbd, &p.Kategori, &p.Jumlah)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update totals
	updateApbdTotals(input.IdApbd)

	c.JSON(http.StatusCreated, p)
}

// PUT /api/apbdes/pendapatan/:id
func updatePendapatanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input ApbdPendapatanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var p ApbdPendapatan
	err = DB.QueryRow(
		`UPDATE apbd_pendapatan SET kategori=$1, jumlah=$2 WHERE id=$3
		 RETURNING id, id_apbd, kategori, jumlah`,
		input.Kategori, input.Jumlah, id,
	).Scan(&p.ID, &p.IdApbd, &p.Kategori, &p.Jumlah)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pendapatan tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updateApbdTotals(p.IdApbd)
	c.JSON(http.StatusOK, p)
}

// DELETE /api/apbdes/pendapatan/:id
func deletePendapatanHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Get apbd id before deleting
	var apbdId int
	DB.QueryRow("SELECT id_apbd FROM apbd_pendapatan WHERE id = $1", id).Scan(&apbdId)

	result, err := DB.Exec("DELETE FROM apbd_pendapatan WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pendapatan tidak ditemukan"})
		return
	}

	if apbdId > 0 {
		updateApbdTotals(apbdId)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pendapatan dihapus"})
}

// --- Pengeluaran Handlers ---

// GET /api/apbdes/:id/pengeluaran
func listPengeluaranHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	apbdId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid APBD ID"})
		return
	}

	rows, err := DB.Query(`SELECT id, id_apbd, bidang, jumlah FROM apbd_pengeluaran WHERE id_apbd = $1 ORDER BY id`, apbdId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []ApbdPengeluaran{}
	for rows.Next() {
		var p ApbdPengeluaran
		if err := rows.Scan(&p.ID, &p.IdApbd, &p.Bidang, &p.Jumlah); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, p)
	}

	c.JSON(http.StatusOK, gin.H{"pengeluaran": list})
}

// POST /api/apbdes/pengeluaran
func createPengeluaranHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input ApbdPengeluaranInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var p ApbdPengeluaran
	err := DB.QueryRow(
		`INSERT INTO apbd_pengeluaran (id_apbd, bidang, jumlah)
		 VALUES ($1, $2, $3)
		 RETURNING id, id_apbd, bidang, jumlah`,
		input.IdApbd, input.Bidang, input.Jumlah,
	).Scan(&p.ID, &p.IdApbd, &p.Bidang, &p.Jumlah)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updateApbdTotals(input.IdApbd)
	c.JSON(http.StatusCreated, p)
}

// PUT /api/apbdes/pengeluaran/:id
func updatePengeluaranHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input ApbdPengeluaranInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var p ApbdPengeluaran
	err = DB.QueryRow(
		`UPDATE apbd_pengeluaran SET bidang=$1, jumlah=$2 WHERE id=$3
		 RETURNING id, id_apbd, bidang, jumlah`,
		input.Bidang, input.Jumlah, id,
	).Scan(&p.ID, &p.IdApbd, &p.Bidang, &p.Jumlah)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengeluaran tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updateApbdTotals(p.IdApbd)
	c.JSON(http.StatusOK, p)
}

// DELETE /api/apbdes/pengeluaran/:id
func deletePengeluaranHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var apbdId int
	DB.QueryRow("SELECT id_apbd FROM apbd_pengeluaran WHERE id = $1", id).Scan(&apbdId)

	result, err := DB.Exec("DELETE FROM apbd_pengeluaran WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengeluaran tidak ditemukan"})
		return
	}

	if apbdId > 0 {
		updateApbdTotals(apbdId)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pengeluaran dihapus"})
}

// --- Helper ---

// updateApbdTotals recalculates and saves totals from child tables
func updateApbdTotals(apbdId int) {
	var totalP, totalK int64
	DB.QueryRow("SELECT COALESCE(SUM(jumlah),0) FROM apbd_pendapatan WHERE id_apbd=$1", apbdId).Scan(&totalP)
	DB.QueryRow("SELECT COALESCE(SUM(jumlah),0) FROM apbd_pengeluaran WHERE id_apbd=$1", apbdId).Scan(&totalK)
	DB.Exec("UPDATE apbd_desa SET total_pendapatan=$1, total_pengeluaran=$2 WHERE id=$3", totalP, totalK, apbdId)
}
