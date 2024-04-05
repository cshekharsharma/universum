# TODOs

- Use sync.Map with memory slabs/shards and allocate items using hashing alogs
- Keep expiry in record itself, and one map is enough instead of 3
- Implement memory size capping based on config
- LRU based eviction if memory is full
- Add MultiGet, MultiSet, MultiDelete, and Stats command
- Implement db stats holder
- Implement AOF loader for persistent
- Implement persistent storage with B+ or LSM tree
