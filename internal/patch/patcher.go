package patch

// PatchOptions controls patching behavior.
type PatchOptions struct {
	// DryRun reports what would be changed without modifying the ROM.
	DryRun bool
	// Backup creates a .bak copy before patching.
	Backup bool
	// OutputPath writes the patched ROM to a different file (empty = in-place).
	OutputPath string
}

// Patcher applies a patch to a ROM file.
type Patcher interface {
	Apply(romPath string, opts PatchOptions) error
}
