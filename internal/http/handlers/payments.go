package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"phsio_track_backend/internal/core"
	"phsio_track_backend/internal/repo"
)

type PaymentHandler struct {
	repo *repo.PaymentRepo
}

func NewPaymentHandler(repo *repo.PaymentRepo) *PaymentHandler {
	return &PaymentHandler{repo: repo}
}

func (h *PaymentHandler) Create(c *gin.Context) {
	var req core.Payment
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

func (h *PaymentHandler) List(c *gin.Context) {
	patientID := c.Query("patient_id")
	owner := c.GetString("user")
	items, err := h.repo.List(c, owner, patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *PaymentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req core.PaymentUpdate
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

func (h *PaymentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	owner := c.GetString("user")
	if err := h.repo.Delete(c, owner, id); err != nil {
		status := http.StatusInternalServerError
		if err == repo.ErrNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
