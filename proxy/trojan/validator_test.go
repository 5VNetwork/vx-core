package trojan

import (
	"crypto/rand"
	"sync"
	"testing"

	"github.com/5vnetwork/vx-core/common/uuid"
)

func createTestAccount() *MemoryAccount {
	uid := uuid.New().String()
	key := make([]byte, 32)
	rand.Read(key)
	password := make([]byte, 16)
	rand.Read(password)

	return &MemoryAccount{
		Uid:      uid,
		Key:      key,
		Password: password,
	}
}

func TestValidator_Add(t *testing.T) {
	validator := &Validator{}
	account := createTestAccount()

	validator.Add(account)

	retrievedByHash := validator.Get(hexString(account.Key))
	if retrievedByHash == nil {
		t.Fatal("Account not found by hash after adding")
	}
	if retrievedByHash.Uid != account.Uid {
		t.Errorf("Expected UUID %v, got %v", account.Uid, retrievedByHash.Uid)
	}

	retrievedByUUID, exists := validator.uuidToAccount.Load(account.Uid)
	if !exists {
		t.Fatal("Account not found by UUID after adding")
	}
	if retrievedByUUID.(*MemoryAccount).Uid != account.Uid {
		t.Errorf("Expected UUID %v, got %v", account.Uid, retrievedByUUID.(*MemoryAccount).Uid)
	}
}

func TestValidator_Del_ValidUUID(t *testing.T) {
	validator := &Validator{}
	account := createTestAccount()

	validator.Add(account)

	err := validator.Del(account.Uid)
	if err != nil {
		t.Fatalf("Expected no error when deleting valid UUID, got: %v", err)
	}

	retrievedByHash := validator.Get(hexString(account.Key))
	if retrievedByHash != nil {
		t.Error("Account should not be found by hash after deletion")
	}

	_, exists := validator.uuidToAccount.Load(account.Uid)
	if exists {
		t.Error("Account should not be found by UUID after deletion")
	}
}

func TestValidator_Del_InvalidUUID(t *testing.T) {
	validator := &Validator{}
	nonExistentUUID := uuid.New().String()

	err := validator.Del(nonExistentUUID)
	if err == nil {
		t.Error("Expected error when deleting non-existent UUID")
	}
	if err.Error() != " not found." {
		t.Errorf("Expected error message ' not found.', got: %v", err.Error())
	}
}

func TestValidator_Get_ValidHash(t *testing.T) {
	validator := &Validator{}
	account := createTestAccount()

	validator.Add(account)

	retrieved := validator.Get(hexString(account.Key))
	if retrieved == nil {
		t.Fatal("Expected account to be found by valid hash")
	}
	if retrieved.Uid != account.Uid {
		t.Errorf("Expected UUID %v, got %v", account.Uid, retrieved.Uid)
	}
	if string(retrieved.Key) != string(account.Key) {
		t.Error("Expected retrieved key to match original key")
	}
}

func TestValidator_Get_InvalidHash(t *testing.T) {
	validator := &Validator{}

	retrieved := validator.Get("nonexistenthash")
	if retrieved != nil {
		t.Error("Expected nil for non-existent hash")
	}

	retrieved = validator.Get("")
	if retrieved != nil {
		t.Error("Expected nil for empty hash")
	}
}

func TestValidator_ConcurrentOperations(t *testing.T) {
	validator := &Validator{}
	const numGoroutines = 100

	accounts := make([]*MemoryAccount, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		accounts[i] = createTestAccount()
	}

	var wg sync.WaitGroup

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(account *MemoryAccount) {
			defer wg.Done()
			validator.Add(account)
		}(accounts[i])
	}
	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		retrieved := validator.Get(hexString(accounts[i].Key))
		if retrieved == nil {
			t.Errorf("Account %d not found after concurrent add", i)
		} else if retrieved.Uid != accounts[i].Uid {
			t.Errorf("Account %d UUID mismatch", i)
		}
	}

	wg.Add(numGoroutines / 2)
	for i := 0; i < numGoroutines/2; i++ {
		go func(account *MemoryAccount) {
			defer wg.Done()
			validator.Del(account.Uid)
		}(accounts[i])
	}
	wg.Wait()

	for i := 0; i < numGoroutines/2; i++ {
		retrieved := validator.Get(hexString(accounts[i].Key))
		if retrieved != nil {
			t.Errorf("Account %d should be deleted after concurrent delete", i)
		}
	}

	for i := numGoroutines / 2; i < numGoroutines; i++ {
		retrieved := validator.Get(hexString(accounts[i].Key))
		if retrieved == nil {
			t.Errorf("Account %d should still exist after concurrent operations", i)
		}
	}
}
