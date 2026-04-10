package main

import (
	"fmt"
	"net/http"
	"strings"

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

	stats := gin.H{
		"gender":            make(map[string]int),
		"religion":          make(map[string]int),
		"education":         make(map[string]int),
		"job":               make(map[string]int),
		"marriage":          make(map[string]int),
		"dusun":             make(map[string]int),
		"age_range":         make(map[string]int),
		"gender_by_dusun":   make(map[string]map[string]int),
		"religion_by_dusun": make(map[string]map[string]int),
		"age_by_dusun":      make(map[string]map[string]int),
	}

	// Helper to run group by counts with normalization
	runQuery := func(field string, category string, target map[string]int, targetByDusun map[string]map[string]int) {
		query := fmt.Sprintf("SELECT COALESCE(%s, 'Tidak Diketahui'), COALESCE(alamat, 'Tidak Diketahui'), COUNT(*) FROM penduduk WHERE dataset_id = $1 GROUP BY %s, alamat", field, field)
		rows, err := DB.Query(query, datasetID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var key string
				var dusunRaw string
				var count int
				if err := rows.Scan(&key, &dusunRaw, &count); err == nil {
					// Normalize the key before adding to target
					cleanKey := normalizeStatValue(category, key)
					cleanDusun := normalizeStatValue("dusun", dusunRaw)
                    
					target[cleanKey] += count
                    
					if targetByDusun != nil {
						if targetByDusun[cleanKey] == nil {
							targetByDusun[cleanKey] = make(map[string]int)
						}
						targetByDusun[cleanKey][cleanDusun] += count
					}
				}
			}
		}
	}

	runQuery("jenis_kelamin", "gender", stats["gender"].(map[string]int), stats["gender_by_dusun"].(map[string]map[string]int))
	runQuery("agama", "religion", stats["religion"].(map[string]int), stats["religion_by_dusun"].(map[string]map[string]int))
	runQuery("pend_terakhir", "education", stats["education"].(map[string]int), nil)
	runQuery("pekerjaan", "job", stats["job"].(map[string]int), nil)
	runQuery("status_kawin", "marriage", stats["marriage"].(map[string]int), nil)
	runQuery("alamat", "dusun", stats["dusun"].(map[string]int), nil)

	// Age Range Calculation
	ageRows, err := DB.Query(`
		SELECT 
			CASE 
				WHEN DATE_PART('year', AGE(tanggal_lahir)) <= 5 THEN '0-5'
				WHEN DATE_PART('year', AGE(tanggal_lahir)) <= 12 THEN '6-12'
				WHEN DATE_PART('year', AGE(tanggal_lahir)) <= 17 THEN '13-17'
				WHEN DATE_PART('year', AGE(tanggal_lahir)) <= 59 THEN '18-59'
				ELSE '60+'
			END as range,
            COALESCE(alamat, 'Tidak Diketahui'),
			COUNT(*) 
		FROM penduduk WHERE dataset_id = $1 GROUP BY range, alamat`, datasetID)

	if err == nil {
		defer ageRows.Close()
		target := stats["age_range"].(map[string]int)
		targetByDusun := stats["age_by_dusun"].(map[string]map[string]int)
        
		for ageRows.Next() {
			var rangeKey string
			var dusunRaw string
			var count int
			if err := ageRows.Scan(&rangeKey, &dusunRaw, &count); err == nil {
				target[rangeKey] += count
                
				cleanDusun := normalizeStatValue("dusun", dusunRaw)
				if targetByDusun[rangeKey] == nil {
					targetByDusun[rangeKey] = make(map[string]int)
				}
				targetByDusun[rangeKey][cleanDusun] += count
			}
		}
	}

	c.JSON(http.StatusOK, stats)
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
