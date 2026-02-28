package main

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ── Tick ────────────────────────────────────────────────────

type animTickMsg time.Time

func animTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

// ── Snapshot diffing ────────────────────────────────────────

type snapshot struct {
	files map[string]string // path → status
}

func newSnapshot(files []fileEntry) snapshot {
	m := make(map[string]string, len(files))
	for _, f := range files {
		m[f.path] = f.status
	}
	return snapshot{files: m}
}

func (s snapshot) diff(current []fileEntry) []changeEvent {
	now := time.Now()
	var events []changeEvent

	currentMap := make(map[string]string, len(current))
	for _, f := range current {
		currentMap[f.path] = f.status
	}

	// Detect added or changed
	for path, status := range currentMap {
		oldStatus, existed := s.files[path]
		if !existed {
			events = append(events, changeEvent{path: path, status: status, at: now})
		} else if oldStatus != status {
			events = append(events, changeEvent{path: path, status: status, at: now})
		}
	}

	// Detect removed
	for path := range s.files {
		if _, exists := currentMap[path]; !exists {
			events = append(events, changeEvent{path: path, status: "D", at: now})
		}
	}

	return events
}

// ── Change events ───────────────────────────────────────────

type changeEvent struct {
	path   string
	status string
	at     time.Time
}

// ── Events ring buffer ──────────────────────────────────────

type eventsRing struct {
	items    []changeEvent
	capacity int
}

func newEventsRing(capacity int) eventsRing {
	return eventsRing{
		items:    make([]changeEvent, 0, capacity),
		capacity: capacity,
	}
}

func (r *eventsRing) push(events []changeEvent) {
	r.items = append(r.items, events...)
	if len(r.items) > r.capacity {
		r.items = r.items[len(r.items)-r.capacity:]
	}
}

func (r *eventsRing) recentPaths(within time.Duration) map[string]bool {
	cutoff := time.Now().Add(-within)
	result := make(map[string]bool)
	for _, e := range r.items {
		if e.at.After(cutoff) {
			result[e.path] = true
		}
	}
	return result
}

func (r *eventsRing) count() int {
	return len(r.items)
}

// ── Owl state ───────────────────────────────────────────────

type owlExpression int

const (
	owlIdle  owlExpression = iota // ( o.o )
	owlBlink                      // ( -.- )
	owlWide                       // ( O.O )
	owlRight                      // ( o.o)>
	owlLeft                       // <(o.o )
)

type owlState struct {
	expr    owlExpression
	exprTTL int // ticks remaining in current expression

	// Countdowns to next trigger (in ticks)
	blinkIn  int // 20-50 ticks (2-5s)
	wideIn   int // 100-200 ticks (10-20s)
	glanceIn int // 200-400 ticks (20-40s)
}

func newOwlState() owlState {
	return owlState{
		expr:     owlIdle,
		blinkIn:  20 + rand.Intn(31),  // 20-50
		wideIn:   100 + rand.Intn(101), // 100-200
		glanceIn: 200 + rand.Intn(201), // 200-400
	}
}

func (o *owlState) tick() {
	// Decrement expression TTL
	if o.exprTTL > 0 {
		o.exprTTL--
		if o.exprTTL == 0 {
			o.expr = owlIdle
		}
	}

	// Decrement trigger countdowns
	o.blinkIn--
	o.wideIn--
	o.glanceIn--

	// Fire triggers (priority: wide > glance > blink)
	if o.wideIn <= 0 {
		o.expr = owlWide
		o.exprTTL = 3 + rand.Intn(6) // 3-8 ticks
		o.wideIn = 100 + rand.Intn(101)
	} else if o.glanceIn <= 0 {
		if rand.Intn(2) == 0 {
			o.expr = owlRight
		} else {
			o.expr = owlLeft
		}
		o.exprTTL = 3 + rand.Intn(6) // 3-8 ticks
		o.glanceIn = 200 + rand.Intn(201)
	} else if o.blinkIn <= 0 {
		o.expr = owlBlink
		o.exprTTL = 1 + rand.Intn(2) // 1-2 ticks
		o.blinkIn = 20 + rand.Intn(31)
	}
}

// owlTop returns the top line of the owl sprite (always 7 chars).
func owlTop() string {
	return ` /\_/\ `
}

// owlBottom returns the bottom line based on current expression (always 7 chars).
func (o *owlState) owlBottom() string {
	switch o.expr {
	case owlBlink:
		return "( -.- )"
	case owlWide:
		return "( O.O )"
	case owlRight:
		return "( o.o)>"
	case owlLeft:
		return "<(o.o )"
	default:
		return "( o.o )"
	}
}

// ── Spinner state ───────────────────────────────────────────

var brailleFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type spinnerState struct {
	frame int
}

func (s *spinnerState) tick() {
	s.frame = (s.frame + 1) % len(brailleFrames)
}

func (s *spinnerState) view() string {
	return brailleFrames[s.frame]
}

