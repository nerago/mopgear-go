#include "textflag.h"

// active version
TEXT ·statBlock_Increment_Mutating(SB), NOSPLIT|NOFRAME, $0-16
    MOVQ    block+0(FP), AX
    MOVQ    other+8(FP), BX
 
    VMOVDQU (AX),Y0
    VPADDD  (BX),Y0,Y0
    VMOVDQU Y0,(AX)

    VMOVDQU 32(AX),X1
    VPADDD  32(BX),X1,X1
    VMOVDQU X1,32(AX)

	RET

// could be slightly better if we could align structs
// would need to be on 32 byte/256 bit boundary
TEXT ·StatBlock_Increment_Assem_Vec_Aligned(SB), NOSPLIT|NOFRAME, $0-16
    MOVQ    block+0(FP), AX
    MOVQ    other+8(FP), BX
 
    VMOVDQA (AX),Y0
    VPADDD  (BX),Y0,Y0
    VMOVDQA Y0,(AX)

    VMOVDQA 32(AX),X1
    VPADDD  32(BX),X1,X1
    VMOVDQA X1,32(AX)

	RET

// naive version, similar to go code
TEXT ·StatBlock_Increment_Loopy(SB), NOSPLIT|NOFRAME, $0-16
    MOVQ    block+0(FP), AX
    MOVQ    other+8(FP), BX
	XORL	CX, CX
	JMP	check_loop

loop_start:    
	MOVL	(AX)(CX*4), DX
	ADDL	(BX)(CX*4), DX
	MOVL	DX, (AX)(CX*4)
	INCQ	CX

check_loop:
	CMPQ	CX, $12
	JLT	loop_start

	RET