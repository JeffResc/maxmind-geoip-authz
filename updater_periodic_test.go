package main

import (
	"testing"
	"time"
)

// TestPeriodicUpdater verifies that PeriodicUpdater calls the download function
// for every tick received on the ticker channel and exits when the channel is
// closed.
func TestPeriodicUpdater(t *testing.T) {
	// set a fake download function to count calls
	count := 0
	downloadGeoIPDBIfUpdatedFn = func() { count++ }
	defer func() { downloadGeoIPDBIfUpdatedFn = downloadGeoIPDBIfUpdated }()

	// create a controllable ticker channel
	tickCh := make(chan time.Time)
	newTicker = func(d time.Duration) <-chan time.Time { return tickCh }
	defer func() { newTicker = time.Tick }()

	done := make(chan struct{})
	go func() {
		periodicUpdater()
		close(done)
	}()

	// send two ticks and then close the channel to stop the goroutine
	tickCh <- time.Now()
	tickCh <- time.Now()
	close(tickCh)

	<-done
	if count != 2 {
		t.Fatalf("download called %d times, want 2", count)
	}
}
