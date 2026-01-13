package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

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
