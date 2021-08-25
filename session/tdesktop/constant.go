package tdesktop

//nolint:revive
const (
	localEncryptIterCount      = 4000 // key derivation iteration count
	localEncryptNoPwdIterCount = 4    // key derivation iteration count without pwd (not secure anyway)
	localEncryptSaltSize       = 32   // 256 bit

	kStrongIterationsCount = 100000

	kWideIdsTag = ^uint64(0)
)
