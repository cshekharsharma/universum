package compaction

import "universum/entity"

func Merge(arr1, arr2 []*entity.RecordKV) []*entity.RecordKV {
	result := make([]*entity.RecordKV, 0, len(arr1)+len(arr2))
	i, j := 0, 0

	for i < len(arr1) && j < len(arr2) {
		if arr1[i].Key < arr2[j].Key {
			result = append(result, arr1[i])
			i++
		} else if arr1[i].Key > arr2[j].Key {
			result = append(result, arr2[j])
			j++
		} else {
			result = append(result, arr2[j])
			i++
			j++
		}
	}

	for i < len(arr1) {
		result = append(result, arr1[i])
		i++
	}
	for j < len(arr2) {
		result = append(result, arr2[j])
		j++
	}

	return result
}
