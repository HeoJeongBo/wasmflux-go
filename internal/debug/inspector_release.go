//go:build !debug

package debug

// Inspector is a no-op in release builds.
type Inspector struct{}

func NewInspector() *Inspector { return &Inspector{} }
func (ins *Inspector) Release() {}
