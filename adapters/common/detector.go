package common

// Detector checks tool presence and Sequoia installation state for a tool
// adapter. It is decoupled from PathResolver — it receives a baseFn function
// instead of holding a *PathResolver reference. This makes Detector testable
// with injected functions and respects the spec's "zero install dependencies"
// requirement.
//
// All fields are private. Consumers construct via NewDetector and interact
// through the Detect() and IsInstalled() methods.
type Detector struct {
	baseFn        func() (string, error)
	isInstalledFn func(base string) bool
	detectFn      func() bool
}

// NewDetector creates a Detector with the given detection functions.
//
// Parameters:
//   - baseFn: resolves the tool's config root directory. When it returns an
//     error, IsInstalled() returns false without calling isInstalledFn.
//   - isInstalledFn: checks whether Sequoia is installed, given the base dir.
//     If nil, IsInstalled() returns false safely.
//   - detectFn: reports whether the tool itself is present on the host.
//     If nil, Detect() returns false safely.
func NewDetector(
	baseFn func() (string, error),
	isInstalledFn func(base string) bool,
	detectFn func() bool,
) *Detector {
	return &Detector{
		baseFn:        baseFn,
		isInstalledFn: isInstalledFn,
		detectFn:      detectFn,
	}
}

// Detect reports whether the tool is present on this machine.
// Returns false if detectFn is nil (safe default).
func (d *Detector) Detect() bool {
	if d.detectFn == nil {
		return false
	}
	return d.detectFn()
}

// IsInstalled reports whether Sequoia has been installed for this tool.
// Calls baseFn() to resolve the config directory, then delegates to
// isInstalledFn. Returns false if baseFn errors or isInstalledFn is nil.
func (d *Detector) IsInstalled() bool {
	if d.isInstalledFn == nil {
		return false
	}
	base, err := d.baseFn()
	if err != nil {
		return false
	}
	return d.isInstalledFn(base)
}
