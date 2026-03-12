#include "textflag.h"

TEXT ·StatBlock_Increment_Mutating(SB), NOSPLIT|NOFRAME, $0-16
    MOVQ mutate+0(FP), AX
    MOVQ  other+8(FP), BX
 
    VMOVDQU      (AX), Y0
    VPADDD       (BX), Y0, Y0
    VMOVDQU        Y0, (AX)

    VMOVDQU    32(AX), X1
    VPADDD     32(BX), X1,X1
    VMOVDQU        X1, 32(AX)

	RET

// active version
TEXT ·StatBlock_Add_Into(SB), NOSPLIT|NOFRAME, $0-24
    MOVQ      a+0(FP), AX
    MOVQ      b+8(FP), BX
    MOVQ   out+16(FP), CX
 
    VMOVDQU      (AX), Y0
    VPADDD       (BX), Y0, Y0
    VMOVDQU        Y0, (CX)

    VMOVDQU    32(AX), X1
    VPADDD     32(BX), X1, X1
    VMOVDQU        X1, 32(CX)

	RET

TEXT ·StatBlock_Equals(SB), NOSPLIT|NOFRAME, $0-24
    MOVQ          a+0(FP), AX
    MOVQ          b+8(FP), BX
    
    // first 256 bits
    VMOVDQU          (AX), Y0
    VPCMPEQD         (BX), Y0, Y2  // 256 bits, 32 bit units, 8 blocks of 32, all set to ones/zeros
    VPMOVMSKB          Y2, CX      // extract top bit from each byte -> 32 bits worth
    
    // remaining 128 bits
    VMOVDQU        32(AX), X1
    VPCMPEQD       32(BX), X1, X3  // 128 bits, 32 bit units, 4 blocks of 32, all set to ones/zeros
    VPMOVMSKB          X3, DX      // extract top bit from each byte -> 16 bits worth
    ORQ       $0xFFFF0000, DX      // extend to 32 bits too

    // build result
    ANDL               CX, DX      // equality is all bits sets
    SUBL      $0xFFFFFFFF, DX      
    SETEQ          ret+16(FP)

    RET
