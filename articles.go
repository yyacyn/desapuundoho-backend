package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Data types
type Article struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Content    string `json:"content"`
	Excerpt    string `json:"excerpt"`
	CoverImage string `json:"cover_image"`
	Author     string `json:"author"`
	Category   string `json:"category"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type ArticleInput struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Excerpt    string `json:"excerpt"`
	CoverImage string `json:"cover_image"`
	Category   string `json:"category"`
	Status     string `json:"status"`
	Slug       string `json:"slug" binding:"required"`
}


// Handlers
// GET /api/articles
func listArticlesHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	status := c.Query("status") // optional filter
	query := "SELECT id, title, slug, content, COALESCE(excerpt,''), COALESCE(cover_image,''), author, COALESCE(category,'Umum'), status, created_at, updated_at FROM articles"
	args := []interface{}{}

	if status != "" && (status == "published" || status == "draft") {
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

	articles := []Article{}
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.Excerpt, &a.CoverImage, &a.Author, &a.Category, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		articles = append(articles, a)
	}

	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

// GET /api/articles/:id
func getArticleHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var a Article
	err = DB.QueryRow(
		"SELECT id, title, slug, content, COALESCE(excerpt,''), COALESCE(cover_image,''), author, COALESCE(category,'Umum'), status, created_at, updated_at FROM articles WHERE id = $1", id,
	).Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.Excerpt, &a.CoverImage, &a.Author, &a.Category, &a.Status, &a.CreatedAt, &a.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, a)
}

// POST /api/articles  (protected)
func createArticleHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	var input ArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "" {
		input.Status = "draft"
	}
	if input.Category == "" {
		input.Category = "Umum"
	}

	author := c.GetString("username")
	if author == "" {
		author = "Admin"
	}

	var a Article
	err := DB.QueryRow(
		`INSERT INTO articles (title, slug, content, excerpt, cover_image, author, category, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, title, slug, content, COALESCE(excerpt,''), COALESCE(cover_image,''), author, category, status, created_at, updated_at`,
		input.Title, input.Slug, input.Content, input.Excerpt, input.CoverImage, author, input.Category, input.Status,
	).Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.Excerpt, &a.CoverImage, &a.Author, &a.Category, &a.Status, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, a)
}

// PUT /api/articles/:id  (protected)
func updateArticleHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input ArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Category == "" {
		input.Category = "Umum"
	}

	var a Article
	err = DB.QueryRow(
		`UPDATE articles SET title=$1, slug=$2, content=$3, excerpt=$4, cover_image=$5, category=$6, status=$7
		 WHERE id=$8
		 RETURNING id, title, slug, content, COALESCE(excerpt,''), COALESCE(cover_image,''), author, category, status, created_at, updated_at`,
		input.Title, input.Slug, input.Content, input.Excerpt, input.CoverImage, input.Category, input.Status, id,
	).Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.Excerpt, &a.CoverImage, &a.Author, &a.Category, &a.Status, &a.CreatedAt, &a.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, a)
}

// DELETE /api/articles/:id  (protected)
func deleteArticleHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database unavailable"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result, err := DB.Exec("DELETE FROM articles WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article deleted"})
}
