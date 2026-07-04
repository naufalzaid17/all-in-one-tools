// Package cryptostream implements password-based, streaming
// authenticated encryption for files of arbitrary size.
//
// Design:
//   - Key derivation: Argon2id (RFC 9106) with parameters stored in the
//     header, so they can be tuned later without breaking old files.
//   - Cipher: AES-256-GCM applied per chunk (default 64 KiB), following
//     the STREAM construction: every chunk's nonce is a random 4-byte
//     prefix plus a 64-bit big-endian counter, and the chunk sequence
//     number together with a final-chunk flag is bound as associated
//     data. Reordering, truncating or duplicating chunks therefore
//     fails authentication.
//
// Memory use is O(chunk size) regardless of input size, which is why
// encryption and decryption operate on io.Reader/io.Writer streams
// backed by temp files rather than byte slices.
package cryptostream

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// File format constants.
var magic = [8]byte{'A', 'I', 'O', 'C', 'R', 'Y', 'P', '1'}

const (
	saltSize        = 16
	noncePrefixSize = 4
	nonceSize       = 12 // noncePrefix (4) + big-endian counter (8)
	keySize         = 32 // AES-256
	// DefaultChunkSize is the plaintext bytes sealed per GCM invocation.
	DefaultChunkSize = 64 * 1024
	maxChunkSize     = 8 << 20
)

// Argon2id defaults (RFC 9106 second recommended option: 64 MiB, 3
// iterations) — strong for interactive use without exhausting small
// machines.
const (
	defaultArgonTime    = 3
	defaultArgonMemory  = 64 * 1024 // KiB
	defaultArgonThreads = 4
)

// ErrWrongPasswordOrCorrupt is returned when authentication fails: the
// password is wrong or the ciphertext was tampered with. The two cases
// are cryptographically indistinguishable.
var ErrWrongPasswordOrCorrupt = errors.New("wrong password or corrupted file")

// ErrNotEncrypted is returned when the input does not carry the
// expected header.
var ErrNotEncrypted = errors.New("input is not an encrypted file (missing header)")

// header is the plaintext preamble of an encrypted file. Everything in
// it is authenticated as associated data of the first chunk, so a
// tampered header fails decryption.
//
// Layout (44 bytes):
//
//	magic(8) | argonTime(4) | argonMemoryKiB(4) | argonThreads(1) |
//	chunkSize(4) | salt(16) | noncePrefix(4) | reserved(3)
type header struct {
	argonTime    uint32
	argonMemory  uint32
	argonThreads uint8
	chunkSize    uint32
	salt         [saltSize]byte
	noncePrefix  [noncePrefixSize]byte
}

const headerSize = 8 + 4 + 4 + 1 + 4 + saltSize + noncePrefixSize + 3

func (h *header) marshal() []byte {
	buf := make([]byte, headerSize)
	copy(buf[0:8], magic[:])
	binary.BigEndian.PutUint32(buf[8:12], h.argonTime)
	binary.BigEndian.PutUint32(buf[12:16], h.argonMemory)
	buf[16] = h.argonThreads
	binary.BigEndian.PutUint32(buf[17:21], h.chunkSize)
	copy(buf[21:21+saltSize], h.salt[:])
	copy(buf[21+saltSize:21+saltSize+noncePrefixSize], h.noncePrefix[:])
	return buf
}

func parseHeader(buf []byte) (*header, error) {
	if len(buf) < headerSize || [8]byte(buf[0:8]) != magic {
		return nil, ErrNotEncrypted
	}
	h := &header{
		argonTime:    binary.BigEndian.Uint32(buf[8:12]),
		argonMemory:  binary.BigEndian.Uint32(buf[12:16]),
		argonThreads: buf[16],
		chunkSize:    binary.BigEndian.Uint32(buf[17:21]),
	}
	copy(h.salt[:], buf[21:21+saltSize])
	copy(h.noncePrefix[:], buf[21+saltSize:21+saltSize+noncePrefixSize])

	if h.argonTime == 0 || h.argonTime > 16 ||
		h.argonMemory == 0 || h.argonMemory > 1<<21 || // ≤ 2 GiB
		h.argonThreads == 0 ||
		h.chunkSize == 0 || h.chunkSize > maxChunkSize {
		return nil, ErrNotEncrypted
	}
	return h, nil
}

func (h *header) deriveKey(password []byte) []byte {
	return argon2.IDKey(password, h.salt[:], h.argonTime, h.argonMemory, h.argonThreads, keySize)
}

func newAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// chunkNonce builds the 12-byte nonce for chunk i.
func chunkNonce(prefix [noncePrefixSize]byte, i uint64) []byte {
	nonce := make([]byte, nonceSize)
	copy(nonce, prefix[:])
	binary.BigEndian.PutUint64(nonce[noncePrefixSize:], i)
	return nonce
}

