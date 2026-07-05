
#include <stdlib.h>
#include <stdint.h>
#include "highs_c_api.h"
#include "callbacks.h"
#include "_cgo_export.h"


// CALLBACKS FULLY BRIDGED OVER TO GO
void goHighsDetailCallbackForBridge(int callbackType, const char* message, const HighsCallbackDataOut* data_out, HighsCallbackDataIn* data_in, void* user_callback_data) {
	HighsInt solverReference = (HighsInt)(uintptr_t) user_callback_data;

    int log_type = data_out->log_type;
    double objective_value = data_out->objective_function_value;
    double mip_gap = data_out->mip_gap;
    double* mip_solution = data_out->mip_solution;
    int mip_solution_size = data_out->mip_solution_size;

//     data.log_type = static_cast<int>(log_type);
//       data.running_time = running_time;
//       data.simplex_iteration_count = simplex_iteration_count;
//       data.ipm_iteration_count = ipm_iteration_count;
//       data.pdlp_iteration_count = pdlp_iteration_count;
//       data.objective_function_value = objective_function_value;
//
//       data.mip_node_count = mip_node_count;
//       data.mip_total_lp_iterations = mip_total_lp_iterations;
//       data.mip_primal_bound = mip_primal_bound;
//       data.mip_dual_bound = mip_dual_bound;
//       data.mip_gap = mip_gap;
//       data.mip_solution_size = mip_solution.size();
//       data.mip_solution =
//           mip_solution.empty() ? nullptr : const_cast<double*>(mip_solution.data());

// 	const void* objective_function_value_p = Highs_getCallbackDataOutItem(data_out, kHighsCallbackDataOutObjectiveFunctionValueName);
//     double objective_function_value = *(double*)(objective_function_value_p);

//     Highs_getCallbackDataOutItem(data_out, kHighsCallbackDataOutMipGapName);


//   } else if (!strcmp(item_name, kHighsCallbackDataOutMipSolutionName)) {
//     return (void*)(data_out->mip_solution);

// double* mip_solution

// 	HighsInt userInterrupt = goHighsCallbackExportedBridge(solverReference, callbackType);
// 	data_in->user_interrupt = userInterrupt;
}

HighsInt GoHighsDetailCallbackBridgedEnable(void* highs, HighsInt solverReference) {
    HighsInt err = Highs_setCallback(highs, goHighsDetailCallbackForBridge, (void*)(uintptr_t) solverReference);
    err |= Highs_startCallback(highs, kHighsCallbackMipSolution);
    err |= Highs_startCallback(highs, kHighsCallbackMipImprovingSolution);
    return err;
}

// CALLBACKS FULLY BRIDGED OVER TO GO
void goHighsCallbackForBridge(int callbackType, const char* message, const HighsCallbackDataOut* data_out, HighsCallbackDataIn* data_in, void* user_callback_data) {
	HighsInt solverReference = (HighsInt)(uintptr_t) user_callback_data;
	HighsInt userInterrupt = goHighsCallbackExportedBridge(solverReference, callbackType);
	data_in->user_interrupt = userInterrupt;
}

HighsInt GoHighsCallbackBridgedEnable(void* highs, HighsInt solverReference) {
    HighsInt err = Highs_setCallback(highs, goHighsCallbackForBridge, (void*)(uintptr_t) solverReference);
    err |= Highs_startCallback(highs, kHighsCallbackSimplexInterrupt);
    err |= Highs_startCallback(highs, kHighsCallbackIpmInterrupt);
    err |= Highs_startCallback(highs, kHighsCallbackMipInterrupt);
    return err;
}

HighsInt GoHighsCallbackBridgedDisable(void* highs, HighsInt solverReference) {
    return Highs_setCallback(highs, nullptr, nullptr);
}

// CALLBACKS INTERNAL TO C, JUST CHECKING FLAG
static bool GoHighsInterruptFlags[GoHighsMaxSolverReference+1];

void goHighsCallbackForInterrupt(int callbackType, const char* message, const HighsCallbackDataOut* data_out, HighsCallbackDataIn* data_in, void* user_callback_data) {
	HighsInt solverReference = (HighsInt)(uintptr_t) user_callback_data;
	HighsInt userInterrupt = GoHighsInterruptFlags[solverReference];
	data_in->user_interrupt = userInterrupt;
}

HighsInt GoHighsInterruptEnable(void* highs, HighsInt solverReference) {
    GoHighsInterruptFlags[solverReference] = false;

    HighsInt err = Highs_setCallback(highs, goHighsCallbackForInterrupt, (void*)(uintptr_t) solverReference);
    err |= Highs_startCallback(highs, kHighsCallbackSimplexInterrupt);
    err |= Highs_startCallback(highs, kHighsCallbackIpmInterrupt);
    err |= Highs_startCallback(highs, kHighsCallbackMipInterrupt);
    return err;
}

HighsInt GoHighsInterruptDisable(void* highs, HighsInt solverReference) {
    GoHighsInterruptFlags[solverReference] = false;
    return Highs_setCallback(highs, nullptr, nullptr);
}

HighsInt GoHighsInterruptSetFlag(void* highs, HighsInt solverReference, HighsInt flagValue) {
    GoHighsInterruptFlags[solverReference] = flagValue;
    return 0;
}
