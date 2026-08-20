package highs

/*
#cgo CFLAGS: -I${SRCDIR}/../internal/highs/include

#cgo linux,amd64 LDFLAGS: ${SRCDIR}/../internal/highs/lib/linux_amd64/libhighs.a -lstdc++ -lm -ldl -lz
#cgo linux,arm64 LDFLAGS: ${SRCDIR}/../internal/highs/lib/linux_arm64/libhighs.a -lstdc++ -lm -ldl -lz
#cgo darwin,amd64 LDFLAGS: ${SRCDIR}/../internal/highs/lib/darwin_amd64/libhighs.a -lc++ -lz
#cgo darwin,arm64 LDFLAGS: ${SRCDIR}/../internal/highs/lib/darwin_arm64/libhighs.a -lc++ -lz
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/../internal/highs/lib/windows_amd64_gpu/ -lhighs -lstdc++

#include <stdlib.h>
#include <stdint.h>
#include "highs_c_api.h"
#include "callbacks.h"
*/
import "C"
import "unsafe"

// Callback can be used with a solver to intercept internal optimization steps.
type Callback func(CallbackType, string, CallbackData) CallbackResult

// CallbackType indicates the HiGHS event type associated with callback.
type CallbackType int

const (
	CallbackTypeLogging CallbackType = iota
	CallbackTypeSimplexInterrupt
	CallbackTypeIpmInterrupt
	CallbackTypeMipSolution
	CallbackTypeMipImprovingSolution
	CallbackTypeMipLogging
	CallbackTypeMipInterrupt
	CallbackTypeMipGetCutPool
	CallbackTypeMipDefineLazyConstraints
	CallbackTypeMipUserSolution
)

func (c CallbackType) String() string {
	switch c {
	case CallbackTypeLogging:
		return "Logging"
	case CallbackTypeSimplexInterrupt:
		return "SimplexInterrupt"
	case CallbackTypeIpmInterrupt:
		return "IpmInterrupt"
	case CallbackTypeMipSolution:
		return "MipSolution"
	case CallbackTypeMipImprovingSolution:
		return "MipImprovingSolution"
	case CallbackTypeMipLogging:
		return "MipLogging "
	case CallbackTypeMipInterrupt:
		return "MipInterrupt"
	case CallbackTypeMipGetCutPool:
		return "MipGetCutPool"
	case CallbackTypeMipDefineLazyConstraints:
		return "MipDefineLazyConstraints"
	case CallbackTypeMipUserSolution:
		return "MipUserSolution"
	default:
		return "Unknown"
	}
}

func (c CallbackType) toC() C.HighsInt {
	switch c {
	case CallbackTypeLogging:
		return C.kHighsCallbackLogging
	case CallbackTypeSimplexInterrupt:
		return C.kHighsCallbackSimplexInterrupt
	case CallbackTypeIpmInterrupt:
		return C.kHighsCallbackIpmInterrupt
	case CallbackTypeMipSolution:
		return C.kHighsCallbackMipSolution
	case CallbackTypeMipImprovingSolution:
		return C.kHighsCallbackMipImprovingSolution
	case CallbackTypeMipLogging:
		return C.kHighsCallbackMipLogging
	case CallbackTypeMipInterrupt:
		return C.kHighsCallbackMipInterrupt
	case CallbackTypeMipGetCutPool:
		return C.kHighsCallbackMipGetCutPool
	case CallbackTypeMipDefineLazyConstraints:
		return C.kHighsCallbackMipDefineLazyConstraints
	case CallbackTypeMipUserSolution:
		return C.kHighsCallbackCallbackMipUserSolution
	default:
		return C.kHighsCallbackLogging
	}
}

func callbackTypeFromC(status C.HighsInt) CallbackType {
	switch status {
	case C.kHighsCallbackLogging:
		return CallbackTypeLogging
	case C.kHighsCallbackSimplexInterrupt:
		return CallbackTypeSimplexInterrupt
	case C.kHighsCallbackIpmInterrupt:
		return CallbackTypeIpmInterrupt
	case C.kHighsCallbackMipSolution:
		return CallbackTypeMipSolution
	case C.kHighsCallbackMipImprovingSolution:
		return CallbackTypeMipImprovingSolution
	case C.kHighsCallbackMipLogging:
		return CallbackTypeMipLogging
	case C.kHighsCallbackMipInterrupt:
		return CallbackTypeMipInterrupt
	case C.kHighsCallbackMipGetCutPool:
		return CallbackTypeMipGetCutPool
	case C.kHighsCallbackMipDefineLazyConstraints:
		return CallbackTypeMipDefineLazyConstraints
	case C.kHighsCallbackCallbackMipUserSolution:
		return CallbackTypeMipUserSolution
	default:
		return CallbackTypeLogging
	}
}

