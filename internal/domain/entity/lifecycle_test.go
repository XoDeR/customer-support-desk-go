package entity_test

import (
	"testing"
	"time"

	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
)

// TicketLifecycle exercises the domain state machine and SLA pause/resume rules
// that gate the agent workflow end-to-end.
func TestTicketLifecycle(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	due := now.Add(8 * time.Hour)
	ticket := entity.Ticket{
		Status:   entity.StatusOpen,
		Priority: entity.PriorityHigh,
		SLADueAt: &due,
	}

	if err := ticket.TransitionTo(entity.StatusClosed); err == nil {
		t.Fatal("expected open→closed to be rejected")
	}

	if err := ticket.TransitionTo(entity.StatusPending); err != nil {
		t.Fatalf("open→pending: %v", err)
	}
	ticket.PauseSLA(now.Add(time.Hour))
	if ticket.SLAPausedAt == nil || ticket.SLARemainingSeconds == nil {
		t.Fatal("pending should pause SLA")
	}
	remaining := *ticket.SLARemainingSeconds
	if remaining <= 0 {
		t.Fatalf("expected positive remaining, got %d", remaining)
	}

	if err := ticket.TransitionTo(entity.StatusOpen); err != nil {
		t.Fatalf("pending→open: %v", err)
	}
	resumeAt := now.Add(2 * time.Hour)
	ticket.ResumeSLA(resumeAt)
	if ticket.SLAPausedAt != nil {
		t.Fatal("resume should clear pause")
	}
	if ticket.SLADueAt == nil || !ticket.SLADueAt.Equal(resumeAt.Add(time.Duration(remaining)*time.Second)) {
		t.Fatalf("unexpected resumed due at %v", ticket.SLADueAt)
	}

	if err := ticket.TransitionTo(entity.StatusResolved); err != nil {
		t.Fatalf("open→resolved: %v", err)
	}
	if err := ticket.TransitionTo(entity.StatusClosed); err != nil {
		t.Fatalf("resolved→closed: %v", err)
	}
}
