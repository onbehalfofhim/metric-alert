package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	errEmptyFilePath = errors.New("file path is empty")
)

// LoadPublicKey загружает публичный ключ из файла
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	if path == "" {
		return nil, errEmptyFilePath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pubIfc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pub, ok := pubIfc.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key is not RSA public key")
	}

	return pub, nil
}

// LoadPrivateKey загружает приватный ключ из файла
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		return nil, errEmptyFilePath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// PKCS#1
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// PKCS#8
	keyIfc, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := keyIfc.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("key is not RSA private key")
	}

	return key, nil
}

// EncryptRSA шифрует данные произвольного размера с помощью гибридной схемы:
// 1. Генерируется случайный AES-ключ.
// 2. Данные шифруются с помощью AES-GCM.
// 3. AES-ключ шифруется публичным RSA-ключом.
// Результат имеет формат:
// [2 байта длины RSA-ключа][RSA(AES key)][nonce][ciphertext]
// Так как RSA не подходит для шифрования больших объёмов данных.
func EncryptRSA(data []byte, pub *rsa.PublicKey) ([]byte, error) {
	// Генерируем случайный 256-битный AES-ключ.
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Создаем AES-блок.
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Создаем AEAD-шифр в режиме GCM.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-GCM: %w", err)
	}

	// Генерируем случайный nonce.
	// Для GCM nonce должен быть уникальным для каждого сообщения.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Шифруем исходные данные с помощью AES-GCM.
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Шифруем AES-ключ с помощью публичного RSA-ключа.
	encryptedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pub,
		aesKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	// Формируем итоговый пакет:
	// [2 байта длины RSA-ключа]
	result := make([]byte, 2)
	binary.BigEndian.PutUint16(result, uint16(len(encryptedKey)))

	// [RSA(AES key)]
	result = append(result, encryptedKey...)
	// [nonce]
	result = append(result, nonce...)
	// [AES(ciphertext)]
	result = append(result, ciphertext...)

	return result, nil
}

// DecryptRSA расшифровывает данные, зашифрованные функцией EncryptRSA.
// Ожидаемый формат входных данных:
// [2 байта длины RSA-ключа][RSA(AES key)][nonce][ciphertext]
func DecryptRSA(data []byte, priv *rsa.PrivateKey) ([]byte, error) {
	// Проверяем, что пакет содержит хотя бы поле длины ключа.
	if len(data) < 2 {
		return nil, fmt.Errorf("invalid encrypted payload")
	}

	// Извлекаем длину RSA-зашифрованного AES-ключа.
	keyLen := int(binary.BigEndian.Uint16(data[:2]))
	if len(data) < 2+keyLen {
		return nil, fmt.Errorf("invalid encrypted payload")
	}

	// Извлекаем RSA-зашифрованный AES-ключ.
	encryptedKey := data[2 : 2+keyLen]

	// Расшифровываем AES-ключ приватным RSA-ключом.
	aesKey, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		priv,
		encryptedKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	// Создаем AES-блок.
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Создаем GCM-дешифратор.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-GCM: %w", err)
	}

	// Определяем размер nonce.
	nonceSize := gcm.NonceSize()
	if len(data) < 2+keyLen+nonceSize {
		return nil, fmt.Errorf("invalid encrypted payload")
	}

	// Извлекаем nonce.
	nonceStart := 2 + keyLen
	nonceEnd := nonceStart + nonceSize
	nonce := data[nonceStart:nonceEnd]

	// Извлекаем зашифрованные данные.
	ciphertext := data[nonceEnd:]

	// Расшифровываем данные.
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	return plaintext, nil
}
