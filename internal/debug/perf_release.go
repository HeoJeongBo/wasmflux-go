//go:build !debug

package debug

// No-op implementations for release builds.
// These are compiled away by the linker.

func Mark(string)        {}
func MeasureStart(string) {}
func MeasureEnd(string) float64 { return 0 }
func ClearMarks()        {}
func ClearMeasures()     {}
