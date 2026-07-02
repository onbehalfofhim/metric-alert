package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func generateKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	return priv, &priv.PublicKey
}

func writePublicKey(t *testing.T, path string, pub *rsa.PublicKey) {
	t.Helper()

	data, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: data,
	}

	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
		t.Fatalf("write public key: %v", err)
	}
}

func writePrivateKey(t *testing.T, path string, priv *rsa.PrivateKey) {
	t.Helper()

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}

	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
}

func TestLoadPublicKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()

		_, pub := generateKeys(t)

		path := filepath.Join(dir, "public.pem")
		writePublicKey(t, path, pub)

		got, err := LoadPublicKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.N.Cmp(pub.N) != 0 {
			t.Fatal("loaded wrong public key")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		_, err := LoadPublicKey("")
		if !errors.Is(err, errEmptyFilePath) {
			t.Fatalf("expected errEmptyFilePath, got %v", err)
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		_, err := LoadPublicKey("unknown.pem")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid pem", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.pem")

		err := os.WriteFile(path, []byte("not pem"), 0o600)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadPublicKey(path)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestLoadPrivateKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()

		priv, _ := generateKeys(t)

		path := filepath.Join(dir, "private.pem")
		writePrivateKey(t, path, priv)

		got, err := LoadPrivateKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.N.Cmp(priv.N) != 0 {
			t.Fatal("loaded wrong private key")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		_, err := LoadPrivateKey("")
		if !errors.Is(err, errEmptyFilePath) {
			t.Fatalf("expected errEmptyFilePath, got %v", err)
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		_, err := LoadPrivateKey("unknown.pem")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid pem", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.pem")

		err := os.WriteFile(path, []byte("not pem"), 0o600)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadPrivateKey(path)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestEncryptDecryptRSA(t *testing.T) {
	priv, pub := generateKeys(t)

	src := []byte("very important secret message")

	encrypted, err := EncryptRSA(src, pub)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if bytes.Equal(src, encrypted) {
		t.Fatal("encrypted data must differ from source")
	}

	decrypted, err := DecryptRSA(encrypted, priv)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if !bytes.Equal(src, decrypted) {
		t.Fatalf("expected %q, got %q", src, decrypted)
	}
}

func TestEncryptDecryptRSA_LargePayload(t *testing.T) {
	priv, pub := generateKeys(t)

	src := bytes.Repeat([]byte("metric"), 10000)

	encrypted, err := EncryptRSA(src, pub)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	decrypted, err := DecryptRSA(encrypted, priv)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if !bytes.Equal(src, decrypted) {
		t.Fatal("payload mismatch")
	}
}

func TestDecryptRSA_InvalidPayload(t *testing.T) {
	priv, _ := generateKeys(t)

	tests := [][]byte{
		nil,
		[]byte{},
		[]byte{1},
		[]byte{0, 10},
	}

	for _, tt := range tests {
		_, err := DecryptRSA(tt, priv)
		if err == nil {
			t.Fatalf("expected error for payload %v", tt)
		}
	}
}

func TestDecryptRSA_CorruptedCiphertext(t *testing.T) {
	priv, pub := generateKeys(t)

	data := []byte("secret")

	encrypted, err := EncryptRSA(data, pub)
	if err != nil {
		t.Fatal(err)
	}

	encrypted[len(encrypted)-1] ^= 0xFF

	_, err = DecryptRSA(encrypted, priv)
	if err == nil {
		t.Fatal("expected decrypt error")
	}
}
