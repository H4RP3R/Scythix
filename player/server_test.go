package player

import (
	"testing"
	"time"
)

func TestPlayerServer_Pause(t *testing.T) {
	playlistDir := "."
	srv := NewPlayerServer(playlistDir)

	if srv.ctrl.Paused {
		t.Fatalf("Invalid initial value: srv.ctrl.Paused=%t", srv.ctrl.Paused)
	}

	err := srv.Pause(&struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Pause() returned error: %v", err)
	}
	if !srv.ctrl.Paused {
		t.Errorf("Pause did not toggle pause state to true")
	}

	// Call again to toggle back
	err = srv.Pause(&struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Pause() returned error: %v", err)
	}
	if srv.ctrl.Paused {
		t.Errorf("Pause did not toggle pause state back to false")
	}
}

func TestPlayerServer_Stop(t *testing.T) {
	srv := NewPlayerServer(".")

	if srv.done == nil {
		t.Fatal("Channel done not initialized")
	}

	if err := srv.Stop(&struct{}{}, &struct{}{}); err != nil {
		t.Fatalf("Stop() returned error: %v", err)
	}

	select {
	case <-srv.done:
		// Channel closed, test passes
	case <-time.After(time.Second):
		t.Error("Timeout waiting for done channel to close")
	}

	// Calling Stop again should not panic and succeed
	if err := srv.Stop(&struct{}{}, &struct{}{}); err != nil {
		t.Fatalf("Second Stop() call returned error: %v", err)
	}
}

func TestPlayerServer_Mute(t *testing.T) {
	playlistDir := "."
	srv := NewPlayerServer(playlistDir)

	if srv.vol.Silent {
		t.Fatalf("Invalid initial value: srv.vol.Silent=%t", srv.vol.Silent)
	}

	err := srv.Mute(&struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Mute() returned error: %v", err)
	}
	if !srv.vol.Silent {
		t.Errorf("Mute did not toggle mute state to true")
	}

	err = srv.Mute(&struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Mute() returned error: %v", err)
	}
	if srv.vol.Silent {
		t.Errorf("Mute did not toggle mute state back to false")
	}
}
