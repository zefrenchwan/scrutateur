package storage_test

import (
	"context"
	"testing"

	"github.com/zefrenchwan/scrutateur.git/storage"
)

const REDIS_TEST_URL = "redis://127.0.0.1:6379/1"

func TestCacheHas(t *testing.T) {
	cache, errBoot := storage.NewCacheStorage(REDIS_TEST_URL)
	if errBoot != nil {
		t.Log("failed to create cache")
		t.FailNow()
	}

	defer cache.Close()
	if found, err := cache.Has(context.Background(), "test"); err != nil {
		t.Log("failed to test value", err)
		t.Fail()
	} else if found {
		t.Log("failed to test has when no value")
		t.Fail()
	} else if err := cache.SetValue(context.Background(), "test", []byte("found")); err != nil {
		t.Log("Set value error", err)
		t.Fail()
	} else if found, err := cache.Has(context.Background(), "test"); err != nil {
		t.Log("failed to test has", err)
		t.Fail()
	} else if !found {
		t.Log("failed to test has when value is in cache")
		t.Fail()
	} else if val, err := cache.GetValue(context.Background(), "test"); err != nil {
		t.Log("Fail to get value", err)
		t.Fail()
	} else if string(val) != "found" {
		t.Log("Mismatch value", err)
		t.Fail()
	} else if err := cache.Delete(context.Background(), "test"); err != nil {
		t.Log("Failed to delete", err)
		t.Fail()
	} else if found, err := cache.Has(context.Background(), "test"); err != nil {
		t.Log("failed to test has", err)
		t.Fail()
	} else if found {
		t.Log("Failed to test has when value is in")
		t.Fail()
	}
}
