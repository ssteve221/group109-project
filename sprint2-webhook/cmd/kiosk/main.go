// Day 4 Pivot — Solstice Events Async Check-In Kiosk
//
// Flow: QR scan → publish to message queue → 202 Pending
//       → vendor processes → POST /webhook/print-callback → checked_in
//
// Run: go run ./cmd/kiosk/
//      Open http://localhost:7070
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/group109/northstar-webhook/internal/checkin"
	"github.com/group109/northstar-webhook/internal/queue"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "7070"
	}

	// In production, callbackURL would be a public HTTPS URL (e.g. via ngrok).
	callbackURL := fmt.Sprintf("http://localhost:%s/webhook/print-callback", port)

	store := checkin.NewStore()
	printQueue := queue.New(100)

	done := make(chan struct{})
	printQueue.StartWorker(callbackURL, done)
	defer close(done)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, kioskHTML)
	})

	mux.HandleFunc("GET /api/attendees", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		json.NewEncoder(w).Encode(store.All())
	})

	mux.HandleFunc("GET /api/status/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/api/status/"):]
		if id == "" {
			http.Error(w, "attendee ID required", http.StatusBadRequest)
			return
		}
		a, err := store.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(a)
	})

	// POST /api/checkin — QR scan entry point. Publishes to queue, returns 202 immediately.
	mux.HandleFunc("POST /api/checkin", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			AttendeeID string `json:"attendeeId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AttendeeID == "" {
			http.Error(w, `{"error":"attendeeId required"}`, http.StatusBadRequest)
			return
		}

		jobID := queue.GenerateJobID()

		attendee, err := store.BeginCheckIn(req.AttendeeID, jobID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			switch e := err.(type) {
			case *checkin.ErrAlreadyProcessing:
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]any{
					"error":   e.Error(),
					"status":  string(e.Current),
					"message": "Duplicate scan prevented — no second badge will be printed.",
				})
			case *checkin.ErrNotFound:
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
			default:
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			}
			return
		}

		job := queue.PrintJob{
			JobID:        jobID,
			AttendeeID:   attendee.ID,
			AttendeeName: attendee.Name,
			EnqueuedAt:   time.Now(),
		}
		if err := printQueue.Publish(job); err != nil {
			log.Printf("[checkin] Queue full — could not enqueue job for %s: %v", req.AttendeeID, err)
			http.Error(w, `{"error":"print queue full, please retry"}`, http.StatusServiceUnavailable)
			return
		}

		log.Printf("[checkin] Scan accepted: attendee=%s job=%s", attendee.ID, jobID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]any{
			"status":      "pending",
			"message":     "Badge print job queued. Waiting for printer confirmation...",
			"attendeeId":  attendee.ID,
			"attendeeName": attendee.Name,
			"jobId":       jobID,
		})
	})

	// POST /webhook/print-callback — vendor calls this when a print job completes.
	mux.HandleFunc("POST /webhook/print-callback", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var callback struct {
			JobID      string `json:"jobId"`
			AttendeeID string `json:"attendeeId"`
			Status     string `json:"status"`
			PrintedAt  string `json:"printedAt"`
			Vendor     string `json:"vendor"`
		}
		if err := json.Unmarshal(body, &callback); err != nil {
			log.Printf("[webhook] ERROR parsing callback: %v | body: %s", err, body)
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("[webhook] Received callback: job=%s attendee=%s status=%s vendor=%s",
			callback.JobID, callback.AttendeeID, callback.Status, callback.Vendor)

		switch callback.Status {
		case "success":
			attendee, err := store.ConfirmPrint(callback.JobID)
			if err != nil {
				log.Printf("[webhook] ERROR confirming job %s: %v", callback.JobID, err)
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			log.Printf("[webhook] ✅ Attendee %s (%s) CHECKED IN", attendee.ID, attendee.Name)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"received": true,
				"status":   "checked_in",
				"attendee": attendee.ID,
			})

		case "failed":
			attendee, err := store.FailPrint(callback.JobID, "vendor reported failure")
			if err != nil {
				log.Printf("[webhook] ERROR failing job %s: %v", callback.JobID, err)
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			log.Printf("[webhook] ❌ Print FAILED for attendee %s", attendee.ID)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"received": true,
				"status":   "failed",
				"attendee": attendee.ID,
			})

		default:
			http.Error(w, "unknown status in callback", http.StatusBadRequest)
		}
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":      "ok",
			"service":     "solstice-checkin-kiosk",
			"mode":        "async-webhook",
			"callbackURL": callbackURL,
			"time":        time.Now().UTC().Format(time.RFC3339),
		})
	})

	addr := ":" + port
	log.Printf("[kiosk] ════════════════════════════════════════")
	log.Printf("[kiosk]  Solstice Events Check-In Kiosk")
	log.Printf("[kiosk]  Mode: ASYNC (webhook push model)")
	log.Printf("[kiosk]  UI:   http://localhost%s", addr)
	log.Printf("[kiosk]  Callback: %s", callbackURL)
	log.Printf("[kiosk] ════════════════════════════════════════")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[kiosk] Fatal: %v", err)
	}
}

// kioskHTML is the self-contained single-page kiosk UI.
const kioskHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Solstice Events — Check-In Kiosk</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap" rel="stylesheet">
<style>
  :root {
    --bg: #0b0f1a;
    --surface: #131929;
    --surface2: #1a2238;
    --border: #1f2d4a;
    --accent: #4f8ef7;
    --accent2: #7c5cfc;
    --success: #22c55e;
    --warning: #f59e0b;
    --danger: #ef4444;
    --text: #e8edf7;
    --text-muted: #6b7fa8;
    --glow: rgba(79,142,247,0.15);
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: 'Inter', sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    background-image: radial-gradient(ellipse at 20% 20%, rgba(79,142,247,0.06) 0%, transparent 60%),
                      radial-gradient(ellipse at 80% 80%, rgba(124,92,252,0.05) 0%, transparent 60%);
  }
  header {
    padding: 24px 32px;
    border-bottom: 1px solid var(--border);
    background: rgba(19,25,41,0.8);
    backdrop-filter: blur(12px);
    display: flex;
    align-items: center;
    justify-content: space-between;
    position: sticky;
    top: 0;
    z-index: 10;
  }
  header h1 { font-size: 1.25rem; font-weight: 700; letter-spacing: -0.02em; }
  header h1 span { color: var(--accent); }
  .badge-mode {
    font-size: 0.72rem;
    background: rgba(79,142,247,0.12);
    border: 1px solid rgba(79,142,247,0.25);
    color: var(--accent);
    padding: 4px 12px;
    border-radius: 20px;
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }
  .main { max-width: 960px; margin: 0 auto; padding: 32px 24px; }
  .pivot-notice {
    background: rgba(245,158,11,0.08);
    border: 1px solid rgba(245,158,11,0.25);
    border-radius: 12px;
    padding: 16px 20px;
    margin-bottom: 28px;
    font-size: 0.85rem;
    color: #fcd34d;
    display: flex;
    gap: 10px;
    align-items: flex-start;
  }
  .pivot-notice .icon { font-size: 1.1rem; flex-shrink: 0; margin-top: 1px; }
  .pivot-notice strong { display: block; margin-bottom: 4px; }
  .section-title {
    font-size: 0.7rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--text-muted);
    margin-bottom: 16px;
  }
  .attendee-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
    margin-bottom: 32px;
  }
  .card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 16px;
    padding: 20px;
    transition: all 0.2s ease;
    position: relative;
    overflow: hidden;
  }
  .card::before {
    content: '';
    position: absolute;
    top: 0; left: 0; right: 0;
    height: 2px;
    background: linear-gradient(90deg, var(--accent), var(--accent2));
    opacity: 0;
    transition: opacity 0.3s;
  }
  .card:hover { border-color: rgba(79,142,247,0.35); box-shadow: 0 0 24px var(--glow); }
  .card:hover::before { opacity: 1; }
  .card.status-pending { border-color: rgba(245,158,11,0.4); }
  .card.status-pending::before { background: #f59e0b; opacity: 1; }
  .card.status-checked_in { border-color: rgba(34,197,94,0.4); }
  .card.status-checked_in::before { background: var(--success); opacity: 1; }
  .card.status-failed { border-color: rgba(239,68,68,0.4); }
  .card.status-failed::before { background: var(--danger); opacity: 1; }
  .card-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 12px; }
  .attendee-name { font-weight: 700; font-size: 1rem; }
  .attendee-id { font-size: 0.72rem; color: var(--text-muted); font-family: monospace; margin-top: 2px; }
  .status-pill {
    font-size: 0.7rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    padding: 4px 10px;
    border-radius: 20px;
    white-space: nowrap;
    flex-shrink: 0;
  }
  .status-pill.unknown  { background: rgba(107,127,168,0.15); color: var(--text-muted); }
  .status-pill.pending  { background: rgba(245,158,11,0.15); color: #fcd34d; animation: pulse 1.5s ease-in-out infinite; }
  .status-pill.checked_in { background: rgba(34,197,94,0.15); color: var(--success); }
  .status-pill.failed   { background: rgba(239,68,68,0.15); color: var(--danger); }
  @keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.5; } }
  .attendee-detail { font-size: 0.8rem; color: var(--text-muted); margin-bottom: 14px; }
  .attendee-detail span { display: block; margin-bottom: 2px; }
  .job-info {
    font-size: 0.72rem;
    font-family: monospace;
    color: rgba(107,127,168,0.7);
    background: rgba(0,0,0,0.2);
    padding: 6px 10px;
    border-radius: 6px;
    margin-bottom: 12px;
    word-break: break-all;
  }
  .scan-btn {
    width: 100%;
    padding: 10px 16px;
    border: none;
    border-radius: 10px;
    cursor: pointer;
    font-family: 'Inter', sans-serif;
    font-size: 0.85rem;
    font-weight: 600;
    transition: all 0.15s ease;
    background: linear-gradient(135deg, var(--accent), var(--accent2));
    color: white;
  }
  .scan-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 16px rgba(79,142,247,0.4); }
  .scan-btn:active { transform: translateY(0); }
  .scan-btn:disabled {
    background: var(--surface2);
    color: var(--text-muted);
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
    border: 1px solid var(--border);
  }
  .scan-btn.pending-btn {
    background: rgba(245,158,11,0.1);
    border: 1px solid rgba(245,158,11,0.3);
    color: #fcd34d;
    animation: pulse 1.5s ease-in-out infinite;
  }
  .log-section { margin-top: 20px; }
  .log-box {
    background: #060b14;
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 16px;
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 0.75rem;
    height: 200px;
    overflow-y: auto;
    color: #7fba7f;
    scroll-behavior: smooth;
  }
  .log-entry { margin-bottom: 4px; }
  .log-entry .ts { color: var(--text-muted); }
  .log-entry.info  { color: #7fba7f; }
  .log-entry.warn  { color: #fcd34d; }
  .log-entry.error { color: #f87171; }
  .log-entry.webhook { color: #60a5fa; }
  .toast-container { position: fixed; bottom: 24px; right: 24px; display: flex; flex-direction: column; gap: 10px; z-index: 999; }
  .toast {
    padding: 14px 20px;
    border-radius: 12px;
    font-size: 0.875rem;
    font-weight: 500;
    min-width: 280px;
    box-shadow: 0 8px 32px rgba(0,0,0,0.4);
    animation: slideIn 0.25s ease;
    display: flex;
    align-items: center;
    gap: 10px;
    border: 1px solid;
  }
  .toast.success { background: rgba(34,197,94,0.12); border-color: rgba(34,197,94,0.3); color: #86efac; }
  .toast.warning { background: rgba(245,158,11,0.12); border-color: rgba(245,158,11,0.3); color: #fcd34d; }
  .toast.error   { background: rgba(239,68,68,0.12);  border-color: rgba(239,68,68,0.3);  color: #fca5a5; }
  @keyframes slideIn { from { transform: translateX(120%); opacity: 0; } to { transform: translateX(0); opacity: 1; } }
  .stats-bar {
    display: flex;
    gap: 16px;
    margin-bottom: 24px;
    flex-wrap: wrap;
  }
  .stat {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 12px 20px;
    flex: 1;
    min-width: 140px;
    text-align: center;
  }
  .stat-value { font-size: 1.8rem; font-weight: 800; letter-spacing: -0.02em; }
  .stat-label { font-size: 0.72rem; color: var(--text-muted); font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; margin-top: 2px; }
  .stat.stat-pending .stat-value  { color: #fcd34d; }
  .stat.stat-checked .stat-value  { color: var(--success); }
  .stat.stat-total .stat-value    { color: var(--accent); }
</style>
</head>
<body>
<header>
  <div>
    <h1>⚡ <span>Solstice</span> Events — Check-In Kiosk</h1>
  </div>
  <div class="badge-mode">🔄 Async Webhook Mode</div>
</header>
<div class="main">
  <div class="pivot-notice">
    <span class="icon">⚠️</span>
    <div>
      <strong>Day 4 Pivot — Badge printer vendor deprecated synchronous API</strong>
      This kiosk now uses an <strong>async message-queue + webhook model</strong>.
      Scanning a QR code publishes a print job to the queue and immediately shows a <em>Pending</em> state.
      The screen updates to <em>Checked In</em> only after the vendor's webhook callback confirms the badge was printed.
    </div>
  </div>

  <div class="stats-bar" id="stats-bar">
    <div class="stat stat-total"><div class="stat-value" id="stat-total">—</div><div class="stat-label">Total Attendees</div></div>
    <div class="stat stat-pending"><div class="stat-value" id="stat-pending">—</div><div class="stat-label">Pending Print</div></div>
    <div class="stat stat-checked"><div class="stat-value" id="stat-checked">—</div><div class="stat-label">Checked In</div></div>
  </div>

  <p class="section-title">📋 Attendee Registry</p>
  <div class="attendee-grid" id="attendee-grid">Loading...</div>

  <div class="log-section">
    <p class="section-title">📡 Event Log</p>
    <div class="log-box" id="log-box"></div>
  </div>
</div>
<div class="toast-container" id="toast-container"></div>

<script>
const STATUS_LABELS = {
  unknown:    { label: 'Not Scanned', icon: '⬜' },
  pending:    { label: '⏳ Printing...', icon: '🟡' },
  checked_in: { label: '✅ Checked In', icon: '🟢' },
  failed:     { label: '❌ Failed', icon: '🔴' },
};

let lastStates = {};

function log(msg, type = 'info') {
  const box = document.getElementById('log-box');
  const ts = new Date().toLocaleTimeString();
  const el = document.createElement('div');
  el.className = 'log-entry ' + type;
  el.innerHTML = '<span class="ts">[' + ts + ']</span> ' + msg;
  box.appendChild(el);
  box.scrollTop = box.scrollHeight;
}

function toast(msg, type = 'success') {
  const c = document.getElementById('toast-container');
  const t = document.createElement('div');
  t.className = 'toast ' + type;
  t.textContent = msg;
  c.appendChild(t);
  setTimeout(() => t.remove(), 4000);
}

function renderCard(a) {
  const sl = STATUS_LABELS[a.status] || { label: a.status, icon: '?' };
  const isActionable = a.status === 'unknown' || a.status === 'failed';
  const isPending = a.status === 'pending';
  const isChecked = a.status === 'checked_in';

  let btnText = '🔍 Scan QR Code';
  let btnClass = 'scan-btn';
  let btnDisabled = '';

  if (isPending) {
    btnText = '⏳ Awaiting Print Callback...';
    btnClass = 'scan-btn pending-btn';
    btnDisabled = 'disabled';
  } else if (isChecked) {
    btnText = '✅ Already Checked In';
    btnDisabled = 'disabled';
  }

  const jobLine = a.jobId
    ? '<div class="job-info">Job: ' + a.jobId + '</div>'
    : '';

  const printedLine = a.printedAt
    ? '<div style="font-size:0.72rem;color:var(--success);margin-bottom:8px;">✅ Printed: ' + new Date(a.printedAt).toLocaleTimeString() + '</div>'
    : '';

  return '<div class="card status-' + a.status + '" id="card-' + a.id + '">' +
    '<div class="card-header">' +
      '<div><div class="attendee-name">' + a.name + '</div><div class="attendee-id">' + a.id + '</div></div>' +
      '<span class="status-pill ' + a.status + '">' + sl.label + '</span>' +
    '</div>' +
    '<div class="attendee-detail">' +
      '<span>📧 ' + a.email + '</span>' +
      '<span>🏢 ' + a.company + '</span>' +
    '</div>' +
    jobLine +
    printedLine +
    '<button class="' + btnClass + '" ' + btnDisabled + ' onclick="scan(\'' + a.id + '\')">' + btnText + '</button>' +
  '</div>';
}

async function refresh() {
  try {
    const res = await fetch('/api/attendees');
    const attendees = await res.json();

    let total = attendees.length, pending = 0, checked = 0;
    let html = '';

    for (const a of attendees) {
      html += renderCard(a);
      if (a.status === 'pending') pending++;
      if (a.status === 'checked_in') checked++;

      // Detect state changes and log / toast them
      const prev = lastStates[a.id];
      if (prev && prev !== a.status) {
        if (a.status === 'checked_in') {
          log('🟢 Webhook received — ' + a.name + ' (' + a.id + ') is now CHECKED IN', 'webhook');
          toast('✅ ' + a.name + ' is Checked In!', 'success');
        } else if (a.status === 'failed') {
          log('🔴 Print FAILED for ' + a.name + ' (' + a.id + ')', 'error');
          toast('❌ Print failed for ' + a.name, 'error');
        } else if (a.status === 'pending') {
          log('🟡 Print job queued for ' + a.name + ' — awaiting webhook...', 'info');
        }
      }
      lastStates[a.id] = a.status;
    }

    document.getElementById('attendee-grid').innerHTML = html;
    document.getElementById('stat-total').textContent = total;
    document.getElementById('stat-pending').textContent = pending;
    document.getElementById('stat-checked').textContent = checked;

  } catch (e) {
    log('⚠️ Polling error: ' + e.message, 'error');
  }
}

async function scan(attendeeId) {
  log('📲 Scanning QR for attendee: ' + attendeeId, 'info');
  try {
    const res = await fetch('/api/checkin', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ attendeeId }),
    });
    const data = await res.json();

    if (res.status === 202) {
      log('📤 Print job queued: ' + data.jobId + ' — status: PENDING', 'info');
      toast('⏳ ' + data.attendeeName + ' — print job queued!', 'warning');
    } else if (res.status === 409) {
      log('🚫 DUPLICATE SCAN blocked: ' + attendeeId + ' (' + data.status + ')', 'warn');
      toast('🚫 Duplicate scan! ' + (data.message || ''), 'error');
    } else {
      log('❌ Check-in error: ' + (data.error || res.status), 'error');
      toast('❌ Error: ' + (data.error || 'Unknown error'), 'error');
    }
  } catch (e) {
    log('❌ Network error: ' + e.message, 'error');
  }
  await refresh();
}

refresh();
setInterval(refresh, 1500);
log('🚀 Kiosk started — async webhook mode active', 'info');
</script>
</body>
</html>`
