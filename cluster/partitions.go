package cluster

import "universum/crypto"

const (
	NumPartitions uint64 = 1 << 12     // 4096
	digestSeed    uint64 = 0x10000001F // 4294967297
)

func GetPartitonID(key string) (uint64, uint64) {
	digest := crypto.MurmurHash64([]byte(key), digestSeed)
	return digest % NumPartitions, digest
}

func GetNodeForPartition(partitionID int64) string {
	return "node1"
}
