package connections

import "testing"

// newS3Harness can be registered by the Go implementation to run this parity
// suite against real code. Tests fail until this harness is registered.
var newS3Harness func(t *testing.T) s3Harness

type s3Harness interface {
	SaveImage(id string, imageBlob []byte) error
	GetImage(id string) ([]byte, error)
	ReturnImage(id string) ([]byte, error)
	LastPut() map[string]any
	LastGet() map[string]any
}

func requireS3Harness(t *testing.T) s3Harness {
	t.Helper()
	if newS3Harness == nil {
		t.Fatalf("s3 parity harness is not registered: register newS3Harness in implementation tests")
	}
	return newS3Harness(t)
}

func TestS3Manager_SaveImage_RequestShape(t *testing.T) {
	h := requireS3Harness(t)

	if err := h.SaveImage("123", []byte("png-data")); err != nil {
		t.Fatalf("SaveImage() error = %v", err)
	}

	put := h.LastPut()
	if put["key"] != "123.png" {
		t.Fatalf("put key = %v, want 123.png", put["key"])
	}
	if put["content_type"] != "image/png" {
		t.Fatalf("put content_type = %v, want image/png", put["content_type"])
	}
}

func TestS3Manager_GetImage_RequestShape(t *testing.T) {
	h := requireS3Harness(t)

	if _, err := h.GetImage("abc"); err != nil {
		t.Fatalf("GetImage() error = %v", err)
	}

	get := h.LastGet()
	if get["key"] != "abc.png" {
		t.Fatalf("get key = %v, want abc.png", get["key"])
	}
	if get["content_type"] != "image/png" {
		t.Fatalf("get content_type = %v, want image/png", get["content_type"])
	}
}

func TestS3Manager_ReturnImage_ForwardsBody(t *testing.T) {
	h := requireS3Harness(t)

	if err := h.SaveImage("xyz", []byte("blob")); err != nil {
		t.Fatalf("SaveImage() error = %v", err)
	}
	got, err := h.ReturnImage("xyz")
	if err != nil {
		t.Fatalf("ReturnImage() error = %v", err)
	}
	if string(got) != "blob" {
		t.Fatalf("ReturnImage() = %q, want %q", string(got), "blob")
	}
}
