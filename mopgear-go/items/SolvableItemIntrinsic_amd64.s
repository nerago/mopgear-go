#include "textflag.h"
#include "go_asm.h"

TEXT ·SolvableItemSet_RecalculateTotal(SB), NOSPLIT|NOFRAME, $0-8
    MOVQ        set+0(FP), AX

    // zero out totals
    VMOVDQU     Y15, Y0
    VMOVDQU     X15, X1

    // slot index = 0
    XORL        SI, SI

item_loop:
    MOVQ        SolvableItemSet_items(AX)(SI*8), BX    // item = set.items[si]
    TESTQ       BX, BX
    JEQ         next_loop
    VPADDD      SolvableItem_total(BX), Y0, Y0         // add to running totals
    VPADDD      SolvableItem_total+32(BX), X1, X1

next_loop:
    INCL        SI
    CMPL        SI, $16
    JNE         item_loop

    VMOVDQU     Y0, SolvableItemSet_total(AX)
    VMOVDQU     X1, SolvableItemSet_total+32(AX)
    RET
