package utils

import "testing"

func TestHashValue_Success(t *testing.T) {
	hash, err := HashValue("mypassword123")
	if err != nil {
		t.Fatalf("HashValue() unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashValue() returned empty hash")
	}
	if hash == "mypassword123" {
		t.Fatal("HashValue() returned plaintext instead of hash")
	}
}

func TestHashValue_DifferentHashes(t *testing.T) {
	hash1, _ := HashValue("password")
	hash2, _ := HashValue("password")
	if hash1 == hash2 {
		t.Fatal("HashValue() returned identical hashes for same input (bcrypt should use random salt)")
	}
}

func TestCheckValuesHash_Correct(t *testing.T) {
	password := "SecurePass123!"
	hash, err := HashValue(password)
	if err != nil {
		t.Fatalf("HashValue() failed: %v", err)
	}
	if !CheckValuesHash(password, hash) {
		t.Fatal("CheckValuesHash() returned false for correct password")
	}
}

func TestCheckValuesHash_Wrong(t *testing.T) {
	hash, _ := HashValue("correct_password")
	if CheckValuesHash("wrong_password", hash) {
		t.Fatal("CheckValuesHash() returned true for wrong password")
	}
}

func TestCheckValuesHash_InvalidHash(t *testing.T) {
	if CheckValuesHash("password", "not_a_bcrypt_hash") {
		t.Fatal("CheckValuesHash() returned true for invalid hash")
	}
}
