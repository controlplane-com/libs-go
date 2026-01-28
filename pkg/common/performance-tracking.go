package common

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

type PerformanceSpan struct {
	name         string
	timesInvoked int
	start        time.Time
	duration     time.Duration
	running      bool
	children     []*PerformanceSpan
	m            sync.Mutex
	parent       *PerformanceSpan
}

var outputStream io.Writer = nil

func EnablePerformanceTracking() {
	outputStream = os.Stdout
}

func SetOutputStream(stream io.Writer) {
	outputStream = stream
}

func NewPerformanceSpan(name string) *PerformanceSpan {
	return &PerformanceSpan{name: name}
}

func (p *PerformanceSpan) Start() *PerformanceSpan {
	p.start = time.Now()
	p.running = true
	p.timesInvoked++
	return p
}

func (p *PerformanceSpan) Stop() *PerformanceSpan {
	if !p.running {
		return p
	}
	p.duration += time.Now().Sub(p.start)
	p.running = false
	for _, c := range p.children {
		if c == nil {
			continue
		}
		c.Stop()
	}
	if p.parent == nil && outputStream != nil {
		io.WriteString(outputStream, p.Summarize()+"\n")
	}
	return p
}

func (p *PerformanceSpan) Duration() time.Duration {
	if p.running {
		return time.Now().Sub(p.start)
	}
	return p.duration
}

func (p *PerformanceSpan) Add(childName string, start bool) *PerformanceSpan {
	child := NewPerformanceSpan(childName)
	if start {
		child.Start()
	}
	p.m.Lock()
	defer p.m.Unlock()
	p.children = append(p.children, child)
	child.parent = p
	return child
}

func (p *PerformanceSpan) AddChild(child *PerformanceSpan) {
	p.m.Lock()
	defer p.m.Unlock()
	child.parent = p
	p.children = append(p.children, child)
}

// Link is deprecated - use AddChild instead
func (p *PerformanceSpan) Link(other *PerformanceSpan) {
	p.AddChild(other)
}

func (p *PerformanceSpan) Summarize() string {
	summary := p.summarize()
	summary["name"] = p.name
	b, _ := json.MarshalIndent(summary, "", "   ")
	return string(b)
}

func (p *PerformanceSpan) summarize() map[string]any {
	summary := map[string]any{}
	summary["runTime"] = p.Duration().Seconds()
	summary["name"] = p.name
	summary["timesInvoked"] = p.timesInvoked
	var childSlice []map[string]any
	for _, c := range p.children {
		childSlice = append(childSlice, c.summarize())
	}
	if len(childSlice) > 0 {
		summary["children"] = childSlice
	}
	return summary
}

// ContextWithSpan stores a PerformanceSpan in the context, or adds a child span to the existing span in the context if one exists
// Returns a new context with the span
func ContextWithSpan(ctx context.Context, span *PerformanceSpan) context.Context {
	parent := SpanFromContext(ctx)
	if parent == nil {
		// No parent span in context, create a new root span
		return context.WithValue(ctx, PerformanceSpanKey, span.Start())
	}
	parent.AddChild(span.Start())
	return context.WithValue(ctx, PerformanceSpanKey, span)
}

// SpanFromContext retrieves a PerformanceSpan from the context
// Returns nil if no span is found in the context
func SpanFromContext(ctx context.Context) *PerformanceSpan {
	if span, ok := ctx.Value(PerformanceSpanKey).(*PerformanceSpan); ok {
		return span
	}
	return nil
}

// AddChildSpan retrieves the parent span from context, adds a child span to it,
// starts the child span, and returns a new context with the child span
// If no parent span exists in context, creates a new root span
func AddChildSpan(ctx context.Context, name string) (context.Context, *PerformanceSpan) {
	span := NewPerformanceSpan(name)
	return ContextWithSpan(ctx, span), span
}
