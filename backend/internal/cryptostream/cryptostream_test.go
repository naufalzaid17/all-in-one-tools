package cryptostream

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"
)

func roundTrip(t *testing.T, plaintext []byte, password string) []byte {
	t.Helper()
	var enc bytes.Buffer
	if err := Encrypt(&enc, bytes.NewReader(plaintext), password); err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	var dec bytes.Buffer
	if err := Decrypt(&dec, bytes.NewReader(enc.Bytes()), password); err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(dec.Bytes(), plaintext) {
		t.Fatalf("round trip mismatch: got %d bytes, want %d", dec.Len(), len(plaintext))
	}
	return enc.Bytes()
}

func TestRoundTripSizes(t *testing.T) {
	sizes := []int{
		0, 1, 100,
		DefaultChunkSize - 1, DefaultChunkSize, DefaultChunkSize + 1,
		2*DefaultChunkSize - 1, 2 * DefaultChunkSize, 2*DefaultChunkSize + 1,
		3*DefaultChunkSize + 12345,
	}
	for _, size := range sizes {
		plaintext := make([]byte, size)
		if _, err := rand.Read(plaintext); err != nil {
			t.Fatal(err)
		}
		roundTrip(t, plaintext, "correct horse battery staple")
	}
}

func TestWrongPassword(t *testing.T) {
	enc := roundTrip(t, []byte("secret payload"), "right-password")
	var out bytes.Buffer
	err := Decrypt(&out, bytes.NewReader(enc), "wrong-password")
	if !errors.Is(err, ErrWrongPasswordOrCorrupt) {
		t.Fatalf("want ErrWrongPasswordOrCorrupt, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatal("plaintext must not be written on auth failure")
	}
}

func TestTamperedCiphertext(t *testing.T) {
	enc := roundTrip(t, bytes.Repeat([]byte("x"), DefaultChunkSize+7), "pw")
	// Flip one bit in the middle of the first chunk's ciphertext.
	enc[headerSize+10] ^= 0x01
	var out bytes.Buffer
	if err := Decrypt(&out, bytes.NewReader(enc), "pw"); !errors.Is(err, ErrWrongPasswordOrCorrupt) {
		t.Fatalf("want ErrWrongPasswordOrCorrupt, got %v", err)
	}
}

func TestTamperedHeader(t *testing.T) {
	enc := roundTrip(t, []byte("data"), "pw")
	enc[9] ^= 0x01 // argonTime byte — authenticated as AAD
	var out bytes.Buffer
	if err := Decrypt(&out, bytes.NewReader(enc), "pw"); err == nil {
		t.Fatal("tampered header must fail")
	}
}

func TestTruncatedFinalChunk(t *testing.T) {
	enc := roundTrip(t, bytes.Repeat([]byte("y"), 2*DefaultChunkSize), "pw")
	truncated := enc[:len(enc)-20]
	var out bytes.Buffer
	if err := Decrypt(&out, bytes.NewReader(truncated), "pw"); !errors.Is(err, ErrWrongPasswordOrCorrupt) {
		t.Fatalf("want ErrWrongPasswordOrCorrupt, got %v", err)
	}
}

func TestRemovedFinalChunk(t *testing.T) {
	// Dropping the whole final chunk must fail because the new last
	// chunk was not sealed with the final flag.
	plaintext := bytes.Repeat([]byte("z"), 2*DefaultChunkSize+50)
	enc := roundTrip(t, plaintext, "pw")
	sealedLen := DefaultChunkSize + 16
	cut := headerSize + 2*sealedLen // keep exactly two full chunks
	var out bytes.Buffer
	if err := Decrypt(&out, bytes.NewReader(enc[:cut]), "pw"); !errors.Is(err, ErrWrongPasswordOrCorrupt) {
		t.Fatalf("want ErrWrongPasswordOrCorrupt, got %v", err)
	}
}

func TestNotEncryptedInput(t *testing.T) {
	var out bytes.Buffer
	if err := Decrypt(&out, bytes.NewReader([]byte("plain old file")), "pw"); !errors.Is(err, ErrNotEncrypted) {
		t.Fatalf("want ErrNotEncrypted, got %v", err)
	}
}
