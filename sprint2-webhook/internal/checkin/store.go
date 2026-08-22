// Package checkin manages attendee check-in state for the Solstice Events kiosk.
//
// State machine: unknown → pending → checked_in (or failed).
// Duplicate scans while pending or checked_in are rejected before any second
// badge job is queued.
package checkin

import (
	"fmt"
	"sync"
	"time"
)

// Status represents the check-in state of an attendee.
type Status string

const (
	StatusUnknown   Status = "unknown"
	StatusPending   Status = "pending"
	StatusCheckedIn Status = "checked_in"
	StatusFailed    Status = "failed"
)

// Attendee holds a conference attendee's registration data and check-in state.
type Attendee struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Company   string    `json:"company"`
	Status    Status    `json:"status"`
	JobID     string    `json:"jobId,omitempty"`
	ScannedAt time.Time `json:"scannedAt,omitempty"`
	PrintedAt time.Time `json:"printedAt,omitempty"`
}

// Store is a thread-safe in-memory attendee registry.
type Store struct {
	mu        sync.RWMutex
	attendees map[string]*Attendee
}

// NewStore returns a Store pre-seeded with the 3 required test attendees.
func NewStore() *Store {
	s := &Store{attendees: make(map[string]*Attendee)}
	demo := []*Attendee{
		{ID: "ATT001", Name: "Alice Kamau", Email: "alice@solstice.dev", Company: "TechNova Ltd", Status: StatusUnknown},
		{ID: "ATT002", Name: "Brian Otieno", Email: "brian@solstice.dev", Company: "DevHouse Africa", Status: StatusUnknown},
		{ID: "ATT003", Name: "Clara Mwangi", Email: "clara@solstice.dev", Company: "CloudPeak Systems", Status: StatusUnknown},
	}
	for _, a := range demo {
		s.attendees[a.ID] = a
	}
	return s
}

// ErrAlreadyProcessing is returned when a duplicate scan is attempted.
type ErrAlreadyProcessing struct {
	AttendeeID string
	Current    Status
}

func (e *ErrAlreadyProcessing) Error() string {
	switch e.Current {
	case StatusPending:
		return fmt.Sprintf("attendee %s already has a print job in progress (pending)", e.AttendeeID)
	case StatusCheckedIn:
		return fmt.Sprintf("attendee %s is already checked in — no second badge will be printed", e.AttendeeID)
	default:
		return fmt.Sprintf("attendee %s cannot be processed (status: %s)", e.AttendeeID, e.Current)
	}
}

// ErrNotFound is returned when an attendee ID is not registered.
type ErrNotFound struct{ ID string }

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("attendee %q not found in registry", e.ID)
}

// Get returns a copy of an attendee record.
func (s *Store) Get(id string) (Attendee, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.attendees[id]
	if !ok {
		return Attendee{}, &ErrNotFound{ID: id}
	}
	return *a, nil
}

// All returns a snapshot of all attendees.
func (s *Store) All() []Attendee {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Attendee, 0, len(s.attendees))
	for _, a := range s.attendees {
		out = append(out, *a)
	}
	return out
}

// BeginCheckIn transitions an attendee from unknown → pending and assigns a
// job ID. Returns ErrAlreadyProcessing if the attendee is pending or checked_in.
func (s *Store) BeginCheckIn(attendeeID, jobID string) (*Attendee, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.attendees[attendeeID]
	if !ok {
		return nil, &ErrNotFound{ID: attendeeID}
	}
	if a.Status == StatusPending || a.Status == StatusCheckedIn {
		return nil, &ErrAlreadyProcessing{AttendeeID: attendeeID, Current: a.Status}
	}

	a.Status = StatusPending
	a.JobID = jobID
	a.ScannedAt = time.Now()
	return a, nil
}

// ConfirmPrint transitions an attendee from pending → checked_in when the
// vendor webhook callback arrives.
func (s *Store) ConfirmPrint(jobID string) (*Attendee, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, a := range s.attendees {
		if a.JobID == jobID {
			if a.Status != StatusPending {
				return nil, fmt.Errorf("job %s: attendee %s is not pending (current: %s)", jobID, a.ID, a.Status)
			}
			a.Status = StatusCheckedIn
			a.PrintedAt = time.Now()
			return a, nil
		}
	}
	return nil, fmt.Errorf("no attendee found with job ID %q", jobID)
}

// FailPrint transitions an attendee from pending → failed.
func (s *Store) FailPrint(jobID, reason string) (*Attendee, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, a := range s.attendees {
		if a.JobID == jobID {
			a.Status = StatusFailed
			return a, nil
		}
	}
	return nil, fmt.Errorf("no attendee found with job ID %q", jobID)
}
