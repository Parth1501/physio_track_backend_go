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

const fullDateLayout = "Monday, 02 January 2006"

// message is the shared layout of every patient message: greeting, intro, key facts,
// session history, outro and signature. Messages never include amounts: seeing money
// discourages patients from continuing treatment.
type message struct {
	name         string
	intro        string
	facts        []string
	historyTitle string
	history      []time.Time
	outro        string
	signature    string
}

// SessionText builds the message sent after a session is recorded.
func SessionText(name string, session time.Time, all []time.Time, signature string) string {
	history, title := all, "Session history:"
	if len(history) > RecentLimit {
		history, title = history[len(history)-RecentLimit:], "Recent sessions:"
	}
	return message{
		name:  name,
		intro: "This is to confirm that your physiotherapy session has been completed.",
		facts: []string{
			"Session date: " + bold(session.Format(fullDateLayout)),
			"Total sessions attended: " + bold(strconv.Itoa(len(all))),
		},
		historyTitle: title,
		history:      history,
		outro:        "We look forward to seeing you at your next session.",
		signature:    signature,
	}.String()
}

// SummaryText builds the full session history message the physiotherapist sends on demand.
func SummaryText(name string, all []time.Time, signature string) string {
	facts := []string{"Total sessions attended: " + bold(strconv.Itoa(len(all)))}
	if last, ok := latest(all); ok {
		facts = append(facts, "Last session: "+bold(last.Format(fullDateLayout)))
	}
	return message{
		name:         name,
		intro:        "Please find below a summary of your physiotherapy sessions.",
		facts:        facts,
		historyTitle: "Session history:",
		history:      all,
		outro:        "For any queries, please feel free to reach out.",
		signature:    signature,
	}.String()
}

// ChatURL returns a click-to-chat link that opens the patient's chat with the text pre-filled.
// Spaces are sent as %20 because some WhatsApp clients show "+" literally.
func ChatURL(phone, text string) string {
	return "https://wa.me/" + phone + "?text=" + strings.ReplaceAll(url.QueryEscape(text), "+", "%20")
}

func (m message) String() string {
	var b strings.Builder
	if name := strings.TrimSpace(m.name); name != "" {
		b.WriteString("Dear " + name + ",\n\n")
	} else {
		b.WriteString("Dear Patient,\n\n")
	}
	b.WriteString(m.intro + "\n\n")
	for _, f := range m.facts {
		b.WriteString(f + "\n")
	}
	if lines := groupByMonth(m.history); len(lines) > 0 {
		b.WriteString("\n" + m.historyTitle + "\n")
		for _, line := range lines {
			b.WriteString(line + "\n")
		}
	}
	b.WriteString("\n" + m.outro + "\n\n")
	// The signature may span several lines (name, qualifications).
	if s := strings.TrimSpace(m.signature); s != "" {
		b.WriteString("Warm regards,\n" + s)
	} else {
		b.WriteString("Thank you.")
	}
	return b.String()
}

// bold uses WhatsApp's *text* formatting.
func bold(s string) string {
	return "*" + s + "*"
}

func latest(dates []time.Time) (time.Time, bool) {
	var last time.Time
	for _, d := range dates {
		if d.After(last) {
			last = d
		}
	}
	return last, !last.IsZero()
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