// chunkAAD binds the header, the chunk sequence number and the
// final-chunk flag into the authentication tag.
func chunkAAD(hdr []byte, i uint64, final bool) []byte {
	aad := make([]byte, len(hdr)+9)
	copy(aad, hdr)
	binary.BigEndian.PutUint64(aad[len(hdr):], i)
	if final {
		aad[len(hdr)+8] = 1
	}
	return aad
}

// Encrypt reads plaintext from r and writes the encrypted container to
// w, deriving the key from password. Memory use is bounded by the chunk
// size.
func Encrypt(w io.Writer, r io.Reader, password string) error {
	h := &header{
		argonTime:    defaultArgonTime,
		argonMemory:  defaultArgonMemory,
		argonThreads: defaultArgonThreads,
		chunkSize:    DefaultChunkSize,
	}
	if _, err := rand.Read(h.salt[:]); err != nil {
		return fmt.Errorf("cryptostream: generate salt: %w", err)
	}
	if _, err := rand.Read(h.noncePrefix[:]); err != nil {
		return fmt.Errorf("cryptostream: generate nonce prefix: %w", err)
	}

	aead, err := newAEAD(h.deriveKey([]byte(password)))
	if err != nil {
		return fmt.Errorf("cryptostream: init cipher: %w", err)
	}

	hdr := h.marshal()
	if _, err := w.Write(hdr); err != nil {
		return fmt.Errorf("cryptostream: write header: %w", err)
	}

	plain := make([]byte, h.chunkSize)
	sealed := make([]byte, 0, int(h.chunkSize)+aead.Overhead())

	// Read ahead one chunk so the final chunk can be flagged in its
	// associated data even when the input length is an exact multiple
	// of the chunk size.
	n, rerr := io.ReadFull(r, plain)
	for i := uint64(0); ; i++ {
		if rerr != nil && !errors.Is(rerr, io.EOF) && !errors.Is(rerr, io.ErrUnexpectedEOF) {
			return fmt.Errorf("cryptostream: read input: %w", rerr)
		}
		current := plain[:n]

		next := make([]byte, h.chunkSize)
		var nextN int
		final := errors.Is(rerr, io.EOF) || errors.Is(rerr, io.ErrUnexpectedEOF)
		if !final {
			nextN, rerr = io.ReadFull(r, next)
			// A follow-up read that yields zero bytes means the
			// current chunk was actually the last one.
			if nextN == 0 && (errors.Is(rerr, io.EOF) || errors.Is(rerr, io.ErrUnexpectedEOF)) {
				final = true
			}
		}

		sealed = aead.Seal(sealed[:0], chunkNonce(h.noncePrefix, i), current, chunkAAD(hdr, i, final))
		if _, err := w.Write(sealed); err != nil {
			return fmt.Errorf("cryptostream: write ciphertext: %w", err)
		}
		if final {
			return nil
		}
		plain, next = next, plain
		n = nextN
	}
}

// Decrypt reads an encrypted container from r, authenticates and
// decrypts it with password, and writes the plaintext to w. It fails
// with ErrWrongPasswordOrCorrupt on any authentication error and never
// writes unauthenticated plaintext.
func Decrypt(w io.Writer, r io.Reader, password string) error {
	hdr := make([]byte, headerSize)
	if _, err := io.ReadFull(r, hdr); err != nil {
		return ErrNotEncrypted
	}
	h, err := parseHeader(hdr)
	if err != nil {
		return err
	}

	aead, err := newAEAD(h.deriveKey([]byte(password)))
	if err != nil {
		return fmt.Errorf("cryptostream: init cipher: %w", err)
	}

	sealedLen := int(h.chunkSize) + aead.Overhead()
	buf := make([]byte, sealedLen)
	next := make([]byte, sealedLen)

	n, rerr := io.ReadFull(r, buf)
	for i := uint64(0); ; i++ {
		if rerr != nil && !errors.Is(rerr, io.EOF) && !errors.Is(rerr, io.ErrUnexpectedEOF) {
			return fmt.Errorf("cryptostream: read input: %w", rerr)
		}
		if n < aead.Overhead() {
			// Even an empty plaintext chunk carries a full tag.
			return ErrWrongPasswordOrCorrupt
		}
		final := errors.Is(rerr, io.EOF) || errors.Is(rerr, io.ErrUnexpectedEOF)

		var nextN int
		if !final {
			nextN, rerr = io.ReadFull(r, next)
			if nextN == 0 && (errors.Is(rerr, io.EOF) || errors.Is(rerr, io.ErrUnexpectedEOF)) {
				final = true
			}
		}

		plain, err := aead.Open(nil, chunkNonce(h.noncePrefix, i), buf[:n], chunkAAD(hdr, i, final))
		if err != nil {
			return ErrWrongPasswordOrCorrupt
		}
		if _, err := w.Write(plain); err != nil {
			return fmt.Errorf("cryptostream: write plaintext: %w", err)
		}
		if final {
			return nil
		}
		buf, next = next, buf
		n = nextN
	}
}
