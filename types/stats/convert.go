package stats

import (
	"fmt"
	"strconv"
	"strings"
)

func (reforge *ReforgeRecipe) Str() string {
	return fmt.Sprintf("(%s -> %s)", reforge.From.Name(), reforge.To.Name())
}

func (block *StatBlock) String() string {
	build := strings.Builder{}
	first := true
	build.WriteString("{")

	for i, value := range block {
		if value != 0 {
			var stat StatType = StatType(i)
			name := stat.Name()

			if first {
				first = false
			} else {
				build.WriteRune(' ')
			}

			build.WriteString(name)
			build.WriteRune('=')
			build.WriteString(strconv.FormatUint(uint64(value), 10))
		}
	}

	build.WriteString("}")
	return build.String()
}
