package handlers

import (
	"net/http"
	"strings"

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

type bulkCreateRequest struct {
	PatientID string          `json:"patient_id"`
	Amount    float64         `json:"amount"`
	Mode      string          `json:"mode"`
	Dates     []core.JSONTime `json:"dates"`
}

const maxBulkPayments = 60

// CreateBulk records one payment per date, all with the same amount and mode.
func (h *PaymentHandler) CreateBulk(c *gin.Context) {
	var req bulkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	mode := strings.ToUpper(strings.TrimSpace(req.Mode))
	if req.PatientID == "" || req.Amount <= 0 || (mode != "CASH" && mode != "ONLINE" && mode != "PENDING") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "patient_id, a positive amount and mode CASH, ONLINE or PENDING are required"})
		return
	}
	if len(req.Dates) == 0 || len(req.Dates) > maxBulkPayments {
		c.JSON(http.StatusBadRequest, gin.H{"error": "between 1 and 60 dates are required"})
		return
	}
	for _, d := range req.Dates {
		if d.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "every date must be set"})
			return
		}
	}
	owner := c.GetString("user")
	items, err := h.repo.CreateBulk(c, owner, req.PatientID, req.Amount, mode, req.Dates)
	if err != nil {
		status := http.StatusInternalServerError
		if err == repo.ErrForbidden {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, items)
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

type settleRequest struct {
	PatientID string  `json:"patient_id"`
	Amount    float64 `json:"amount"`
	Mode      string  `json:"mode"`
}

// Settle applies a lump-sum payment to the patient's pending sessions and returns the updated payment list.
func (h *PaymentHandler) Settle(c *gin.Context) {
	var req settleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	mode := strings.ToUpper(strings.TrimSpace(req.Mode))
	if req.PatientID == "" || req.Amount <= 0 || (mode != "CASH" && mode != "ONLINE") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "patient_id, a positive amount and mode CASH or ONLINE are required"})
		return
	}
	owner := c.GetString("user")
	if err := h.repo.Settle(c, owner, req.PatientID, req.Amount, mode); err != nil {
		status := http.StatusInternalServerError
		switch err {
		case repo.ErrExceedsPending:
			status = http.StatusBadRequest
		case repo.ErrForbidden:
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	items, err := h.repo.List(c, owner, req.PatientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
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
