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

func compareResults(t *testing.T, result []*entity.RecordKV, expectedKeys []string, expectedValues []int) {
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys, got %d", len(expectedKeys), len(result))
	}
	for i, recordKV := range result {
		if recordKV.Key != expectedKeys[i] {
			t.Errorf("expected key %s, got %s", expectedKeys[i], recordKV.Key)
		}
		if recordKV.Record.(*entity.ScalarRecord).Value != expectedValues[i] {
			t.Errorf("expected value %d for key %s, got %d", expectedValues[i], expectedKeys[i], recordKV.Record.(*entity.ScalarRecord).Value)
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

	result := Merge(arr1, arr2)
	expectedKeys := []string{"apple", "apricot", "banana", "blueberry", "cherry", "date"}
	expectedValues := []int{1, 4, 2, 5, 3, 6}
	compareResults(t, result, expectedKeys, expectedValues)
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

	result := Merge(arr1, arr2)
	expectedKeys := []string{"apple", "banana", "date"}
	expectedValues := []int{1, 3, 6}
	compareResults(t, result, expectedKeys, expectedValues)
}

func TestMultiWayMerge_EmptyInput(t *testing.T) {
	arr1 := []*entity.RecordKV{}

	arr2 := []*entity.RecordKV{
		createRecordKV("banana", 3),
	}

	result := Merge(arr1, arr2)
	expectedKeys := []string{"banana"}
	expectedValues := []int{3}
	compareResults(t, result, expectedKeys, expectedValues)
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

	result := Merge(arr1, arr2)
	result = Merge(result, arr3) // Chaining two merges

	expectedKeys := []string{"apple", "banana", "cherry", "date"}
	expectedValues := []int{1, 2, 5, 6}
	compareResults(t, result, expectedKeys, expectedValues)
}

func TestMultiWayMerge_SingleArray(t *testing.T) {
	arr1 := []*entity.RecordKV{
		createRecordKV("apple", 1),
		createRecordKV("banana", 2),
	}

	result := Merge(arr1, []*entity.RecordKV{})
	expectedKeys := []string{"apple", "banana"}
	expectedValues := []int{1, 2}
	compareResults(t, result, expectedKeys, expectedValues)
}