// CallbackData is passed to callback with internal solver state.
type CallbackData struct {
	CallbackType                CallbackType // callback_type
	LogType                     int32        // log_type HighsLogType
	RunningTime                 float64      // running_time
	SimplexIterationCount       int32        // simplex_iteration_count
	IpmIterationCount           int32        // ipm_iteration_count
	PdlpIterationCount          int32        // pdlp_iteration_count
	ObjectiveFunctionValue      float64      // objective_function_value
	MIPNodeCount                int64        // mip_node_count
	MIPTotalLpIterations        int64        // mip_total_lp_iterations
	MIPPrimalBound              float64      // mip_primal_bound
	MIPDualBound                float64      // mip_dual_bound
	MIPGap                      float64      // mip_gap
	MIPSolution                 []float64    // mip_solution
	CutPool                     CallbackCutPool
	ExternalSolutionQueryOrigin int32 // external_solution_query_origin ExternalMipSolutionQueryOrigin
}

type CallbackCutPool struct {
	NumColumns int32
	UpperLower []CutPoolUpperLower // cutpool_num_cut...
	Matrix     []Nonzero           // cutpool_num_nz...
}

type CutPoolUpperLower struct {
	Lower float64
	Upper float64
}

// CallbackResult is returned from callback to solver, can influence process.
// The zero value is always safe to return.
type CallbackResult struct {
	// UserInterrupt is only permitted for certain CallbackType events, otherwise may trigger a solver error.
	UserInterrupt bool

	// UserHasSolution and UserSolutionColValues are only checked for CallbackTypeMipUserSolution.
	UserHasSolution       bool
	UserSolutionColValues []float64
}

// goHighsCallbackExportedBridge is a shared function called by from the C side of the bridge,
// it dispatches the event to the relevant Solver instance.
//
//export goHighsCallbackExportedBridge
func goHighsCallbackExportedBridge(solverReference C.HighsInt, c_callback_type C.HighsInt, c_message *C.char, c_data_out *C.HighsCallbackDataOut, c_data_in *C.HighsCallbackDataIn) {
	solver := solverReferenceArray[solverReference]
	if solver == nil {
		return
	}

	callback := solver.callback
	if callback == nil {
		return
	}

	callback_type := callbackTypeFromC(c_callback_type)
	message := C.GoString(c_message)

	data := CallbackData{}
	data.CallbackType = callback_type
	data.LogType = int32(c_data_out.log_type)
	data.RunningTime = float64(c_data_out.running_time)
	data.SimplexIterationCount = int32(c_data_out.simplex_iteration_count)
	data.IpmIterationCount = int32(c_data_out.ipm_iteration_count)
	data.PdlpIterationCount = int32(c_data_out.pdlp_iteration_count)
	data.ObjectiveFunctionValue = float64(c_data_out.objective_function_value)
	data.MIPNodeCount = int64(c_data_out.mip_node_count)
	data.MIPTotalLpIterations = int64(c_data_out.mip_total_lp_iterations)
	data.MIPPrimalBound = float64(c_data_out.mip_primal_bound)
	data.MIPDualBound = float64(c_data_out.mip_dual_bound)
	data.MIPGap = float64(c_data_out.mip_gap)
	data.MIPSolution = buildMIPSolution(c_data_out.mip_solution_size, c_data_out.mip_solution)
	data.CutPool.NumColumns = int32(c_data_out.cutpool_num_col)
	data.CutPool.UpperLower = buildCutPoolUpperLower(c_data_out.cutpool_num_cut,
		c_data_out.cutpool_lower, c_data_out.cutpool_upper)
	data.CutPool.Matrix = buildCutPoolMatrix(c_data_out.cutpool_num_cut, c_data_out.cutpool_num_nz,
		c_data_out.cutpool_start, c_data_out.cutpool_index, c_data_out.cutpool_value)
	data.ExternalSolutionQueryOrigin = int32(c_data_out.external_solution_query_origin)

	inputs := callback(callback_type, message, data)

	if inputs.UserInterrupt {
		c_data_in.user_interrupt = 1
	} else {
		c_data_in.user_interrupt = 0
	}
	if inputs.UserHasSolution {
		c_data_in.user_has_solution = 1
	} else {
		c_data_in.user_has_solution = 0
	}
	if len(inputs.UserSolutionColValues) > 0 {
		if len(inputs.UserSolutionColValues) != int(c_data_in.user_solution_size) {
			panic("user solution length doesn't match expected size")
		}
		var solution []C.double = unsafe.Slice(c_data_in.user_solution, c_data_in.user_solution_size)
		for i := range c_data_in.user_solution_size {
			solution[i] = C.double(inputs.UserSolutionColValues[i])
		}
	}
}

