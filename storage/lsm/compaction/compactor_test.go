package compaction

import (
	"fmt"
	"reflect"
	"testing"
	"time"
	"universum/config"
	"universum/entity"
	"universum/storage/lsm/sstable"
)

func setupConfig(t *testing.T) {
	config.Store = config.GetSkeleton()
	config.Store.Storage.LSM.DataStorageDirectory = t.TempDir()
	config.Store.Storage.LSM.BloomFilterMaxRecords = 100
	config.Store.Storage.LSM.BloomFalsePositiveRate = 0.01
	config.Store.Storage.LSM.WriteBufferSize = 1024
	config.Store.Storage.LSM.WriteBlockSize = 1024
}

func createDummySSTable(id int64, records []*entity.RecordKV) *sstable.SSTable {
	sst, _ := sstable.NewSSTable(
		fmt.Sprintf("%d.%s", id, sstable.SstFileExtension),
		sstable.SSTmodeWrite,
		config.Store.Storage.LSM.BloomFilterMaxRecords,
		config.Store.Storage.LSM.BloomFalsePositiveRate,
	)

	sst.FlushRecordsToSSTable(records)
	return sst
}

func TestAddSSTable(t *testing.T) {
	setupConfig(t)
	compactor := NewCompactor()

	records := []*entity.RecordKV{
		{Key: "key1", Record: &entity.ScalarRecord{Value: "value1"}},
		{Key: "key2", Record: &entity.ScalarRecord{Value: "value2"}},
	}
	sst := createDummySSTable(1, records)
	compactor.AddSSTable(0, sst)

	if len(compactor.LevelSSTables[0]) != 1 {
		t.Errorf("Expected 1 SSTable in level 0, got %d", len(compactor.LevelSSTables[0]))
	}
}

func TestCompactLevel(t *testing.T) {
	setupConfig(t)
	compactor := NewCompactor()

	records1 := []*entity.RecordKV{
		{Key: "key1", Record: &entity.ScalarRecord{Value: "value1"}},
		{Key: "key2", Record: &entity.ScalarRecord{Value: "value2"}},
	}
	records2 := []*entity.RecordKV{
		{Key: "key3", Record: &entity.ScalarRecord{Value: "value3"}},
		{Key: "key4", Record: &entity.ScalarRecord{Value: "value4"}},
	}
	records3 := []*entity.RecordKV{
		{Key: "key5", Record: &entity.ScalarRecord{Value: "value5"}},
		{Key: "key6", Record: &entity.ScalarRecord{Value: "value6"}},
	}
	sst1 := createDummySSTable(1, records1)
	sst2 := createDummySSTable(2, records2)
	sst3 := createDummySSTable(3, records3)

	compactor.AddSSTable(0, sst1)
	compactor.AddSSTable(0, sst2)
	compactor.AddSSTable(0, sst3)

	SSTReplacementChan = make(chan *SSTReplacement, 1)
	go func() {
		for range SSTReplacementChan {
			time.Sleep(10 * time.Microsecond)
		}
	}()

	go compactor.Compact()
	time.Sleep(1 * time.Second)

	if len(compactor.LevelSSTables[0]) != 0 {
		t.Errorf("Expected 0 SSTables in level 0 after compaction, got %d", len(compactor.LevelSSTables[0]))
	}

	if len(compactor.LevelSSTables[1]) != 1 {
		t.Errorf("Expected 1 SSTable in level 1, got %d", len(compactor.LevelSSTables[1]))
	}
}

