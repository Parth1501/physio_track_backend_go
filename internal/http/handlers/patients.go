package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"phsio_track_backend/internal/core"
	"phsio_track_backend/internal/repo"
)

type PatientHandler struct {
	repo *repo.PatientRepo
}

func NewPatientHandler(repo *repo.PatientRepo) *PatientHandler {
	return &PatientHandler{repo: repo}
}

func (h *PatientHandler) Create(c *gin.Context) {
	var req core.Patient
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	owner := c.GetString("user")
	if err := h.repo.Create(c, owner, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *PatientHandler) List(c *gin.Context) {
	owner := c.GetString("user")
	items, err := h.repo.List(c, owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *PatientHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	owner := c.GetString("user")
	item, err := h.repo.GetByID(c, owner, id)
	if err != nil {
		status := http.StatusNotFound
		if err == repo.ErrNotFound {
			status = http.StatusNotFound
		} else {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *PatientHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req core.PatientUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	owner := c.GetString("user")
	updated, err := h.repo.Update(c, owner, id, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == repo.ErrNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}
