package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// PendudukDataset represents a snapshot for a specific year
type PendudukDataset struct {
	ID           int    `json:"id"`
	Tahun        int    `json:"tahun"`
	NamaFile     string `json:"nama_file"`
	TotalRecords int    `json:"total_records"`
	CreatedAt    string `json:"created_at"`
}

// Penduduk represents a single resident record linked to a dataset
type Penduduk struct {
	ID                int    `json:"id"`
	DatasetID         int    `json:"dataset_id"`
	NIK               string `json:"nik"`
	NoKK              string `json:"no_kk"`
	Nama              string `json:"nama"`
	JenisKelamin      string `json:"jenis_kelamin"`
	StatusKawin       string `json:"status_kawin"`
	TempatLahir       string `json:"tempat_lahir"`
	TanggalLahir      string `json:"tanggal_lahir"`
	Agama             string `json:"agama"`
	PendTerakhir      string `json:"pend_terakhir"`
	Pekerjaan         string `json:"pekerjaan"`
	BisaBaca          bool   `json:"bisa_baca"`
	Kewarganegaraan   string `json:"kewarganegaraan"`
	Alamat            string `json:"alamat"`
	KedudukanKeluarga string `json:"kedudukan_keluarga"`
}

// --- DATASET HANDLERS ---

// ListDatasetsHandler returns all years/snapshots
// listDatasetsHandler godoc
// @Summary      List Population Datasets
// @Description  Get a list of all population snapshots (years)
// @Tags         penduduk
// @Produce      json
// @Success      200  {array}   PendudukDataset
// @Failure      500  {object}  map[string]string
// @Router       /penduduk/datasets [get]
func listDatasetsHandler(c *gin.Context) {
	rows, err := DB.Query("SELECT id, tahun, nama_file, total_records, created_at FROM penduduk_datasets ORDER BY tahun DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []PendudukDataset
	for rows.Next() {
		var d PendudukDataset
		err := rows.Scan(&d.ID, &d.Tahun, &d.NamaFile, &d.TotalRecords, &d.CreatedAt)
		if err == nil {
			list = append(list, d)
		}
	}
	c.JSON(http.StatusOK, list)
}

// CreateDatasetHandler initializes a new year snapshot
// createDatasetHandler godoc
// @Summary      Create Population Dataset
// @Description  Initialize a new population year snapshot (Admin only)
// @Tags         penduduk
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        dataset  body      PendudukDataset  true  "Dataset Data"
// @Success      201      {object}  PendudukDataset
// @Failure      400      {object}  map[string]string
// @Router       /penduduk/datasets [post]
func createDatasetHandler(c *gin.Context) {
	var d PendudukDataset
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := DB.QueryRow(
		"INSERT INTO penduduk_datasets (tahun, nama_file) VALUES ($1, $2) RETURNING id",
		d.Tahun, d.NamaFile,
	).Scan(&d.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dataset: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, d)
}

// DeleteDatasetHandler removes a whole year's data
// deleteDatasetHandler godoc
// @Summary      Delete Population Dataset
// @Description  Remove a whole year's population data (Admin only)
// @Tags         penduduk
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "Dataset ID"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /penduduk/datasets/{id} [delete]
func deleteDatasetHandler(c *gin.Context) {
	id := c.Param("id")
	_, err := DB.Exec("DELETE FROM penduduk_datasets WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// --- RESIDENT RECORD HANDLERS ---

// ListPendudukByDatasetHandler returns all residents for a specific year
// listPendudukByDatasetHandler godoc
// @Summary      List Residents by Dataset
// @Description  Get all resident records for a specific year snapshot
// @Tags         penduduk
// @Produce      json
// @Param        id   path      int  true  "Dataset ID"
// @Success      200  {object}  map[string][]Penduduk
// @Failure      500  {object}  map[string]string
// @Router       /penduduk/datasets/{id}/records [get]
func listPendudukByDatasetHandler(c *gin.Context) {
	datasetID := c.Param("id")

	query := `SELECT id, dataset_id, nik, no_kk, nama, jenis_kelamin, status_kawin, tempat_lahir, tanggal_lahir, agama, pend_terakhir, pekerjaan, bisa_baca, kewarganegaraan, alamat, kedudukan_keluarga 
	          FROM penduduk WHERE dataset_id = $1 ORDER BY id ASC`

	rows, err := DB.Query(query, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []Penduduk
	for rows.Next() {
		var p Penduduk
		err := rows.Scan(
			&p.ID, &p.DatasetID, &p.NIK, &p.NoKK, &p.Nama, &p.JenisKelamin, &p.StatusKawin,
			&p.TempatLahir, &p.TanggalLahir, &p.Agama, &p.PendTerakhir,
			&p.Pekerjaan, &p.BisaBaca, &p.Kewarganegaraan, &p.Alamat, &p.KedudukanKeluarga,
		)
		if err == nil {
			list = append(list, p)
		}
	}
	c.JSON(http.StatusOK, gin.H{"penduduk": list})
}

// GetPendudukStatsHandler returns aggregated data for charts
// getPendudukStatsHandler godoc
// @Summary      Get Population Statistics
// @Description  Get aggregated population data for charts and graphs
// @Tags         penduduk
// @Produce      json
// @Param        id   path      int  true  "Dataset ID"
// @Success      200  {object}  object
// @Router       /penduduk/datasets/{id}/stats [get]
func getPendudukStatsHandler(c *gin.Context) {
	datasetID := c.Param("id")

	rows, err := DB.Query(`
		SELECT id, jenis_kelamin, tanggal_lahir, agama, pend_terakhir, pekerjaan, alamat, COALESCE(no_kk, ''), COALESCE(nik, ''), COALESCE(status_kawin, '')
		FROM penduduk WHERE dataset_id = $1`, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type TempPenduduk struct {
		ID           int
		JenisKelamin string
		TanggalLahir string
		Agama        string
		PendTerakhir string
		Pekerjaan    string
		Alamat       string
		NoKK         string
		NIK          string
		StatusKawin  string
	}

	var list []TempPenduduk
	for rows.Next() {
		var p TempPenduduk
		if err := rows.Scan(&p.ID, &p.JenisKelamin, &p.TanggalLahir, &p.Agama, &p.PendTerakhir, &p.Pekerjaan, &p.Alamat, &p.NoKK, &p.NIK, &p.StatusKawin); err == nil {
			list = append(list, p)
		}
	}

	// 1. Base maps for Overview dashboard page compatibility
	genderMap := make(map[string]int)
	religionMap := make(map[string]int)
	educationMap := make(map[string]int)
	jobMap := make(map[string]int)
	marriageMap := make(map[string]int)
	dusunMap := make(map[string]int)
	ageRangeMap := make(map[string]int)
	genderByDusun := make(map[string]map[string]int)
	religionByDusun := make(map[string]map[string]int)
	ageByDusun := make(map[string]map[string]int)

	var perempuanCount int
	var lakiLakiCount int

	for _, p := range list {
		cleanJK := normalizeStatValue("gender", p.JenisKelamin)
		cleanAgama := normalizeStatValue("religion", p.Agama)
		cleanPend := normalizeStatValue("education", p.PendTerakhir)
		cleanPek := normalizeStatValue("job", p.Pekerjaan)
		cleanKawin := normalizeStatValue("marriage", p.StatusKawin)
		cleanDusun := normalizeStatValue("dusun", p.Alamat)

		if cleanJK == "Perempuan" {
			perempuanCount++
		} else if cleanJK == "Laki-laki" {
			lakiLakiCount++
		}

		var ageRangeKey string
		t, err := parseBirthDate(p.TanggalLahir)
		if err == nil {
			age := calculateAgeAtYear(t, time.Now().Year())
			if age <= 5 {
				ageRangeKey = "0-5"
			} else if age <= 12 {
				ageRangeKey = "6-12"
			} else if age <= 17 {
				ageRangeKey = "13-17"
			} else if age <= 59 {
				ageRangeKey = "18-59"
			} else {
				ageRangeKey = "60+"
			}
		} else {
			ageRangeKey = "Tidak Diketahui"
		}

		genderMap[cleanJK]++
		religionMap[cleanAgama]++
		educationMap[cleanPend]++
		jobMap[cleanPek]++
		marriageMap[cleanKawin]++
		dusunMap[cleanDusun]++
		ageRangeMap[ageRangeKey]++

		if genderByDusun[cleanJK] == nil {
			genderByDusun[cleanJK] = make(map[string]int)
		}
		genderByDusun[cleanJK][cleanDusun]++

		if religionByDusun[cleanAgama] == nil {
			religionByDusun[cleanAgama] = make(map[string]int)
		}
		religionByDusun[cleanAgama][cleanDusun]++

		if ageByDusun[ageRangeKey] == nil {
			ageByDusun[ageRangeKey] = make(map[string]int)
		}
		ageByDusun[ageRangeKey][cleanDusun]++
	}

	// 2. Pre-aggregations for public frontend charts
	uniqueKKs := make(map[string]bool)
	for _, p := range list {
		key := strings.TrimSpace(p.NoKK)
		if key == "" {
			key = strings.TrimSpace(p.NIK)
		}
		if key == "" {
			key = fmt.Sprintf("%d", p.ID)
		}
		uniqueKKs[key] = true
	}
	kepalaKeluargaCount := len(uniqueKKs)

	ageBuckets := []string{"0-4", "5-9", "10-14", "15-19", "20-24", "25-29", "30-34", "35-39", "40-44", "45-49", "50-54", "55-59", "60-64", "65-69", "70-74", "75+"}
	pyramidMap := make(map[string]map[string]int)
	for _, b := range ageBuckets {
		pyramidMap[b] = map[string]int{"lakiLaki": 0, "perempuan": 0}
	}

	for _, p := range list {
		t, err := parseBirthDate(p.TanggalLahir)
		if err == nil {
			age := calculateAgeAtYear(t, time.Now().Year())
			bucket := getAgeBucket(age)
			cleanJK := normalizeStatValue("gender", p.JenisKelamin)
			if cleanJK == "Perempuan" {
				pyramidMap[bucket]["perempuan"]++
			} else if cleanJK == "Laki-laki" {
				pyramidMap[bucket]["lakiLaki"]++
			}
		}
	}

	var agePyramidList []gin.H
	for _, b := range ageBuckets {
		agePyramidList = append(agePyramidList, gin.H{
			"usia":      b,
			"lakiLaki":  pyramidMap[b]["lakiLaki"],
			"perempuan": pyramidMap[b]["perempuan"],
		})
	}

	var dusunList []gin.H
	for name, count := range dusunMap {
		dusunList = append(dusunList, gin.H{
			"name":       name,
			"population": count,
		})
	}

	educationBuckets := []string{"Belum/Tidak Sekolah", "SD Sederajat", "SMP Sederajat", "SMA Sederajat", "D3", "D4", "S1", "Lainnya", "Tidak Diketahui"}
	var pendidikanList []gin.H
	for _, b := range educationBuckets {
		pendidikanList = append(pendidikanList, gin.H{
			"name":  b,
			"value": educationMap[b],
		})
	}

	jobBuckets := []string{"Belum/Tidak Bekerja", "IRT", "Pelajar/Mahasiswa", "Petani", "Wiraswasta", "ASN/TNI/POLRI", "Perangkat Desa", "Pensiunan", "Tukang", "Sopir", "Honorer", "Karyawan", "Lainnya"}
	var pekerjaanList []gin.H
	for _, b := range jobBuckets {
		pekerjaanList = append(pekerjaanList, gin.H{
			"name":  b,
			"value": jobMap[b],
		})
	}

	religionBuckets := []string{"Islam", "Kristen", "Katolik", "Hindu", "Budha", "Konghucu", "Tidak Diketahui"}
	var religionList []gin.H
	for _, b := range religionBuckets {
		religionList = append(religionList, gin.H{
			"name":  b,
			"value": religionMap[b],
		})
	}

	currentYear := time.Now().Year()
	forecastYears := []int{currentYear, currentYear + 1, currentYear + 5}
	var wajibPilihList []gin.H
	for _, yr := range forecastYears {
		count := 0
		for _, p := range list {
			t, err := parseBirthDate(p.TanggalLahir)
			if err == nil {
				age := calculateAgeAtYear(t, yr)
				if age >= 17 {
					count++
				}
			}
		}
		wajibPilihList = append(wajibPilihList, gin.H{
			"year":  fmt.Sprintf("%d", yr),
			"value": count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"gender":            genderMap,
		"religion":          religionMap,
		"education":         educationMap,
		"job":               jobMap,
		"marriage":          marriageMap,
		"dusun":             dusunMap,
		"age_range":         ageRangeMap,
		"gender_by_dusun":   genderByDusun,
		"religion_by_dusun": religionByDusun,
		"age_by_dusun":      ageByDusun,

		"total_penduduk":  len(list),
		"kepala_keluarga": kepalaKeluargaCount,
		"perempuan":       perempuanCount,
		"laki_laki":       lakiLakiCount,
		"age_pyramid":     agePyramidList,
		"dusun_list":       dusunList,
		"pendidikan":      pendidikanList,
		"pekerjaan":       pekerjaanList,
		"agama":           religionList,
		"wajib_pilih":     wajibPilihList,
	})
}

// PatchPendudukHandler handles inline cell updates (Excel-like)
// patchPendudukHandler godoc
// @Summary      Patch Resident Record
// @Description  Update specific fields of a resident record (Admin only)
// @Tags         penduduk
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id       path      int     true  "Resident ID"
// @Param        updates  body      object  true  "Field Updates"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Router       /penduduk/records/{id} [patch]
func patchPendudukHandler(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Dynamic query building for PATCH
	query := "UPDATE penduduk SET "
	var params []interface{}
	i := 1
	for field, value := range updates {
		// Basic security: only allow specific fields to be patched
		allowed := map[string]bool{
			"nik": true, "no_kk": true, "nama": true, "jenis_kelamin": true,
			"status_kawin": true, "tempat_lahir": true, "tanggal_lahir": true,
			"agama": true, "pend_terakhir": true, "pekerjaan": true,
			"bisa_baca": true, "kewarganegaraan": true, "alamat": true,
			"kedudukan_keluarga": true,
		}
		if !allowed[field] {
			continue
		}

		query += fmt.Sprintf("%s = $%d, ", field, i)
		params = append(params, value)
		i++
	}
	query = strings.TrimSuffix(query, ", ")
	query += fmt.Sprintf(" WHERE id = $%d", i)
	params = append(params, id)

	_, err := DB.Exec(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// BulkCreatePendudukHandler handles bulk insertion into a dataset
// bulkCreatePendudukHandler godoc
// @Summary      Bulk Create Residents
// @Description  Mass insert or update resident records for a dataset (Admin only)
// @Tags         penduduk
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id    path      int         true  "Dataset ID"
// @Param        list  body      []Penduduk  true  "List of residents"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Router       /penduduk/datasets/{id}/bulk [post]
func bulkCreatePendudukHandler(c *gin.Context) {
	datasetID := c.Param("id")
	var list []Penduduk
	if err := c.ShouldBindJSON(&list); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}
	defer tx.Rollback()

	// SQL with UPSERT capability
	sqlQuery := `INSERT INTO penduduk (dataset_id, nik, no_kk, nama, jenis_kelamin, status_kawin, tempat_lahir, tanggal_lahir, agama, pend_terakhir, pekerjaan, bisa_baca, kewarganegaraan, alamat, kedudukan_keluarga) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			  ON CONFLICT (nik, dataset_id) DO UPDATE SET
			  no_kk = EXCLUDED.no_kk,
			  nama = EXCLUDED.nama,
			  jenis_kelamin = EXCLUDED.jenis_kelamin,
			  status_kawin = EXCLUDED.status_kawin,
			  tempat_lahir = EXCLUDED.tempat_lahir,
			  tanggal_lahir = EXCLUDED.tanggal_lahir,
			  agama = EXCLUDED.agama,
			  pend_terakhir = EXCLUDED.pend_terakhir,
			  pekerjaan = EXCLUDED.pekerjaan,
			  bisa_baca = EXCLUDED.bisa_baca,
			  kewarganegaraan = EXCLUDED.kewarganegaraan,
			  alamat = EXCLUDED.alamat,
			  kedudukan_keluarga = EXCLUDED.kedudukan_keluarga`

	stmt, err := tx.Prepare(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare statement: " + err.Error()})
		return
	}
	defer stmt.Close()

	count := 0
	for _, p := range list {
		_, err := stmt.Exec(
			datasetID, p.NIK, p.NoKK, p.Nama, p.JenisKelamin, p.StatusKawin, p.TempatLahir,
			p.TanggalLahir, p.Agama, p.PendTerakhir, p.Pekerjaan, p.BisaBaca,
			p.Kewarganegaraan, p.Alamat, p.KedudukanKeluarga,
		)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error at NIK %s: %v", p.NIK, err)})
			return
		}
		count++
	}

	// Update the total_records in dataset table
	_, _ = tx.Exec("UPDATE penduduk_datasets SET total_records = (SELECT COUNT(*) FROM penduduk WHERE dataset_id = $1) WHERE id = $1", datasetID)

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Imported %d records", count), "count": count})
}

// DeleteRecordHandler removes a single resident record
// deleteRecordHandler godoc
// @Summary      Delete Resident Record
// @Description  Remove a single resident record (Admin only)
// @Tags         penduduk
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "Resident ID"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /penduduk/records/{id} [delete]
func deleteRecordHandler(c *gin.Context) {
	id := c.Param("id")
	_, err := DB.Exec("DELETE FROM penduduk WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// normalizeStatValue cleans up messy user input into standardized categories
func normalizeStatValue(category string, value string) string {
	val := strings.ToLower(strings.TrimSpace(value))
	if val == "" || val == "tidak diketahui" || val == "-" {
		return "Tidak Diketahui"
	}

	switch category {
	case "gender":
		if strings.HasPrefix(val, "l") {
			return "Laki-laki"
		}
		if strings.HasPrefix(val, "p") {
			return "Perempuan"
		}

	case "education":
		// Tamat checks first (more specific)
		if strings.Contains(val, "tamat sd") || (strings.Contains(val, "sd") && !strings.Contains(val, "blm") && !strings.Contains(val, "tdk")) {
			return "SD Sederajat"
		}
		if strings.Contains(val, "sltp") || strings.Contains(val, "smp") {
			return "SMP Sederajat"
		}
		if strings.Contains(val, "slta") || strings.Contains(val, "sma") {
			return "SMA Sederajat"
		}
		if strings.Contains(val, "strata 1") || strings.Contains(val, "strata i") || strings.Contains(val, "s1") {
			return "S1"
		}
		if strings.Contains(val, "diploma iii") || strings.Contains(val, "diplom iii") || strings.Contains(val, "d3") {
			return "D3"
		}
		if strings.Contains(val, "diploma iv") || strings.Contains(val, "d4") {
			return "D4"
		}
		// Catch-all last
		return "Belum/Tidak Sekolah"

	case "job":
		// Specific broad categories first
		if strings.Contains(val, "tani") || strings.Contains(val, "peta") || strings.Contains(val, "petn") ||
			strings.Contains(val, "petr") || strings.Contains(val, "peten") {
			return "Petani"
		}
		if strings.Contains(val, "wasta") || strings.Contains(val, "wiras") {
			return "Wiraswasta"
		}
		if strings.Contains(val, "pns") || strings.Contains(val, "p3k") || strings.Contains(val, "abri") ||
			strings.Contains(val, "polri") || strings.Contains(val, "asn") {
			return "ASN/TNI/POLRI"
		}
		if strings.Contains(val, "irt") || strings.Contains(val, "rumah tangga") {
			return "IRT"
		}
		if strings.Contains(val, "pelaj") || strings.Contains(val, "pelej") || strings.Contains(val, "peljar") ||
			strings.Contains(val, "mahasiswa") || strings.Contains(val, "mahasiswi") || strings.Contains(val, "mahaisiswi") {
			return "Pelajar/Mahasiswa"
		}
		if strings.Contains(val, "blm") || strings.Contains(val, "tdk") || strings.Contains(val, "tidak") ||
			strings.Contains(val, "belum") || strings.Contains(val, "kerja") {
			return "Belum/Tidak Bekerja"
		}

		// Specific roles
		if strings.Contains(val, "desa") {
			return "Perangkat Desa"
		}
		if strings.Contains(val, "pensiun") {
			return "Pensiunan"
		}
		if strings.Contains(val, "tukang") {
			return "Tukang"
		}
		if strings.Contains(val, "sopir") || strings.Contains(val, "supir") {
			return "Sopir"
		}
		if strings.Contains(val, "honorer") {
			return "Honorer"
		}
		if strings.Contains(val, "karyawan") {
			return "Karyawan"
		}
		if strings.Contains(val, "anak") {
			return "Belum/Tidak Bekerja"
		}

	case "marriage":
		if strings.Contains(val, "cerai hidup") {
			return "Cerai Hidup"
		}
		if strings.Contains(val, "cerai mati") {
			return "Cerai Mati"
		}
		if strings.Contains(val, "cerai") {
			return "Cerai Hidup" // fallback for bare "cerai"
		}
		if strings.Contains(val, "belum") || strings.Contains(val, "blm") {
			return "Belum Kawin"
		}
		if strings.Contains(val, "kawin") || strings.Contains(val, "kwin") {
			return "Kawin"
		}

	case "religion":
		if strings.Contains(val, "islam") || strings.Contains(val, "jslam") {
			return "Islam"
		}
		if strings.Contains(val, "kristen") {
			return "Kristen"
		}
		if strings.Contains(val, "katolik") {
			return "Katolik"
		}

	case "dusun":
		if strings.Contains(val, "dusun 1") {
			return "Dusun 1"
		}
		if strings.Contains(val, "dusun 2") {
			return "Dusun 2"
		}
		if strings.Contains(val, "dusun 3") {
			return "Dusun 3"
		}
		if strings.Contains(val, "dusun 4") {
			return "Dusun 4"
		}
		if strings.Contains(val, "dusun 5") {
			return "Dusun 5"
		}
	}

	// Fallback: capitalize first letter
	if len(val) > 0 {
		return strings.ToUpper(val[0:1]) + val[1:]
	}
	return "Tidak Diketahui"
}

func parseBirthDate(tgl string) (time.Time, error) {
	tgl = strings.TrimSpace(tgl)
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	for _, f := range formats {
		t, err := time.Parse(f, tgl)
		if err == nil {
			return t, nil
		}
	}
	if strings.Contains(tgl, "T") {
		parts := strings.Split(tgl, "T")
		t, err := time.Parse("2006-01-02", parts[0])
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func calculateAgeAtYear(birthDate time.Time, year int) int {
	age := year - birthDate.Year()
	if age < 0 {
		return 0
	}
	return age
}

func getAgeBucket(age int) string {
	if age <= 4 {
		return "0-4"
	}
	if age <= 9 {
		return "5-9"
	}
	if age <= 14 {
		return "10-14"
	}
	if age <= 19 {
		return "15-19"
	}
	if age <= 24 {
		return "20-24"
	}
	if age <= 29 {
		return "25-29"
	}
	if age <= 34 {
		return "30-34"
	}
	if age <= 39 {
		return "35-39"
	}
	if age <= 44 {
		return "40-44"
	}
	if age <= 49 {
		return "45-49"
	}
	if age <= 54 {
		return "50-54"
	}
	if age <= 59 {
		return "55-59"
	}
	if age <= 64 {
		return "60-64"
	}
	if age <= 69 {
		return "65-69"
	}
	if age <= 74 {
		return "70-74"
	}
	return "75+"
}

