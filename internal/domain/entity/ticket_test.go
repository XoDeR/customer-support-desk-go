package entity

import (
	"testing"
	"time"
)

func TestTicketTransitionTo(t *testing.T) {
	cases := []struct {
		from, to Status
		want     bool
	}{{StatusOpen, StatusPending, true}, {StatusOpen, StatusResolved, true}, {StatusOpen, StatusClosed, false}, {StatusPending, StatusOpen, true}, {StatusResolved, StatusClosed, true}, {StatusClosed, StatusOpen, false}}
	for _, tt := range cases {
		ticket := Ticket{Status: tt.from}
		err := ticket.TransitionTo(tt.to)
		if (err == nil) != tt.want {
			t.Errorf("%s -> %s: err=%v", tt.from, tt.to, err)
		}
	}
}
func TestTicketPauseResumeSLA(t *testing.T) {
	now := testTime()
	due := now.Add(2e9)
	ticket := Ticket{SLADueAt: &due}
	ticket.PauseSLA(now)
	if ticket.SLAPausedAt == nil || ticket.SLARemainingSeconds == nil {
		t.Fatal("SLA was not paused")
	}
	later := now.Add(10e9)
	ticket.ResumeSLA(later)
	if ticket.SLAPausedAt != nil || ticket.SLADueAt.Sub(later) != 2e9 {
		t.Fatal("SLA was not resumed with remaining duration")
	}
}
func testTime() (v time.Time) { return time.Unix(0, 0) }
