package sessions

import "testing"

func TestValidStatus(t *testing.T) {
	for _, status := range []Status{StatusDraft, StatusActive, StatusCompleted} {
		if !validStatus(status) {
			t.Errorf("expected %q to be valid", status)
		}
	}
	if validStatus("archived") {
		t.Fatal("expected archived to be invalid")
	}
}

func TestAllowedContentType(t *testing.T) {
	if !allowedContentType("image", "image/png; charset=binary") {
		t.Fatal("expected image/png to be allowed")
	}
	if allowedContentType("image", "audio/mpeg") {
		t.Fatal("expected audio/mpeg to be rejected for images")
	}
	if !allowedContentType("audio", "audio/mpeg") {
		t.Fatal("expected audio/mpeg to be allowed")
	}
}
