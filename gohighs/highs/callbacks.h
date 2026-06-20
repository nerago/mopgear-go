#ifndef GOHIGHS_CALLBACKS_H
#define GOHIGHS_CALLBACKS_H

#include "highs_c_api.h"

#define GoHighsMaxSolverReference 1023

HighsInt GoHighsCallbackBridgedEnable(void* highs, HighsInt solverReference);
HighsInt GoHighsCallbackBridgedDisable(void* highs, HighsInt solverReference);
HighsInt GoHighsInterruptEnable(void* highs, HighsInt solverReference);
HighsInt GoHighsInterruptDisable(void* highs, HighsInt solverReference);
HighsInt GoHighsInterruptSetFlag(void* highs, HighsInt solverReference, HighsInt flagValue);

#endif