func TestMergeSSTables(t *testing.T) {
	setupConfig(t)
	compactor := NewCompactor()

	futureTime := time.Now().Unix() + 1000
	pastTime := time.Now().Unix() - 1000

	records1 := []*entity.RecordKV{
		{Key: "key1", Record: &entity.ScalarRecord{Value: "value1", Expiry: pastTime, State: entity.RecordStateActive}},
		{Key: "key2", Record: &entity.ScalarRecord{Value: "value2", Expiry: futureTime, State: entity.RecordStateActive}},
		{Key: "key3", Record: &entity.ScalarRecord{Value: "value3", Expiry: futureTime, State: entity.RecordStateActive}},
		{Key: "key4", Record: &entity.ScalarRecord{Value: "value4", Expiry: futureTime, State: entity.RecordStateTombstoned}},
		{Key: "key5", Record: &entity.ScalarRecord{Value: "value5", Expiry: futureTime, State: entity.RecordStateActive}},
		{Key: "key8", Record: &entity.ScalarRecord{Value: "value8", Expiry: pastTime, State: entity.RecordStateActive}},
	}
	records2 := []*entity.RecordKV{
		{Key: "key2", Record: &entity.ScalarRecord{Value: "value2", Expiry: pastTime, State: entity.RecordStateActive}},
		{Key: "key5", Record: &entity.ScalarRecord{Value: "value5+1", Expiry: futureTime, State: entity.RecordStateActive}},
		{Key: "key6", Record: &entity.ScalarRecord{Value: "value6", Expiry: futureTime, State: entity.RecordStateActive}},
		{Key: "key7", Record: &entity.ScalarRecord{Value: "value7", Expiry: futureTime, State: entity.RecordStateActive}},
		{Key: "key8", Record: &entity.ScalarRecord{Value: "value8+1", Expiry: futureTime, State: entity.RecordStateActive}},
	}

	sst1 := createDummySSTable(1, records1)
	sst2 := createDummySSTable(2, records2)

	mergedSST, err := compactor.mergeSSTables([]*sstable.SSTable{sst1, sst2})
	if err != nil {
		t.Errorf("Merge failed: %v", err)
	}

	mergedRecords, err := mergedSST.GetAllRecords()
	if err != nil {
		t.Errorf("Failed to get merged records: %v", err)
	}

	expectedMergedList := []*entity.RecordKV{
		{Key: "key3", Record: &entity.ScalarRecord{Value: "value3", Expiry: futureTime, State: entity.RecordStateActive}},
		{Key: "key5", Record: &entity.ScalarRecord{Value: "value5+1", Expiry: futureTime, State: entity.RecordStateActive}},
		{Key: "key6", Record: &entity.ScalarRecord{Value: "value6", Expiry: futureTime, State: entity.RecordStateActive}},
		{Key: "key7", Record: &entity.ScalarRecord{Value: "value7", Expiry: futureTime, State: entity.RecordStateActive}},
		{Key: "key8", Record: &entity.ScalarRecord{Value: "value8+1", Expiry: futureTime, State: entity.RecordStateActive}},
	}

	if !reflect.DeepEqual(mergedRecords, expectedMergedList) {
		t.Errorf("Expected %v, got %v", expectedMergedList, mergedRecords)
	}
}

func TestDeleteOldSSTables(t *testing.T) {
	setupConfig(t)
	compactor := NewCompactor()

	records := []*entity.RecordKV{
		{Key: "key1", Record: &entity.ScalarRecord{Value: "value1"}},
	}
	sst1 := createDummySSTable(1, records)
	sst2 := createDummySSTable(2, records)

	compactor.AddSSTable(0, sst1)
	compactor.AddSSTable(0, sst2)

	err := compactor.deleteOldSSTables([]*sstable.SSTable{sst1, sst2})
	if err != nil {
		t.Errorf("Failed to delete old SSTables: %v", err)
	}

	_, err = sst1.GetAllRecords()
	if err == nil {
		t.Errorf("Expected SSTable %s to be deleted", sst1.Filename)
	}

	_, err = sst2.GetAllRecords()
	if err == nil {
		t.Errorf("Expected SSTable %s to be deleted", sst2.Filename)
	}
}

func TestGetMergedSSTFileName(t *testing.T) {
	setupConfig(t)
	compactor := NewCompactor()

	sst1 := createDummySSTable(123456, nil)
	sst2 := createDummySSTable(654321, nil)

	mergedFileName, err := compactor.getMergedSSTFileName([]*sstable.SSTable{sst1, sst2})
	if err != nil {
		t.Errorf("Failed to generate merged SST file name: %v", err)
	}

	expectedFileName := fmt.Sprintf("%d.%s", 123456, sstable.SstFileExtension)
	if mergedFileName != expectedFileName {
		t.Errorf("Expected %s, got %s", expectedFileName, mergedFileName)
	}
}
