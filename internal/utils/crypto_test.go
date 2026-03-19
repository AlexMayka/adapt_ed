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

// ── GeneratePassword ────────────────────────────────────────────────────────

func TestGeneratePassword_Length(t *testing.T) {
	for _, length := range []int{8, 16, 32} {
		password, err := GeneratePassword(length)
		if err != nil {
			t.Fatalf("GeneratePassword(%d) error: %v", length, err)
		}
		if len(password) != length {
			t.Fatalf("GeneratePassword(%d) len = %d", length, len(password))
		}
	}
}

func TestGeneratePassword_Unique(t *testing.T) {
	p1, _ := GeneratePassword(16)
	p2, _ := GeneratePassword(16)
	if p1 == p2 {
		t.Fatal("GeneratePassword() returned identical passwords")
	}
}

func TestGeneratePassword_ValidChars(t *testing.T) {
	password, _ := GeneratePassword(100)
	for _, ch := range password {
		found := false
		for _, valid := range passwordChars {
			if ch == valid {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("GeneratePassword() produced invalid char: %q", ch)
		}
	}
}
