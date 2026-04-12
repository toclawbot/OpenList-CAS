package _189pc

import "testing"

func TestShouldRestoreSourceFromCAS(t *testing.T) {
	driver := &Cloud189PC{Addition: Addition{RestoreSourceFromCAS: true}}
	if !driver.shouldRestoreSourceFromCAS("movie.mkv.cas") {
		t.Fatal("expected .cas restore to be enabled")
	}
	if driver.shouldRestoreSourceFromCAS("movie.mkv") {
		t.Fatal("did not expect non-.cas file to trigger restore")
	}
}