func buildMIPSolution(mip_solution_size C.HighsInt, mip_solution *C.double) []float64 {
	if mip_solution != nil && mip_solution_size > 0 {
		var cSolutionArray []C.double = unsafe.Slice(mip_solution, mip_solution_size)
		mipSolution := make([]float64, mip_solution_size)
		for i := range mip_solution_size {
			mipSolution[i] = float64(cSolutionArray[i])
		}
		return mipSolution
	}
	return nil
}

func buildCutPoolUpperLower(num_cut C.HighsInt, lower_ptr *C.double, upper_ptr *C.double) []CutPoolUpperLower {
	if num_cut > 0 && lower_ptr != nil && upper_ptr != nil {
		var cLowerArray []C.double = unsafe.Slice(lower_ptr, num_cut)
		var cUpperArray []C.double = unsafe.Slice(upper_ptr, num_cut)
		upperLowerSlice := make([]CutPoolUpperLower, num_cut)
		for i := range num_cut {
			upperLowerSlice[i] = CutPoolUpperLower{
				Lower: float64(cLowerArray[i]),
				Upper: float64(cUpperArray[i]),
			}
		}
		return upperLowerSlice
	}
	return nil
}

func buildCutPoolMatrix(num_cut C.HighsInt, num_nz C.HighsInt, start_ptr *C.HighsInt, index_ptr *C.HighsInt, value_ptr *C.double) []Nonzero {
	if num_cut > 0 && num_nz > 0 && start_ptr != nil && index_ptr != nil && value_ptr != nil {
		var cStartArray []C.HighsInt = unsafe.Slice(start_ptr, num_cut+1)
		var cIndexArray []C.HighsInt = unsafe.Slice(index_ptr, num_nz)
		var cValueArray []C.double = unsafe.Slice(value_ptr, num_nz)
		nonZeroSlice := make([]Nonzero, 0, num_nz)
		for cutRow := range num_cut {
			for i := cStartArray[cutRow]; i < cStartArray[cutRow+1]; i++ {
				colNum := cIndexArray[i]
				value := cValueArray[i]
				nonZeroSlice = append(nonZeroSlice, Nonzero{
					Row: int(cutRow),
					Col: int(colNum),
					Val: float64(value),
				})
			}
		}
		return nonZeroSlice
	}
	return nil
}

func (s *Solver) InterruptSupportEnable() error {
	closeMutex.RLock()
	defer closeMutex.RUnlock()

	s.callback = nil
	s.Interrupted = false
	status := Status(C.GoHighsInterruptEnable(s.ptr, C.HighsInt(s.refNum)))
	return newError("EnableInterruptSupport", status)
}

func (s *Solver) InterruptSupportDisable() error {
	closeMutex.RLock()
	defer closeMutex.RUnlock()

	s.callback = nil
	s.Interrupted = false
	status := Status(C.GoHighsInterruptDisable(s.ptr, C.HighsInt(s.refNum)))
	return newError("DisableInterruptSupport", status)
}

func (s *Solver) InterruptSetFlag(value bool) error {
	closeMutex.RLock()
	defer closeMutex.RUnlock()

	s.Interrupted = value
	var cValue C.HighsInt = 0
	if value {
		cValue = 1
	}
	status := Status(C.GoHighsInterruptSetFlag(s.ptr, C.HighsInt(s.refNum), cValue))
	return newError("InterruptSetFlag", status)
}

func (s *Solver) SetCallback(callback Callback, callbackTypes []CallbackType) error {
	closeMutex.RLock()
	defer closeMutex.RUnlock()

	s.callback = callback
	s.Interrupted = false

	var cCallbackTypes []C.HighsInt
	var pCallbackTypes *C.HighsInt
	if len(callbackTypes) > 0 {
		cCallbackTypes = make([]C.HighsInt, len(callbackTypes))
		for i, cb := range callbackTypes {
			cCallbackTypes[i] = cb.toC()
		}
		pCallbackTypes = (*C.HighsInt)(&cCallbackTypes[0])
	}

	status := Status(C.GoHighsCallbackBridgedEnable(s.ptr, C.HighsInt(s.refNum),
		C.HighsInt(len(callbackTypes)), pCallbackTypes))
	return newError("SetCallback", status)
}

func (s *Solver) ClearCallback() error {
	closeMutex.RLock()
	defer closeMutex.RUnlock()

	s.callback = nil
	s.Interrupted = false
	status := Status(C.GoHighsCallbackBridgedDisable(s.ptr, C.HighsInt(s.refNum)))
	return newError("ClearCallback", status)
}
