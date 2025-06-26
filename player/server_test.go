package player

import (
	"testing"
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
