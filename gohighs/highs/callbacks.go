package highs

type HighsCallback func(CallbackType, string, HighsCallbackDataOut) HighsCallbackDataIn

// names don't follow Go convention but are left matching high's C API so that documentation aligns.
type HighsCallbackDataOut struct {
	Callback_type            CallbackType
	Log_type                 int32
	Running_time             float64
	Simplex_iteration_count  int32
	Ipm_iteration_count      int32
	Pdlp_iteration_count     int32
	Objective_function_value float64
	Mip_node_count           int64
	Mip_total_lp_iterations  int64
	Mip_primal_bound         float64
	Mip_dual_bound           float64
	Mip_gap                  float64
	Mip_solution             []float64
}

type HighsCallbackDataIn struct {
	User_interrupt    bool
	User_has_solution bool
	User_solution     []float64
}
