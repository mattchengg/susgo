package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"
)

// ProgressReporter is a common interface for progress reporting
// Used by both CLI (ProgressBar) and GUI (GUIProgressReporter)
type ProgressReporter interface {
	SetTotal(total int64)
	SetCurrent(current int64)
	Add(delta int64)
	SetStatus(message string)
	Finish()
}

type ProgressBar struct {
	total      int64
	current    int64
	width      int
	startTime  time.Time
	lastUpdate time.Time
	mu         sync.Mutex
	done       chan struct{}
}

func NewProgressBar(total int64) *ProgressBar {
	return &ProgressBar{
		total:     total,
		width:     40,
		startTime: time.Now(),
		done:      make(chan struct{}),
	}
}

func (p *ProgressBar) Start() {
	go p.render()
}

func (p *ProgressBar) SetTotal(total int64) {
	p.mu.Lock()
	p.total = total
	p.mu.Unlock()
}

func (p *ProgressBar) Add(n int64) {
	p.mu.Lock()
	p.current += n
	p.mu.Unlock()
}

func (p *ProgressBar) SetCurrent(n int64) {
	p.mu.Lock()
	p.current = n
	p.mu.Unlock()
}

func (p *ProgressBar) SetStatus(message string) {
	// CLI progress bar doesn't display status messages separately
	// This is a no-op for CLI compatibility with ProgressReporter interface
}

func (p *ProgressBar) Finish() {
	close(p.done)
	p.printBar()
	fmt.Println()
}

func (p *ProgressBar) render() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.printBar()
		}
	}
}

func (p *ProgressBar) printBar() {
	p.mu.Lock()
	current := p.current
	total := p.total
	p.mu.Unlock()

	if total <= 0 {
		return
	}

	pct := float64(current) / float64(total)
	filled := int(pct * float64(p.width))
	if filled > p.width {
		filled = p.width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	elapsed := time.Since(p.startTime).Seconds()
	speed := float64(current) / elapsed
	if elapsed < 0.1 {
		speed = 0
	}

	var eta string
	if speed > 0 {
		remaining := float64(total-current) / speed
		if remaining < 60 {
			eta = fmt.Sprintf("%.0fs", remaining)
		} else if remaining < 3600 {
			eta = fmt.Sprintf("%.0fm%.0fs", remaining/60, float64(int(remaining)%60))
		} else {
			eta = fmt.Sprintf("%.0fh%.0fm", remaining/3600, float64(int(remaining)%3600)/60)
		}
	} else {
		eta = "--"
	}

	fmt.Printf("\r[%s] %5.1f%% %s/%s %s/s ETA %s  ",
		bar,
		pct*100,
		formatSize(current),
		formatSize(total),
		formatSize(int64(speed)),
		eta,
	)
}

func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// GUIProgressReporter implements ProgressReporter for Fyne GUI
type GUIProgressReporter struct {
	bar       *widget.ProgressBar
	label     *widget.Label
	total     int64
	current   int64
	startTime time.Time
}

func NewGUIProgressReporter(bar *widget.ProgressBar, label *widget.Label) *GUIProgressReporter {
	return &GUIProgressReporter{
		bar:       bar,
		label:     label,
		startTime: time.Now(),
	}
}

func (g *GUIProgressReporter) SetTotal(total int64) {
	g.total = total
	g.bar.Max = float64(total)
}

func (g *GUIProgressReporter) SetCurrent(current int64) {
	g.current = current
	g.bar.SetValue(float64(current))
	g.updateLabel()
}

func (g *GUIProgressReporter) Add(delta int64) {
	g.current += delta
	g.bar.SetValue(float64(g.current))
	g.updateLabel()
}

func (g *GUIProgressReporter) SetStatus(message string) {
	g.label.SetText(message)
}

func (g *GUIProgressReporter) Finish() {
	g.bar.SetValue(float64(g.total))
	g.label.SetText("✅ Complete!")
}

func (g *GUIProgressReporter) updateLabel() {
	if g.total <= 0 {
		return
	}
	pct := float64(g.current) / float64(g.total) * 100
	elapsed := time.Since(g.startTime).Seconds()
	speed := float64(g.current) / elapsed

	var eta string
	if speed > 0 {
		remaining := float64(g.total-g.current) / speed
		if remaining < 60 {
			eta = fmt.Sprintf("%.0fs", remaining)
		} else {
			eta = fmt.Sprintf("%.0fm", remaining/60)
		}
	} else {
		eta = "calculating..."
	}

	g.label.SetText(fmt.Sprintf("%.1f%% - %s/%s - %s/s - ETA: %s",
		pct,
		formatSize(g.current),
		formatSize(g.total),
		formatSize(int64(speed)),
		eta))
}
