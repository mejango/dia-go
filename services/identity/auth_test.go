package dia

import (
	"bytes"
	"strings"
	"testing"
)

func TestPassword(t *testing.T) {
	hash, e := HashPassword("correct horse")
	if e != nil {
		t.Fatal(e)
	}
	again, e := HashPassword("correct horse")
	if e != nil {
		t.Fatal(e)
	}
	if hash == again {
		t.Fatal("salt reused")
	}
	if !CheckPassword("correct horse", hash) || CheckPassword("wrong", hash) {
		t.Fatal("password check")
	}
	p := strings.Split(hash, "$")
	for _, bad := range []string{"", hash + "$extra", strings.Replace(hash, p[1], "999999999999", 1), strings.Replace(hash, p[1], "1", 1), strings.Replace(hash, p[2], "AA", 1), strings.Replace(hash, p[3], "AA", 1), strings.Replace(hash, "sha256", "sha1", 1)} {
		if CheckPassword("correct horse", bad) {
			t.Fatal("accepted bad hash")
		}
	}
	for _, bad := range []string{"", strings.Repeat("a", 4097)} {
		if _, e = HashPassword(bad); e == nil {
			t.Fatal("password bounds")
		}
		if CheckPassword(bad, hash) {
			t.Fatal("check bounds")
		}
	}
}

func TestSeal(t *testing.T) {
	key := bytes.Repeat([]byte{7}, 32)
	plain := []byte("secret")
	a, e := Seal(key, plain)
	if e != nil {
		t.Fatal(e)
	}
	b, e := Seal(key, plain)
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Equal(a, b) {
		t.Fatal("nonce reused")
	}
	got, e := Open(key, a)
	if e != nil || !bytes.Equal(got, plain) {
		t.Fatal("roundtrip", e)
	}
	a[len(a)-1] ^= 1
	for _, bad := range [][]byte{nil, {}, a, a[:10]} {
		if got, e = Open(key, bad); e == nil || got != nil {
			t.Fatal("accepted corruption")
		}
	}
	if _, e = Open(bytes.Repeat([]byte{8}, 32), b); e == nil {
		t.Fatal("accepted wrong key")
	}
	if _, e = Seal([]byte("bad"), plain); e == nil {
		t.Fatal("accepted bad key")
	}
	empty, e := Seal(key, nil)
	if e != nil {
		t.Fatal(e)
	}
	if got, e = Open(key, empty); e != nil || len(got) != 0 {
		t.Fatal("empty roundtrip")
	}
}

func TestPasswordKnownAnswer(t *testing.T) {
	// Independently generated with Python hashlib.pbkdf2_hmac, SHA-256, 600000 rounds.
	const encoded = "pbkdf2-sha256$600000$AAECAwQFBgcICQoLDA0ODw$O8NxGOYlCT6bee0Ikw6nr3OJWRIz_dkt3fNpNx5g28A"
	if !CheckPassword("password", encoded) {
		t.Fatal("PBKDF2 known answer mismatch")
	}
	if CheckPassword("password", strings.Replace(encoded, "AAEC", "AA\nEC", 1)) {
		t.Fatal("noncanonical salt accepted")
	}
}
