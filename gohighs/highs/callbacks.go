package highs

type HighsCallback func(CallbackType, string, HighsCallbackDataOut) HighsCallbackDataIn

// names don't follow Go convention but are left exactly matching high's C API so that documentation aligns.
type HighsCallbackDataOut struct {
	callback_type            CallbackType
	log_type                 int32
	running_time             float64
	simplex_iteration_count  int32
	ipm_iteration_count      int32
	pdlp_iteration_count     int32
	objective_function_value float64
	mip_node_count           int64
	mip_total_lp_iterations  int64
	mip_primal_bound         float64
	mip_dual_bound           float64
	mip_gap                  float64
	mip_solution             []float64
}

type HighsCallbackDataIn struct {
	user_interrupt    bool
	user_has_solution bool
	user_solution     []float64
}
