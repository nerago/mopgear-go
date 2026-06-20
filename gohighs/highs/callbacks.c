
#include <stdlib.h>
#include <stdint.h>
#include "highs_c_api.h"
#include "callbacks.h"
#include "_cgo_export.h"

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
    HighsInt err = Highs_setCallback(highs, goHighsCallbackForInterrupt, (void*)(uintptr_t) solverReference);
    err |= Highs_startCallback(highs, kHighsCallbackSimplexInterrupt);
    err |= Highs_startCallback(highs, kHighsCallbackIpmInterrupt);
    err |= Highs_startCallback(highs, kHighsCallbackMipInterrupt);
    return err;
}

HighsInt GoHighsInterruptDisable(void* highs, HighsInt solverReference) {
    return Highs_setCallback(highs, nullptr, nullptr);
}

HighsInt GoHighsInterruptSetFlag(void* highs, HighsInt solverReference, HighsInt flagValue) {
    GoHighsInterruptFlags[solverReference] = flagValue;
    return 0;
}
