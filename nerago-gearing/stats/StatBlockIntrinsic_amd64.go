package stats

func StatBlock_Add_Into(a, b, out *StatBlock)

func StatBlock_Increment_Mutating(mutate *StatBlock, other *StatBlock)

func StatBlock_Equals(a, b *StatBlock) bool

func StatBlock_AddAndSubtract_Into(add1, add2, subtract, out *StatBlock)

func StatBlock_MultiplyForTotalSum(a, b *StatBlock) float64
