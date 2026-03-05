#include "textflag.h"

TEXT ·StatBlock_Increment_Mutating(SB), NOSPLIT|NOFRAME, $0-16
    MOVQ    block+0(FP), AX
    MOVQ    other+8(FP), BX
 
    VMOVDQU (AX),Y0
    VPADDD  (BX),Y0,Y0
    VMOVDQU Y0,(AX)

    VMOVDQU 32(AX),X1
    VPADDD  32(BX),X1,X1
    VMOVDQU X1,32(AX)

	RET

// active version
TEXT ·StatBlock_Add_Into(SB), NOSPLIT|NOFRAME, $0-24
    MOVQ    a+0(FP),  AX
    MOVQ    b+8(FP),  BX
    MOVQ    c+16(FP), CX
 
    VMOVDQU (AX),Y0
    VPADDD  (BX),Y0,Y0
    VMOVDQU Y0,(CX)

    VMOVDQU 32(AX),X1
    VPADDD  32(BX),X1,X1
    VMOVDQU X1,32(CX)

	RET
