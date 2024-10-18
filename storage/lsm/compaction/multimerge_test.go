package compaction

import (
	"testing"
	"universum/entity"
)

func createRecordKV(key string, value int) *entity.RecordKV {
	return &entity.RecordKV{
		Key:    key,
		Record: &entity.ScalarRecord{Value: value},
	}
}

func compareResults(t *testing.T, result []*entity.RecordKV, expectedKeys []string) {
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys, got %d", len(expectedKeys), len(result))
	}
	for i, recordKV := range result {
		if recordKV.Key != expectedKeys[i] {
			t.Errorf("expected key %s, got %s", expectedKeys[i], recordKV.Key)
		}
	}
}

func TestMultiWayMerge_Basic(t *testing.T) {
	arr1 := []*entity.RecordKV{
		createRecordKV("apple", 1),
		createRecordKV("banana", 2),
		createRecordKV("cherry", 3),
	}

	arr2 := []*entity.RecordKV{
		createRecordKV("apricot", 4),
		createRecordKV("blueberry", 5),
		createRecordKV("date", 6),
	}

	result := MultiWayMerge([][]*entity.RecordKV{arr1, arr2})
	expectedKeys := []string{"apple", "apricot", "banana", "blueberry", "cherry", "date"}
	compareResults(t, result, expectedKeys)
}

func TestMultiWayMerge_DuplicateKeys(t *testing.T) {
	arr1 := []*entity.RecordKV{
		createRecordKV("apple", 1),
		createRecordKV("banana", 2),
	}

	arr2 := []*entity.RecordKV{
		createRecordKV("banana", 3),
		createRecordKV("date", 6),
	}

	result := MultiWayMerge([][]*entity.RecordKV{arr1, arr2})
	expectedKeys := []string{"apple", "banana", "banana", "date"}
	compareResults(t, result, expectedKeys)
}

func TestMultiWayMerge_EmptyInput(t *testing.T) {
	arr1 := []*entity.RecordKV{}

	arr2 := []*entity.RecordKV{
		createRecordKV("banana", 3),
	}

	result := MultiWayMerge([][]*entity.RecordKV{arr1, arr2})
	expectedKeys := []string{"banana"}
	compareResults(t, result, expectedKeys)
}

func TestMultiWayMerge_MultipleArrays(t *testing.T) {
	arr1 := []*entity.RecordKV{
		createRecordKV("apple", 1),
	}

	arr2 := []*entity.RecordKV{
		createRecordKV("banana", 2),
		createRecordKV("date", 6),
	}

	arr3 := []*entity.RecordKV{
		createRecordKV("cherry", 5),
	}

	result := MultiWayMerge([][]*entity.RecordKV{arr1, arr2, arr3})
	expectedKeys := []string{"apple", "banana", "cherry", "date"}
	compareResults(t, result, expectedKeys)
}

func TestMultiWayMerge_SingleArray(t *testing.T) {
	arr1 := []*entity.RecordKV{
		createRecordKV("apple", 1),
		createRecordKV("banana", 2),
	}

	result := MultiWayMerge([][]*entity.RecordKV{arr1})
	expectedKeys := []string{"apple", "banana"}
	compareResults(t, result, expectedKeys)
}
