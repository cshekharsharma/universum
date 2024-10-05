package memtable

import (
	"reflect"
	"testing"
	"universum/entity"
)

func TestListMapMemTable_Exists(t *testing.T) {
	lm := &ListMapMemTable{}
	exists, version := lm.Exists("testKey")
	if exists != false || version != 0 {
		t.Errorf("Expected (false, 0), got (%v, %v)", exists, version)
	}
}

func TestListMapMemTable_Get(t *testing.T) {
	lm := &ListMapMemTable{}
	record, version := lm.Get("testKey")
	expectedRecord := &entity.ScalarRecord{}
	if !reflect.DeepEqual(record, expectedRecord) || version != 0 {
		t.Errorf("Expected (%v, 0), got (%v, %v)", expectedRecord, record, version)
	}
}

func TestListMapMemTable_Set(t *testing.T) {
	lm := &ListMapMemTable{}
	success, version := lm.Set("testKey", "testValue", 0)
	if success != false || version != 0 {
		t.Errorf("Expected (false, 0), got (%v, %v)", success, version)
	}
}

func TestListMapMemTable_Delete(t *testing.T) {
	lm := &ListMapMemTable{}
	success, version := lm.Delete("testKey")
	if success != false || version != 0 {
		t.Errorf("Expected (false, 0), got (%v, %v)", success, version)
	}
}

func TestListMapMemTable_IncrDecrInteger(t *testing.T) {
	lm := &ListMapMemTable{}
	result, version := lm.IncrDecrInteger("testKey", 1, true)
	if result != 0 || version != 0 {
		t.Errorf("Expected (0, 0), got (%v, %v)", result, version)
	}
}

func TestListMapMemTable_Append(t *testing.T) {
	lm := &ListMapMemTable{}
	length, version := lm.Append("testKey", "value")
	if length != 0 || version != 0 {
		t.Errorf("Expected (0, 0), got (%v, %v)", length, version)
	}
}

func TestListMapMemTable_MGet(t *testing.T) {
	lm := &ListMapMemTable{}
	resultMap, version := lm.MGet([]string{"key1", "key2"})
	if len(resultMap) != 0 || version != 0 {
		t.Errorf("Expected (empty map, 0), got (%v, %v)", resultMap, version)
	}
}

func TestListMapMemTable_MSet(t *testing.T) {
	lm := &ListMapMemTable{}
	resultMap, version := lm.MSet(map[string]interface{}{"key1": "value1"})
	if len(resultMap) != 0 || version != 0 {
		t.Errorf("Expected (empty map, 0), got (%v, %v)", resultMap, version)
	}
}

func TestListMapMemTable_MDelete(t *testing.T) {
	lm := &ListMapMemTable{}
	resultMap, version := lm.MDelete([]string{"key1", "key2"})
	if len(resultMap) != 0 || version != 0 {
		t.Errorf("Expected (empty map, 0), got (%v, %v)", resultMap, version)
	}
}

func TestListMapMemTable_TTL(t *testing.T) {
	lm := &ListMapMemTable{}
	ttl, version := lm.TTL("testKey")
	if ttl != 0 || version != 0 {
		t.Errorf("Expected (0, 0), got (%v, %v)", ttl, version)
	}
}

func TestListMapMemTable_Expire(t *testing.T) {
	lm := &ListMapMemTable{}
	success, version := lm.Expire("testKey", 60)
	if success != false || version != 0 {
		t.Errorf("Expected (false, 0), got (%v, %v)", success, version)
	}
}

func TestListMapMemTable_GetSize(t *testing.T) {
	lm := &ListMapMemTable{}
	size := lm.GetSize()
	if size != 0 {
		t.Errorf("Expected size 0, got %v", size)
	}
}

func TestListMapMemTable_IsFull(t *testing.T) {
	lm := &ListMapMemTable{}
	isFull := lm.IsFull()
	if isFull != false {
		t.Errorf("Expected IsFull false, got %v", isFull)
	}
}

func TestListMapMemTable_GetRecordCount(t *testing.T) {
	lm := &ListMapMemTable{}
	count := lm.GetRecordCount()
	if count != 0 {
		t.Errorf("Expected record count 0, got %v", count)
	}
}

func TestListMapMemTable_GetAllRecords(t *testing.T) {
	lm := &ListMapMemTable{}
	recordsList := lm.GetAllRecords()

	if recordsList != nil {
		t.Errorf("Expected nil, got %v", recordsList)
	}
}
