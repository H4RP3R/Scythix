package player

import (
	"testing"
	"time"
)

func TestPlayerServer_Pause(t *testing.T) {
	playlistDir := "."
	srv := NewPlayerServer(playlistDir)

	if srv.ctrl.Paused {
		t.Fatalf("invalid initial value: srv.ctrl.Paused=%t", srv.ctrl.Paused)
	}

	err := srv.Pause(&struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Pause() returned error: %v", err)
	}
	if !srv.ctrl.Paused {
		t.Errorf("Pause() did not toggle pause state to true")
	}

	// Call again to toggle back
	err = srv.Pause(&struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Pause() returned error: %v", err)
	}
	if srv.ctrl.Paused {
		t.Errorf("Pause() did not toggle pause state back to false")
	}
}

func TestPlayerServer_Stop(t *testing.T) {
	srv := NewPlayerServer(".")

	if srv.done == nil {
		t.Fatal("channel done not initialized")
	}

	if err := srv.Stop(&struct{}{}, &struct{}{}); err != nil {
		t.Fatalf("Stop() returned error: %v", err)
	}

	select {
	case <-srv.done:
		// Channel closed, test passes
	case <-time.After(time.Second):
		t.Error("timeout waiting for done channel to close")
	}

	// Calling Stop again should not panic and succeed
	if err := srv.Stop(&struct{}{}, &struct{}{}); err != nil {
		t.Fatalf("second Stop() call returned error: %v", err)
	}
}

func TestPlayerServer_Mute(t *testing.T) {
	playlistDir := "."
	srv := NewPlayerServer(playlistDir)

	if srv.vol.Silent {
		t.Fatalf("invalid initial value: srv.vol.Silent=%t", srv.vol.Silent)
	}

	err := srv.Mute(&struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Mute() returned error: %v", err)
	}
	if !srv.vol.Silent {
		t.Errorf("Mute() did not toggle mute state to true")
	}

	err = srv.Mute(&struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Mute() returned error: %v", err)
	}
	if srv.vol.Silent {
		t.Errorf("Mute() did not toggle mute state back to false")
	}
}

func TestPlayerServer_TurnUp(t *testing.T) {
	var initVol float64 = mapScaleToVolume(13)
	var initSilent bool = true
	playlistDir := "."

	srv := NewPlayerServer(playlistDir)
	srv.vol.Volume = initVol
	// Mute the volume to verify that TurnUp() correctly unmutes it
	srv.vol.Silent = initSilent

	scaleExpect := mapVolumeToScale(initVol + volStep)
	var currentScale float64
	err := srv.TurnUp(&struct{}{}, &currentScale)
	if err != nil {
		t.Errorf("TurnUp() returned error: %v", err)
	}
	if srv.vol.Silent {
		t.Errorf("TurnUp() failed to unmute volume: srv.vol.Silent=%t", srv.vol.Silent)
	}
	if currentScale != scaleExpect {
		t.Errorf("want currentScale=%v, got %v", scaleExpect, currentScale)
	}

	// Increase volume to maximum
	steps := int((volLimitMax - mapScaleToVolume(currentScale)) / volStep)
	prevScale := currentScale
	for i := 0; i < steps; i++ {
		err = srv.TurnUp(&struct{}{}, &currentScale)
		if err != nil {
			t.Errorf("TurnUp() returned error: %v", err)
		}
		if currentScale <= prevScale {
			t.Errorf("TurnUp() failed to increase volume: prevScale=%v, currentScale=%v", prevScale, currentScale)
		}
		prevScale = currentScale
	}

	// Try to increase the volume above the maximum
	err = srv.TurnUp(&struct{}{}, &currentScale)
	if err != nil {
		t.Errorf("TurnUp() returned error increasing above maximum: %v", err)
	}
	currentVol := mapScaleToVolume(currentScale)
	if currentVol > volLimitMax {
		t.Errorf("TurnUp() exceeded the maximum volume: want currentScale<=%v, got %v", volLimitMax, currentVol)
	}
}

func TestPlayerServer_TurnDown(t *testing.T) {
	var initVol float64 = mapScaleToVolume(16)
	var initSilent bool = true
	playlistDir := "."

	srv := NewPlayerServer(playlistDir)
	srv.vol.Volume = initVol
	// TurnDown() like TurnUp() implies unmuting the volume
	// Unmuting should also be tested.
	srv.vol.Silent = initSilent

	scaleExpect := mapVolumeToScale(initVol - volStep)
	var currentScale float64
	err := srv.TurnDown(&struct{}{}, &currentScale)
	if err != nil {
		t.Errorf("TurnDown() returned error: %v", err)
	}
	if srv.vol.Silent {
		t.Errorf("TurnDown() failed to unmute volume: srv.vol.Silent=%t", srv.vol.Silent)
	}
	if currentScale != scaleExpect {
		t.Errorf("want currentScale=%v, got %v", scaleExpect, currentScale)
	}

	// Turn down the volume to the lowest limit
	steps := int((mapScaleToVolume(currentScale) - volLimitMin) / volStep)
	prevScale := currentScale
	for i := 0; i < steps; i++ {
		err = srv.TurnDown(&struct{}{}, &currentScale)
		if err != nil {
			t.Errorf("TurnDown() returned error: %v", err)
		}
		if currentScale >= prevScale {
			t.Errorf("TurnDown() failed to reduce volume: prevScale=%v, currentScale=%v", prevScale, currentScale)
		}
		prevScale = currentScale
	}

	// Try to reduce the volume bellow the minimum
	err = srv.TurnDown(&struct{}{}, &currentScale)
	if err != nil {
		t.Errorf("TurnDown() returned error when decreasing below minimum: %v", err)
	}
	currentVol := mapScaleToVolume(currentScale)
	if currentVol < volLimitMin {
		t.Errorf("volume decreased below minimum: want currentScale<=%v, got %v", volLimitMin, currentVol)
	}
}
