package dslib

import (
	"testing"
)

func TestSkipList(t *testing.T) {
	sl := NewSkipList()

	sl.Insert("a", "Value a", 10)
	sl.Insert("b", "Value b", 11)
	sl.Insert("c", "Value c", 12)
	sl.Insert("d", "Value d", 13)

	tests := []struct {
		key      string
		expected string
		expiry   int64
		found    bool
	}{
		{"a", "Value a", 10, true},
		{"b", "Value b", 11, true},
		{"c", "Value c", 12, true},
		{"d", "Value d", 13, true},
		{"e", "", 0, false}, // Not found case
	}

	for _, tt := range tests {
		t.Run("Search", func(t *testing.T) {
			found, value, expiry := sl.Search(tt.key)
			if found != tt.found {
				t.Errorf("Search(%s) = found %v, want %v", tt.key, found, tt.found)
			}
			if found && (value != tt.expected || expiry != tt.expiry) {
				t.Errorf("Search(%s) = %v [%d], want %v [%d]", tt.key, value, expiry, tt.expected, tt.expiry)
			}
		})
	}

	for _, tt := range tests {
		t.Run("Get", func(t *testing.T) {
			found, value, expiry := sl.Get(tt.key)
			if found != tt.found {
				t.Errorf("Get(%s) = found %v, want %v", tt.key, found, tt.found)
			}
			if found && (value != tt.expected || expiry != tt.expiry) {
				t.Errorf("Get(%s) = %v [%d], want %v [%d]", tt.key, value, expiry, tt.expected, tt.expiry)
			}
		})
	}

	t.Run("UpdateExistingKey", func(t *testing.T) {
		sl.Insert("b", "Updated Value b", 20)
		found, value, expiry := sl.Search("b")
		if !found {
			t.Errorf("Expected to find key b after updating, but it wasn't found")
		}
		if value != "Updated Value b" {
			t.Errorf("Expected updated value 'Updated Value b', but got %v", value)
		}

		if expiry != 20 {
			t.Errorf("Expected updated expiry '20', but got %v", expiry)
		}
	})

	t.Run("DeleteExistingKey", func(t *testing.T) {
		deleted := sl.Remove("b")
		if !deleted {
			t.Errorf("Expected Delete(b) to return true, got false")
		}
		found, _, _ := sl.Search("b")
		if found {
			t.Errorf("Expected key b to be deleted, but it still exists")
		}
	})

	t.Run("DeleteNonExistingKey", func(t *testing.T) {
		deleted := sl.Remove("e")
		if deleted {
			t.Errorf("Expected Delete(50) to return false, but got true")
		}
	})

	t.Run("SkipListSize", func(t *testing.T) {
		expectedSize := 3
		if sl.Size() != expectedSize {
			t.Errorf("Expected size %d, got %d", expectedSize, sl.Size())
		}
	})

	t.Run("InsertAndSearchNewElement", func(t *testing.T) {
		sl.Insert("bc", "Value bc", 40)
		found, value, expiry := sl.Search("bc")
		if !found {
			t.Errorf("Expected to find key bc, but it wasn't found")
		}
		if value != "Value bc" {
			t.Errorf("Expected value 'Value bc', but got %v", value)
		}

		if expiry != 40 {
			t.Errorf("Expected updated expiry '40', but got %v", expiry)
		}
	})

	t.Run("GetAllRecords", func(t *testing.T) {
		recordList := sl.GetAllRecords()

		if len(recordList) != 4 {
			t.Errorf("Expected 4 keys, got %d", len(recordList))
		}

		for i := 0; i < 4; i++ {
			if recordList[i] == nil {
				t.Errorf("expected record to not be nil, but nil found at index %d", i)
			}
		}
	})
}
