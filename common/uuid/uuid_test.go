package uuid

import (
	"testing"

	"github.com/5vnetwork/vx-core/common"

	"github.com/google/go-cmp/cmp"
)

func TestParseBytes(t *testing.T) {
	str := "2418d087-648d-4990-86e8-19dca1d006d3"
	bytes := []byte{0x24, 0x18, 0xd0, 0x87, 0x64, 0x8d, 0x49, 0x90, 0x86, 0xe8, 0x19, 0xdc, 0xa1, 0xd0, 0x06, 0xd3}

	uuid, err := ParseBytes(bytes)
	common.Must(err)
	if diff := cmp.Diff(uuid.String(), str); diff != "" {
		t.Error(diff)
	}

	_, err = ParseBytes([]byte{1, 3, 2, 4})
	if err == nil {
		t.Fatal("Expect error but nil")
	}
}

func TestParseString(t *testing.T) {
	str := "2418d087-648d-4990-86e8-19dca1d006d3"
	expectedBytes := []byte{0x24, 0x18, 0xd0, 0x87, 0x64, 0x8d, 0x49, 0x90, 0x86, 0xe8, 0x19, 0xdc, 0xa1, 0xd0, 0x06, 0xd3}

	uuid, err := ParseString(str)
	common.Must(err)
	if r := cmp.Diff(expectedBytes, uuid.Bytes()); r != "" {
		t.Fatal(r)
	}

	_, err = ParseString("2418d087")
	if err == nil {
		t.Fatal("Expect error but nil")
	}

	_, err = ParseString("2418d087-648k-4990-86e8-19dca1d006d3")
	if err == nil {
		t.Fatal("Expect error but nil")
	}
}

func TestNewUUID(t *testing.T) {
	uuid := New()
	uuid2, err := ParseString(uuid.String())

	common.Must(err)
	if uuid.String() != uuid2.String() {
		t.Error("uuid string: ", uuid.String(), " != ", uuid2.String())
	}
	if r := cmp.Diff(uuid.Bytes(), uuid2.Bytes()); r != "" {
		t.Error(r)
	}
}

func TestRandom(t *testing.T) {
	uuid := New()
	uuid2 := New()

	if uuid.String() == uuid2.String() {
		t.Error("duplicated uuid")
	}
}

func TestEquals(t *testing.T) {
	var uuid *UUID
	var uuid2 *UUID
	if !uuid.Equals(uuid2) {
		t.Error("empty uuid should equal")
	}

	uuid3 := New()
	if uuid.Equals(&uuid3) {
		t.Error("nil uuid equals non-nil uuid")
	}
}

func TestStringToUUID(t *testing.T) {
	t.Run("ValidUUIDString", func(t *testing.T) {
		// Test with valid UUID string (with dashes)
		validUUID := "2418d087-648d-4990-86e8-19dca1d006d3"
		uuid := StringToUUID(validUUID)
		if uuid.String() != validUUID {
			t.Errorf("Expected %s, got %s", validUUID, uuid.String())
		}
	})

	t.Run("ValidUUIDStringNoDashes", func(t *testing.T) {
		// Test with valid UUID string (without dashes)
		validUUID := "2418d087648d499086e819dca1d006d3"
		expectedUUID := "2418d087-648d-4990-86e8-19dca1d006d3"
		uuid := StringToUUID(validUUID)
		if uuid.String() != expectedUUID {
			t.Errorf("Expected %s, got %s", expectedUUID, uuid.String())
		}
	})

	t.Run("InvalidStringGeneratesV5", func(t *testing.T) {
		// Test with non-UUID string (should generate UUID v5)
		testString := "test-user-123"
		uuid := StringToUUID(testString)

		// Verify it's a valid UUID
		if uuid.String() == "" {
			t.Error("Generated UUID should not be empty")
		}

		// Verify version bits (should be version 5 = 0x50)
		versionByte := uuid[6]
		version := (versionByte >> 4) & 0x0f
		if version != 5 {
			t.Errorf("Expected UUID version 5, got version %d", version)
		}

		// Verify variant bits (should be RFC 4122 = 0b10xx)
		variantByte := uuid[8]
		variant := (variantByte >> 6) & 0x03
		if variant != 2 {
			t.Errorf("Expected RFC 4122 variant (2), got %d", variant)
		}
	})

	t.Run("DeterministicGeneration", func(t *testing.T) {
		// Test that same input produces same UUID
		testString := "my-service-name"
		uuid1 := StringToUUID(testString)
		uuid2 := StringToUUID(testString)

		if uuid1.String() != uuid2.String() {
			t.Errorf("Same input should produce same UUID: %s != %s", uuid1.String(), uuid2.String())
		}

		if !uuid1.Equals(&uuid2) {
			t.Error("UUIDs generated from same input should be equal")
		}
	})

	t.Run("DifferentInputsDifferentUUIDs", func(t *testing.T) {
		// Test that different inputs produce different UUIDs
		uuid1 := StringToUUID("string1")
		uuid2 := StringToUUID("string2")

		if uuid1.String() == uuid2.String() {
			t.Error("Different inputs should produce different UUIDs")
		}

		if uuid1.Equals(&uuid2) {
			t.Error("Different inputs should not produce equal UUIDs")
		}
	})

	t.Run("EmptyString", func(t *testing.T) {
		// Test with empty string
		uuid := StringToUUID("")

		// Should generate a valid UUID v5
		if uuid.String() == "" {
			t.Error("Generated UUID should not be empty")
		}

		// Verify deterministic
		uuid2 := StringToUUID("")
		if !uuid.Equals(&uuid2) {
			t.Error("Empty string should always produce same UUID")
		}
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		// Test with special characters
		testStrings := []string{
			"user@example.com",
			"path/to/resource",
			"key:value:pair",
			"name with spaces",
			"unicode-字符",
		}

		for _, str := range testStrings {
			uuid1 := StringToUUID(str)
			// Verify deterministic
			uuid2 := StringToUUID(str)
			if !uuid1.Equals(&uuid2) {
				t.Errorf("String %q should produce deterministic UUID", str)
			}
		}
	})
}

