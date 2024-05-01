package main

import (
	"testing"
)

func TestTify(t *testing.T) {
	state := "?manifest=https%3A%2F%2Farchiginnasio.jarvis.memooria.org%2Fmeta%2Fiiif%2Fb5ea7501-9bbb-4435-97ce-a2f7061316aa%2Fmanifest&tify=%7B%22pages%22%3A%5B23%5D%2C%22pan%22%3A%7B%22x%22%3A0.304%2C%22y%22%3A0.534%7D%2C%22view%22%3A%22thumbnails%22%2C%22zoom%22%3A1.728%7D"
	expectedStates := States{
		{Manifest: "https://archiginnasio.jarvis.memooria.org/meta/iiif/b5ea7501-9bbb-4435-97ce-a2f7061316aa/manifest", Options: []byte("{\"pages\":[23],\"pan\":{\"x\":0.304,\"y\":0.534},\"view\":\"thumbnails\",\"zoom\":1.728}")},
	}

	states, err := Tify(state)
	if err != nil {
		t.Errorf("Tify returned an error: %v", err)
	}

	if len(states) != len(expectedStates) {
		t.Errorf("Expected %d states, but got %d", len(expectedStates), len(states))
	}

	for i, expectedState := range expectedStates {
		if states[i].Manifest != expectedState.Manifest {
			t.Errorf("Expected manifest %s, but got %s", expectedState.Manifest, states[i].Manifest)
		}
		if string(states[i].Options) != string(expectedState.Options) {
			t.Errorf("Expected options %s, but got %s", string(expectedState.Options), string(states[i].Options))
		}
	}
}
