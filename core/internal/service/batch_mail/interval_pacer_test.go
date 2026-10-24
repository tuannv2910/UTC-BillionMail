package batch_mail

import (
	"context"
	"testing"
	"time"
)

func TestIntervalPacer_SpacesSends(t *testing.T) {
	p := newIntervalPacer(40 * time.Millisecond)
	ctx := context.Background()

	start := time.Now()
	for i := 0; i < 4; i++ {
		if err := p.Wait(ctx); err != nil {
			t.Fatal(err)
		}
	}
	// first send is immediate, the next three wait 40ms each
	if elapsed := time.Since(start); elapsed < 110*time.Millisecond {
		t.Errorf("sends were not spaced: %s", elapsed)
	}
}

func TestIntervalPacer_NilNeverWaits(t *testing.T) {
	var p *intervalPacer = newIntervalPacer(0)
	if p != nil {
		t.Fatal("zero interval should disable pacing")
	}
	start := time.Now()
	for i := 0; i < 100; i++ {
		_ = p.Wait(context.Background())
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Error("nil pacer should not wait")
	}
}

func TestIntervalPacer_ContextCancel(t *testing.T) {
	p := newIntervalPacer(time.Hour)
	_ = p.Wait(context.Background()) // consume the immediate slot

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := p.Wait(ctx); err == nil {
		t.Error("expected context error")
	}
}

func TestNormalizeSendInterval(t *testing.T) {
	cases := map[int]int{-5: 0, 0: 0, 30: 30, 99999: MaxSendIntervalSeconds}
	for in, want := range cases {
		if got := normalizeSendInterval(in); got != want {
			t.Errorf("normalizeSendInterval(%d) = %d, want %d", in, got, want)
		}
	}
}
