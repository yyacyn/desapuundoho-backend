package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Stunting struct {
	ID                 int      `json:"id"`
	NikAnak            string   `json:"nik_anak"`
	NamaAnak           string   `json:"nama_anak"`
	LokasiDusun        string   `json:"lokasi_dusun"`
	Dusun              string   `json:"dusun"`
	TanggalLahir       string   `json:"tanggal_lahir"`
	UmurBulan          int      `json:"umur_bulan"`
	TinggiBadan        *float64 `json:"tinggi_badan,omitempty"`
	BeratBadan         *float64 `json:"berat_badan,omitempty"`
	Status             string   `json:"status"`
	TanggalPemeriksaan string   `json:"tanggal_pemeriksaan"`
	CreatedAt          string   `json:"created_at"`
}

type StuntingInput struct {
	NikAnak            string `json:"nik_anak"`
	NamaAnak           string `json:"nama_anak" binding:"required"`
	LokasiDusun        string `json:"lokasi_dusun" binding:"required"`
	TanggalLahir       string `json:"tanggal_lahir"`
	TinggiBadan        string `json:"tinggi_badan"`
	BeratBadan         string `json:"berat_badan"`
	Status             string `json:"status"`
	TanggalPemeriksaan string `json:"tanggal_pemeriksaan"`
}

func listStuntingHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	dusun := strings.TrimSpace(c.Query("dusun"))

	query := `SELECT id, COALESCE(nik_anak,''), nama_anak, lokasi_dusun,
		COALESCE(TO_CHAR(tanggal_lahir, 'YYYY-MM-DD'), ''), tinggi_badan, berat_badan,
		status, COALESCE(TO_CHAR(tanggal_pemeriksaan, 'YYYY-MM-DD'), ''),
		COALESCE(TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')
		FROM stunting`
	args := []interface{}{}
	conditions := []string{}

	if search != "" {
		idx := len(args) + 1
		conditions = append(conditions, fmt.Sprintf("(nama_anak ILIKE $%d OR lokasi_dusun ILIKE $%d OR COALESCE(nik_anak,'') ILIKE $%d OR status ILIKE $%d)", idx, idx, idx, idx))
		args = append(args, "%"+search+"%")
	}

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, status)
	}

	if dusun != "" {
		conditions = append(conditions, fmt.Sprintf("lokasi_dusun = $%d", len(args)+1))
		args = append(args, dusun)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []Stunting{}
	for rows.Next() {
		item, err := scanStuntingRow(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, item)
	}

	c.JSON(http.StatusOK, gin.H{"stunting": list})
}

func getStuntingHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	row := DB.QueryRow(
		`SELECT id, COALESCE(nik_anak,''), nama_anak, lokasi_dusun,
		COALESCE(TO_CHAR(tanggal_lahir, 'YYYY-MM-DD'), ''), tinggi_badan, berat_badan,
		status, COALESCE(TO_CHAR(tanggal_pemeriksaan, 'YYYY-MM-DD'), ''),
		COALESCE(TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')
		FROM stunting WHERE id = $1`,
		id,
	)

	item, err := scanStuntingRow(row)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data stunting tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

func createStuntingHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input StuntingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "" {
		input.Status = "Normal"
	}

	var tanggalLahir interface{}
	if strings.TrimSpace(input.TanggalLahir) == "" {
		tanggalLahir = nil
	} else {
		tanggalLahir = input.TanggalLahir
	}

	tinggiBadan, err := parseOptionalFloat(input.TinggiBadan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tinggi_badan harus berupa angka"})
		return
	}

	beratBadan, err := parseOptionalFloat(input.BeratBadan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "berat_badan harus berupa angka"})
		return
	}

	var tanggalPemeriksaan interface{}
	if strings.TrimSpace(input.TanggalPemeriksaan) == "" {
		tanggalPemeriksaan = nil
	} else {
		tanggalPemeriksaan = input.TanggalPemeriksaan
	}

	var created Stunting
	err = DB.QueryRow(
		`INSERT INTO stunting (nik_anak, nama_anak, lokasi_dusun, tanggal_lahir, tinggi_badan, berat_badan, status, tanggal_pemeriksaan)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, COALESCE(nik_anak,''), nama_anak, lokasi_dusun,
		 COALESCE(TO_CHAR(tanggal_lahir, 'YYYY-MM-DD'), ''), tinggi_badan, berat_badan,
		 status, COALESCE(TO_CHAR(tanggal_pemeriksaan, 'YYYY-MM-DD'), ''),
		 COALESCE(TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')`,
		input.NikAnak,
		input.NamaAnak,
		input.LokasiDusun,
		tanggalLahir,
		tinggiBadan,
		beratBadan,
		input.Status,
		tanggalPemeriksaan,
	).Scan(
		&created.ID,
		&created.NikAnak,
		&created.NamaAnak,
		&created.LokasiDusun,
		&created.TanggalLahir,
		&created.TinggiBadan,
		&created.BeratBadan,
		&created.Status,
		&created.TanggalPemeriksaan,
		&created.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created.Dusun = created.LokasiDusun
	created.UmurBulan = calculateAgeMonths(created.TanggalLahir)

	c.JSON(http.StatusCreated, created)
}

func updateStuntingHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input StuntingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "" {
		input.Status = "Normal"
	}

	var tanggalLahir interface{}
	if strings.TrimSpace(input.TanggalLahir) == "" {
		tanggalLahir = nil
	} else {
		tanggalLahir = input.TanggalLahir
	}

	tinggiBadan, err := parseOptionalFloat(input.TinggiBadan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tinggi_badan harus berupa angka"})
		return
	}

	beratBadan, err := parseOptionalFloat(input.BeratBadan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "berat_badan harus berupa angka"})
		return
	}

	var tanggalPemeriksaan interface{}
	if strings.TrimSpace(input.TanggalPemeriksaan) == "" {
		tanggalPemeriksaan = nil
	} else {
		tanggalPemeriksaan = input.TanggalPemeriksaan
	}

	var updated Stunting
	err = DB.QueryRow(
		`UPDATE stunting
		 SET nik_anak = $1, nama_anak = $2, lokasi_dusun = $3, tanggal_lahir = $4,
		 tinggi_badan = $5, berat_badan = $6, status = $7, tanggal_pemeriksaan = $8
		 WHERE id = $9
		 RETURNING id, COALESCE(nik_anak,''), nama_anak, lokasi_dusun,
		 COALESCE(TO_CHAR(tanggal_lahir, 'YYYY-MM-DD'), ''), tinggi_badan, berat_badan,
		 status, COALESCE(TO_CHAR(tanggal_pemeriksaan, 'YYYY-MM-DD'), ''),
		 COALESCE(TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')`,
		input.NikAnak,
		input.NamaAnak,
		input.LokasiDusun,
		tanggalLahir,
		tinggiBadan,
		beratBadan,
		input.Status,
		tanggalPemeriksaan,
		id,
	).Scan(
		&updated.ID,
		&updated.NikAnak,
		&updated.NamaAnak,
		&updated.LokasiDusun,
		&updated.TanggalLahir,
		&updated.TinggiBadan,
		&updated.BeratBadan,
		&updated.Status,
		&updated.TanggalPemeriksaan,
		&updated.CreatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data stunting tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated.Dusun = updated.LokasiDusun
	updated.UmurBulan = calculateAgeMonths(updated.TanggalLahir)

	c.JSON(http.StatusOK, updated)
}

func deleteStuntingHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result, err := DB.Exec("DELETE FROM stunting WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data stunting tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data stunting berhasil dihapus"})
}

func scanStuntingRow(row scanner) (Stunting, error) {
	var item Stunting
	var tinggi sql.NullFloat64
	var berat sql.NullFloat64
	var tanggalLahir sql.NullString
	var tanggalPemeriksaan sql.NullString
	var createdAt sql.NullString

	err := row.Scan(
		&item.ID,
		&item.NikAnak,
		&item.NamaAnak,
		&item.LokasiDusun,
		&tanggalLahir,
		&tinggi,
		&berat,
		&item.Status,
		&tanggalPemeriksaan,
		&createdAt,
	)
	if err != nil {
		return Stunting{}, err
	}

	item.Dusun = item.LokasiDusun
	item.TanggalLahir = tanggalLahir.String
	item.TanggalPemeriksaan = tanggalPemeriksaan.String
	item.CreatedAt = createdAt.String
	item.UmurBulan = calculateAgeMonths(item.TanggalLahir)

	if tinggi.Valid {
		value := tinggi.Float64
		item.TinggiBadan = &value
	}
	if berat.Valid {
		value := berat.Float64
		item.BeratBadan = &value
	}

	return item, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func parseOptionalFloat(value string) (*float64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func calculateAgeMonths(dateString string) int {
	if strings.TrimSpace(dateString) == "" {
		return 0
	}

	birthDate, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		return 0
	}

	now := time.Now()
	years := now.Year() - birthDate.Year()
	months := int(now.Month() - birthDate.Month())
	ageMonths := years*12 + months
	if now.Day() < birthDate.Day() {
		ageMonths--
	}

	if ageMonths < 0 {
		return 0
	}

	return ageMonths
}
