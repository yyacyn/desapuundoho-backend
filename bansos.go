package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Bansos struct {
	ID                int    `json:"id"`
	NamaProgram       string `json:"nama_program"`
	NikPenerima       string `json:"nik_penerima"`
	NamaPenerima      string `json:"nama_penerima"`
	LokasiDusun       string `json:"lokasi_dusun"`
	Status            string `json:"status"`
	TanggalPenyaluran string `json:"tanggal_penyaluran"`
	Keterangan        string `json:"keterangan"`
	CreatedAt         string `json:"created_at"`
}

type BansosInput struct {
	NamaProgram       string `json:"nama_program" binding:"required"`
	NikPenerima       string `json:"nik_penerima"`
	NamaPenerima      string `json:"nama_penerima" binding:"required"`
	LokasiDusun       string `json:"lokasi_dusun" binding:"required"`
	Status            string `json:"status"`
	TanggalPenyaluran string `json:"tanggal_penyaluran"`
	Keterangan        string `json:"keterangan"`
}

func listBansosHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	status := c.Query("status")
	nik := c.Query("nik")

	query := `SELECT id, nama_program, COALESCE(nik_penerima,''), nama_penerima, lokasi_dusun,
		status, COALESCE(TO_CHAR(tanggal_penyaluran, 'YYYY-MM-DD'), ''), COALESCE(keterangan,''),
		COALESCE(TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')
		FROM bansos`
	args := []interface{}{}
	conditions := []string{}

	if status != "" {
		conditions = append(conditions, "status = $"+strconv.Itoa(len(args)+1))
		args = append(args, status)
	}

	if nik != "" {
		conditions = append(conditions, "nik_penerima = $"+strconv.Itoa(len(args)+1))
		args = append(args, nik)
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []Bansos{}
	for rows.Next() {
		var b Bansos
		if err := rows.Scan(
			&b.ID,
			&b.NamaProgram,
			&b.NikPenerima,
			&b.NamaPenerima,
			&b.LokasiDusun,
			&b.Status,
			&b.TanggalPenyaluran,
			&b.Keterangan,
			&b.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, b)
	}

	c.JSON(http.StatusOK, gin.H{"bansos": list})
}

func getBansosHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var b Bansos
	err = DB.QueryRow(
		`SELECT id, nama_program, COALESCE(nik_penerima,''), nama_penerima, lokasi_dusun,
		status, COALESCE(TO_CHAR(tanggal_penyaluran, 'YYYY-MM-DD'), ''), COALESCE(keterangan,''),
		COALESCE(TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')
		FROM bansos WHERE id = $1`,
		id,
	).Scan(
		&b.ID,
		&b.NamaProgram,
		&b.NikPenerima,
		&b.NamaPenerima,
		&b.LokasiDusun,
		&b.Status,
		&b.TanggalPenyaluran,
		&b.Keterangan,
		&b.CreatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data bansos tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, b)
}

func createBansosHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input BansosInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "" {
		input.Status = "Menunggu"
	}

	var tanggalPenyaluran interface{}
	if input.TanggalPenyaluran == "" {
		tanggalPenyaluran = nil
	} else {
		tanggalPenyaluran = input.TanggalPenyaluran
	}

	var b Bansos
	err := DB.QueryRow(
		`INSERT INTO bansos (nama_program, nik_penerima, nama_penerima, lokasi_dusun, status, tanggal_penyaluran, keterangan)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, nama_program, COALESCE(nik_penerima,''), nama_penerima, lokasi_dusun,
		 status, COALESCE(TO_CHAR(tanggal_penyaluran, 'YYYY-MM-DD'), ''), COALESCE(keterangan,''),
		 COALESCE(TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')`,
		input.NamaProgram,
		input.NikPenerima,
		input.NamaPenerima,
		input.LokasiDusun,
		input.Status,
		tanggalPenyaluran,
		input.Keterangan,
	).Scan(
		&b.ID,
		&b.NamaProgram,
		&b.NikPenerima,
		&b.NamaPenerima,
		&b.LokasiDusun,
		&b.Status,
		&b.TanggalPenyaluran,
		&b.Keterangan,
		&b.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, b)
}

func updateBansosHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input BansosInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "" {
		input.Status = "Menunggu"
	}

	var tanggalPenyaluran interface{}
	if input.TanggalPenyaluran == "" {
		tanggalPenyaluran = nil
	} else {
		tanggalPenyaluran = input.TanggalPenyaluran
	}

	var b Bansos
	err = DB.QueryRow(
		`UPDATE bansos
		 SET nama_program = $1, nik_penerima = $2, nama_penerima = $3, lokasi_dusun = $4,
		 status = $5, tanggal_penyaluran = $6, keterangan = $7
		 WHERE id = $8
		 RETURNING id, nama_program, COALESCE(nik_penerima,''), nama_penerima, lokasi_dusun,
		 status, COALESCE(TO_CHAR(tanggal_penyaluran, 'YYYY-MM-DD'), ''), COALESCE(keterangan,''),
		 COALESCE(TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')`,
		input.NamaProgram,
		input.NikPenerima,
		input.NamaPenerima,
		input.LokasiDusun,
		input.Status,
		tanggalPenyaluran,
		input.Keterangan,
		id,
	).Scan(
		&b.ID,
		&b.NamaProgram,
		&b.NikPenerima,
		&b.NamaPenerima,
		&b.LokasiDusun,
		&b.Status,
		&b.TanggalPenyaluran,
		&b.Keterangan,
		&b.CreatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data bansos tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, b)
}

func deleteBansosHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result, err := DB.Exec("DELETE FROM bansos WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data bansos tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data bansos berhasil dihapus"})
}
