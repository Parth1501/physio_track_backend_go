package whatsapp

import (
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// RecentLimit is how many session dates the per-session message lists.
const RecentLimit = 10

// SessionText builds the message sent after a session is recorded.
// It never includes amounts: seeing money discourages patients from continuing treatment.
func SessionText(name string, session time.Time, all []time.Time, signature string) string {
	recent := all
	if len(recent) > RecentLimit {
		recent = recent[len(recent)-RecentLimit:]
	}
	var b strings.Builder
	b.WriteString(greeting(name))
	b.WriteString("This is to confirm that your physiotherapy session on *" + session.Format("02 Jan 2006") + "* has been completed.\n\n")
	b.WriteString("Total sessions attended: *" + strconv.Itoa(len(all)) + "*\n")
	if len(all) > len(recent) {
		b.WriteString("Recent session dates:\n")
	} else {
		b.WriteString("Session dates:\n")
	}
	writeDates(&b, recent)
	b.WriteString("\nWe look forward to seeing you at your next session.\n")
	writeClosing(&b, signature)
	return b.String()
}

// SummaryText builds the full session history message the physiotherapist sends on demand.
func SummaryText(name string, all []time.Time, signature string) string {
	var b strings.Builder
	b.WriteString(greeting(name))
	b.WriteString("Please find below a summary of your physiotherapy sessions.\n\n")
	b.WriteString("Total sessions attended: *" + strconv.Itoa(len(all)) + "*\n")
	b.WriteString("Session dates:\n")
	writeDates(&b, all)
	b.WriteString("\nFor any queries, please feel free to reach out.\n")
	writeClosing(&b, signature)
	return b.String()
}

// ChatURL returns a click-to-chat link that opens the patient's chat with the text pre-filled.
// Spaces are sent as %20 because some WhatsApp clients show "+" literally.
func ChatURL(phone, text string) string {
	return "https://wa.me/" + phone + "?text=" + strings.ReplaceAll(url.QueryEscape(text), "+", "%20")
}

func greeting(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Dear Patient,\n\n"
	}
	return "Dear " + name + ",\n\n"
}

func writeDates(b *strings.Builder, dates []time.Time) {
	for _, line := range groupByMonth(dates) {
		b.WriteString(line + "\n")
	}
}

// writeClosing ends the message with the signature, which may span several lines (name, qualifications).
func writeClosing(b *strings.Builder, signature string) {
	if s := strings.TrimSpace(signature); s != "" {
		b.WriteString("\nWarm regards,\n" + s)
		return
	}
	b.WriteString("\nThank you.")
}

// groupByMonth renders dates oldest first as one line per month, e.g. "Oct 2026: 01, 02, 03".
func groupByMonth(dates []time.Time) []string {
	sorted := append([]time.Time(nil), dates...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Before(sorted[j]) })

	var lines []string
	var month string
	var days []string
	flush := func() {
		if month != "" {
			lines = append(lines, month+": "+strings.Join(days, ", "))
		}
	}
	for _, d := range sorted {
		m := d.Format("Jan 2006")
		if m != month {
			flush()
			month, days = m, nil
		}
		days = append(days, d.Format("02"))
	}
	flush()
	return lines
}
