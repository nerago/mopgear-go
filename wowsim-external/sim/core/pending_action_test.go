package core

// import (
// 	"testing"
// 	"time"
// )

// func BenchmarkArrayAddIncreasing(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueArray{}
// 		queue.reset()
// 		b.StartTimer()

// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 		}
// 	}
// }

// func BenchmarkArrayAddDecreasing(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueArray{}
// 		queue.reset()
// 		b.StartTimer()

// 		for when := 20; when >= 1; when-- {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 		}
// 	}
// }

// func BenchmarkSingleAddIncreasing(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueSingle{}
// 		queue.reset()
// 		b.StartTimer()

// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 		}
// 	}
// }

// func BenchmarkSingleAddDecreasing(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueSingle{}
// 		queue.reset()
// 		b.StartTimer()

// 		for when := 20; when >= 1; when-- {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 		}
// 	}
// }

// func BenchmarkDoubleAddIncreasing(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueDouble{}
// 		queue.reset()
// 		b.StartTimer()

// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 		}
// 	}
// }

// func BenchmarkDoubleAddDecreasing(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueDouble{}
// 		queue.reset()
// 		b.StartTimer()

// 		for when := 20; when >= 1; when-- {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 		}
// 	}
// }

// func BenchmarkArrayCancelVarious(b *testing.B) {
// 	cancelIndex := 0

// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueArray{}
// 		queue.reset()
// 		items := []*PendingAction{}
// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 			items = append(items, add)
// 		}
// 		b.StartTimer()

// 		queue.cancel(items[cancelIndex])
// 		cancelIndex = (cancelIndex + 1) % 20
// 	}
// }

// func BenchmarkSingleCancelVarious(b *testing.B) {
// 	cancelIndex := 0

// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueSingle{}
// 		queue.reset()
// 		items := []*PendingAction{}
// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 			items = append(items, add)
// 		}
// 		b.StartTimer()

// 		queue.cancel(items[cancelIndex])
// 		cancelIndex = (cancelIndex + 1) % 20
// 	}
// }

// func BenchmarkDoubleCancelVarious(b *testing.B) {
// 	cancelIndex := 0

// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueDouble{}
// 		queue.reset()
// 		items := []*PendingAction{}
// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 			items = append(items, add)
// 		}
// 		b.StartTimer()

// 		queue.cancel(items[cancelIndex])
// 		cancelIndex = (cancelIndex + 1) % 20
// 	}
// }

// func BenchmarkArrayCancelFirst(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueArray{}
// 		queue.reset()
// 		items := []*PendingAction{}
// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 			items = append(items, add)
// 		}
// 		b.StartTimer()

// 		queue.cancel(items[0])
// 	}
// }

// func BenchmarkSingleCancelFirst(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueSingle{}
// 		queue.reset()
// 		items := []*PendingAction{}
// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 			items = append(items, add)
// 		}
// 		b.StartTimer()

// 		queue.cancel(items[0])
// 	}
// }

// func BenchmarkDoubleCancelFirst(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueDouble{}
// 		queue.reset()
// 		items := []*PendingAction{}
// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 			items = append(items, add)
// 		}
// 		b.StartTimer()

// 		queue.cancel(items[0])
// 	}
// }

// func BenchmarkArrayCancelLast(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueArray{}
// 		queue.reset()
// 		items := []*PendingAction{}
// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 			items = append(items, add)
// 		}
// 		b.StartTimer()

// 		queue.cancel(items[len(items)-1])
// 	}
// }

// func BenchmarkSingleCancelLast(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueSingle{}
// 		queue.reset()
// 		items := []*PendingAction{}
// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 			items = append(items, add)
// 		}
// 		b.StartTimer()

// 		queue.cancel(items[len(items)-1])
// 	}
// }

// func BenchmarkDoubleCancelLast(b *testing.B) {
// 	for b.Loop() {
// 		b.StopTimer()
// 		queue := PendingActionQueueDouble{}
// 		queue.reset()
// 		items := []*PendingAction{}
// 		for when := 1; when <= 20; when++ {
// 			add := &PendingAction{
// 				NextActionAt: time.Duration(when),
// 				OnAction:     func(sim *Simulation) {},
// 			}
// 			queue.add(add)
// 			items = append(items, add)
// 		}
// 		b.StartTimer()

// 		queue.cancel(items[len(items)-1])
// 	}
// }
