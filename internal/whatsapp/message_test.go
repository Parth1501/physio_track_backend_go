package whatsapp

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func day(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestGroupByMonth(t *testing.T) {
	got := groupByMonth([]time.Time{day("2026-10-02"), day("2026-09-28"), day("2026-10-01"), day("2026-09-30")})
	want := []string{"Sep 2026: 28, 30", "Oct 2026: 01, 02"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("groupByMonth = %q; want %q", got, want)
	}
}

func TestSessionTextCapsRecentDates(t *testing.T) {
	var all []time.Time
	start := day("2026-09-01")
	for i := 0; i < 12; i++ {
		all = append(all, start.AddDate(0, 0, i))
	}
	text := SessionText("Rahul", all[11], all, "")

	for _, want := range []string{"Dear Rahul,", "*12 Sep 2026*", "Total sessions attended: *12*", "Recent session dates:", "Sep 2026: 03, 04, 05, 06, 07, 08, 09, 10, 11, 12"} {
		if !strings.Contains(text, want) {
			t.Errorf("SessionText missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "01, 02") {
		t.Errorf("SessionText should list only the last %d dates:\n%s", RecentLimit, text)
	}
	if !strings.HasSuffix(text, "Thank you.") || strings.Contains(text, "regards") {
		t.Errorf("SessionText without a signature should end with a plain thank-you:\n%s", text)
	}
}

func TestSummaryTextListsAllDatesWithoutAmounts(t *testing.T) {
	all := []time.Time{day("2026-08-30"), day("2026-09-02"), day("2026-09-05")}
	text := SummaryText("Rahul", all, "Dr. Dency Singwala\n(MPT, COMT, CKT)")

	for _, want := range []string{"Dear Rahul,", "Total sessions attended: *3*", "Aug 2026: 30", "Sep 2026: 02, 05"} {
		if !strings.Contains(text, want) {
			t.Errorf("SummaryText missing %q:\n%s", want, text)
		}
	}
	if !strings.HasSuffix(text, "Warm regards,\nDr. Dency Singwala\n(MPT, COMT, CKT)") {
		t.Errorf("SummaryText should end with the signature:\n%s", text)
	}
	for _, banned := range []string{"₹", "rs.", "inr", "amount", "paid", "pending"} {
		if strings.Contains(strings.ToLower(text), strings.ToLower(banned)) {
			t.Errorf("SummaryText must not mention money (%q):\n%s", banned, text)
		}
	}
}

func TestChatURLRoundTrips(t *testing.T) {
	text := "Hi Rahul,\nSessions: *2* ✅ & 1+1 more"
	raw := ChatURL("919876543210", text)
	if strings.Contains(raw, "+") {
		t.Errorf("chat URL should encode spaces as %%20, not +: %s", raw)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "wa.me" || u.Path != "/919876543210" {
		t.Errorf("unexpected chat URL %s", u)
	}
	if got := u.Query().Get("text"); got != text {
		t.Errorf("text = %q; want %q", got, text)
	}
}
