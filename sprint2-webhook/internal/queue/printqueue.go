// Package queue implements a buffered in-memory print job queue using a Go
// channel. In production this would be replaced by RabbitMQ, AWS SQS, or
// a similar message broker.
package queue

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// PrintJob represents a badge print request published by the kiosk.
type PrintJob struct {
	JobID        string
	AttendeeID   string
	AttendeeName string
	EnqueuedAt   time.Time
}

// PrintQueue is a buffered channel-based message queue.
type PrintQueue struct {
	ch chan PrintJob
}

// New returns a PrintQueue with the given buffer capacity.
func New(capacity int) *PrintQueue {
	return &PrintQueue{ch: make(chan PrintJob, capacity)}
}

// Publish enqueues a job non-blocking. Returns an error if the queue is full.
func (q *PrintQueue) Publish(job PrintJob) error {
	select {
	case q.ch <- job:
		log.Printf("[queue] Published job %s for attendee %s", job.JobID, job.AttendeeID)
		return nil
	default:
		return fmt.Errorf("print queue full — job %s rejected", job.JobID)
	}
}

// StartWorker launches a goroutine that consumes jobs and fires the vendor
// callback once each simulated print job completes. done signals shutdown.
func (q *PrintQueue) StartWorker(callbackURL string, done <-chan struct{}) {
	go func() {
		log.Printf("[queue-worker] Started. Listening for print jobs...")
		for {
			select {
			case job := <-q.ch:
				go q.processJob(job, callbackURL)
			case <-done:
				log.Println("[queue-worker] Shutdown signal received. Stopping.")
				return
			}
		}
	}()
}

// processJob simulates vendor processing (1.5–3.5s delay) then POSTs the
// callback to callbackURL, mimicking the vendor's async push notification.
func (q *PrintQueue) processJob(job PrintJob, callbackURL string) {
	log.Printf("[queue-worker] Processing job %s (attendee: %s)...", job.JobID, job.AttendeeID)

	delay := time.Duration(1500+rand.Intn(2000)) * time.Millisecond
	time.Sleep(delay)

	payload := fmt.Sprintf(
		`{"jobId":"%s","attendeeId":"%s","status":"success","printedAt":"%s","vendor":"SolsticePrint v2"}`,
		job.JobID,
		job.AttendeeID,
		time.Now().UTC().Format(time.RFC3339),
	)

	resp, err := http.Post(callbackURL, "application/json", strings.NewReader(payload))
	if err != nil {
		log.Printf("[queue-worker] ERROR: callback failed for job %s: %v", job.JobID, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("[queue-worker] Job %s callback delivered (HTTP %d, delay: %v)", job.JobID, resp.StatusCode, delay)
}

// GenerateJobID returns a unique print job identifier.
func GenerateJobID() string {
	return fmt.Sprintf("PJ-%d-%04d", time.Now().UnixMilli(), rand.Intn(9999))
}
