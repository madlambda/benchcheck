// Package fake defines a fake function to be used
// on fake benchmarks that are used for testing purposes.
// We change the performance of the fake function on version
// control so use it as a way to check benchcheck behavior.
package fake

import "time"

func Do() {
	time.Sleep(500 * time.Millisecond)
}
