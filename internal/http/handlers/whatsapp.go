package handlers

import (
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	"phsio_track_backend/internal/repo"
	"phsio_track_backend/internal/whatsapp"
)

type WhatsAppHandler struct {
	patients  *repo.PatientRepo
	payments  *repo.PaymentRepo
	signature string
}

func NewWhatsAppHandler(patients *repo.PatientRepo, payments *repo.PaymentRepo, signature string) *WhatsAppHandler {
	return &WhatsAppHandler{patients: patients, payments: payments, signature: signature}
}

// Message builds a pre-filled WhatsApp message for the patient; the app opens the returned
// click-to-chat URL so it is sent from the physiotherapist's own WhatsApp.
// kind=session (with payment_id) describes one recorded session; kind=summary lists every session.
func (h *WhatsAppHandler) Message(c *gin.Context) {
	kind := c.Query("kind")
	paymentID := c.Query("payment_id")
	if kind != "session" && kind != "summary" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be session or summary"})
		return
	}
	if kind == "session" && paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required for a session message"})
		return
	}

	owner := c.GetString("user")
	patient, err := h.patients.GetByID(c, owner, c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if err == repo.ErrNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	// The opt-out only stops the automatic prompt after a session; the summary is always sent on demand.
	if kind == "session" && patient.WhatsAppOptIn != nil && !*patient.WhatsAppOptIn {
		c.JSON(http.StatusConflict, gin.H{"error": "patient opted out of WhatsApp updates"})
		return
	}
	phone, err := whatsapp.NormalizeIndian(patient.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payments, err := h.payments.List(c, owner, patient.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Each session is one payment row; rows split off by a partial settle-up belong to an existing session.
	var sessions []time.Time
	var sessionDate time.Time
	for _, p := range payments {
		if p.PartialOf != "" || p.Date.IsZero() {
			continue
		}
		sessions = append(sessions, p.Date.Time)
		if p.ID == paymentID {
			sessionDate = p.Date.Time
		}
	}
	if len(sessions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no sessions recorded yet"})
		return
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].Before(sessions[j]) })

	var text string
	if kind == "session" {
		if sessionDate.IsZero() {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		text = whatsapp.SessionText(patient.FullName, sessionDate, sessions, h.signature)
	} else {
		text = whatsapp.SummaryText(patient.FullName, sessions, h.signature)
	}

	c.JSON(http.StatusOK, gin.H{
		"phone": phone,
		"text":  text,
		"url":   whatsapp.ChatURL(phone, text),
	})
}