func TestNewV5(t *testing.T) {
	t.Run("BasicGeneration", func(t *testing.T) {
		// Test basic UUID v5 generation
		uuid := NewV5(NamespaceURL, "test")

		// Verify it's not empty
		if uuid == emptyUUID {
			t.Error("Generated UUID should not be empty")
		}

		// Verify version 5
		versionByte := uuid[6]
		version := (versionByte >> 4) & 0x0f
		if version != 5 {
			t.Errorf("Expected UUID version 5, got version %d", version)
		}

		// Verify RFC 4122 variant
		variantByte := uuid[8]
		variant := (variantByte >> 6) & 0x03
		if variant != 2 {
			t.Errorf("Expected RFC 4122 variant (2), got %d", variant)
		}
	})

	t.Run("Deterministic", func(t *testing.T) {
		// Test deterministic generation
		name := "test-name"
		uuid1 := NewV5(NamespaceURL, name)
		uuid2 := NewV5(NamespaceURL, name)

		if uuid1.String() != uuid2.String() {
			t.Errorf("Same input should produce same UUID: %s != %s", uuid1.String(), uuid2.String())
		}
	})

	t.Run("DifferentNamespacesDifferentUUIDs", func(t *testing.T) {
		// Create a different namespace for testing
		namespaceTest := UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}

		name := "test"
		uuid1 := NewV5(NamespaceURL, name)
		uuid2 := NewV5(namespaceTest, name)

		if uuid1.String() == uuid2.String() {
			t.Error("Different namespaces should produce different UUIDs for same name")
		}
	})

	t.Run("DifferentNamesDifferentUUIDs", func(t *testing.T) {
		// Test that different names produce different UUIDs
		uuid1 := NewV5(NamespaceURL, "name1")
		uuid2 := NewV5(NamespaceURL, "name2")

		if uuid1.String() == uuid2.String() {
			t.Error("Different names should produce different UUIDs")
		}
	})

	t.Run("KnownTestVector", func(t *testing.T) {
		// Test with a known UUID v5 test vector
		// This is a well-known test case: generating UUID v5 for "www.example.com" in DNS namespace
		// Expected: 2ed6657d-e927-568b-95e1-2665a8aea6a2
		namespaceDNS := UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
		uuid := NewV5(namespaceDNS, "www.example.com")
		expected := "2ed6657d-e927-568b-95e1-2665a8aea6a2"

		if uuid.String() != expected {
			t.Errorf("Expected known test vector %s, got %s", expected, uuid.String())
		}
	})

	t.Run("EmptyName", func(t *testing.T) {
		// Test with empty name
		uuid := NewV5(NamespaceURL, "")

		// Should still produce valid UUID v5
		versionByte := uuid[6]
		version := (versionByte >> 4) & 0x0f
		if version != 5 {
			t.Errorf("Expected UUID version 5, got version %d", version)
		}

		// Should be deterministic
		uuid2 := NewV5(NamespaceURL, "")
		if uuid.String() != uuid2.String() {
			t.Error("Empty name should produce deterministic UUID")
		}
	})
}
