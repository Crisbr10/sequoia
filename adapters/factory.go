package adapters

// NewAdapter returns the adapter registered under id in DefaultRegistry.
// It returns ErrUnknownAdapter (wrapped) if no adapter with that ID exists.
// This is a convenience wrapper around DefaultRegistry.Get.
func NewAdapter(id string) (ToolAdapter, error) {
	return DefaultRegistry.Get(id)
}
