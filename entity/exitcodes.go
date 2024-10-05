package entity

const (
	ExitCodeSuccess        int = 0 << 1 // 0
	ExitCodeStartupFailure int = 1 << 0 // 1
	ExitCodeSocketError    int = 1 << 1 // 2
	ExitCodeInturrupted    int = 1 << 2 // 4
	ExitcodeUnknown        int = 1 << 3 // 8
)
