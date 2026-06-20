/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/*                                                                       */
/*    This file is part of the HiGHS linear optimization suite           */
/*                                                                       */
/*    Available as open-source under the MIT License                     */
/*                                                                       */
/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/**@file lp_data/HighsCallback.h
 * @brief
 */
#ifndef LP_DATA_HIGHSCALLBACK_H_
#define LP_DATA_HIGHSCALLBACK_H_

// forward declaration to avoid circular dependency
enum class HighsStatus;
// class Highs;

// #include "lp_data/HStruct.h"
#include <stdbool.h>
#include "lp_data/HConst.h"
#include "lp_data/HighsCallbackStruct.h"

enum ExternalMipSolutionQueryOrigin {
  kExternalMipSolutionQueryOriginAfterSetup = 0,
  kExternalMipSolutionQueryOriginBeforeDive,
  kExternalMipSolutionQueryOriginEvaluateRootNode0,
  kExternalMipSolutionQueryOriginEvaluateRootNode1,
  kExternalMipSolutionQueryOriginEvaluateRootNode2,
  kExternalMipSolutionQueryOriginEvaluateRootNode3,
  kExternalMipSolutionQueryOriginEvaluateRootNode4
};

/**
 * Struct to handle callback output data
 */
struct HighsCallbackOutput {
  void* highs;
  enum HighsLogType log_type;
  double running_time;
  HighsInt simplex_iteration_count;
  HighsInt ipm_iteration_count;
  HighsInt pdlp_iteration_count;
  double objective_function_value;
  int64_t mip_node_count;
  int64_t mip_total_lp_iterations;
  double mip_primal_bound;
  double mip_dual_bound;
  double mip_gap;
  // std::vector<double> mip_solution;
  // HighsInt cutpool_num_col;
  // HighsInt cutpool_num_cut;
  // std::vector<HighsInt> cutpool_start;
  // std::vector<HighsInt> cutpool_index;
  // std::vector<double> cutpool_value;
  // std::vector<double> cutpool_lower;
  // std::vector<double> cutpool_upper;
  // ExternalMipSolutionQueryOrigin external_solution_query_origin;

  // operator HighsCallbackDataOut() const;
};

struct HighsCallbackInput {
  void* highs;
  bool user_interrupt;
  bool user_has_solution;
  // std::vector<double> user_solution;

  // HighsStatus setSolution(HighsInt num_entries, const double* value);

  // HighsStatus setSolution(HighsInt num_entries, const HighsInt* index,
  //                         const double* value);

  // HighsStatus repairSolution();

  // operator HighsCallbackDataIn() const;
  // HighsCallbackInput operator=(const HighsCallbackDataIn& data_in);
};

#endif /* LP_DATA_HIGHSCALLBACK_H_ */
