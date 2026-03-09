#include "textflag.h"

TEXT ·StatBlock_Increment_Mutating(SB), NOSPLIT|NOFRAME, $0-16
    MOVQ  block+0(FP), AX
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
    MOVQ     c+16(FP), CX
 
    VMOVDQU      (AX), Y0
    VPADDD       (BX), Y0, Y0
    VMOVDQU        Y0, (CX)

    VMOVDQU    32(AX), X1
    VPADDD     32(BX), X1, X1
    VMOVDQU        X1, 32(CX)

	RET

TEXT ·StatBlock_Equals_Assem(SB), NOSPLIT|NOFRAME, $0-24
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




    //VPAND	X2, X3, X4 // OR PANDN for not
    //VPMOVMSKB X4, CX // 16 bits worth    0000???? 
	//VPMOVMSKB Y2, DX // on a byte basis so 32 bits worth      ????????
    
    //VTESTPD X2, X3

    //VMOVDQU    X4, (AX)

    //VMOVDQU  Y2, (AX)
    //MOVD  CX, (AX)


    // 128 bits of values, packed as 32 bit lots, mask is 4 bits worth
    
    //VPCMPEQB $4, 32(BX), X2, X3  // 32(BX) != X1 -> K1

    //EVEX.256.66.0F3A.W0 3E /r ib VPCMPUB k1 {k2}, ymm2, ymm3/m256, imm8
    //Compare packed unsigned byte values in ymm3/m256 and ymm2 using bits 2:0 of imm8 as a comparison predicate with writemask k2 and leave the result in mask register k1.

    // VPCMPB $79, Z9, Z9, K2, K1 
    //VPCMPB $81, X1, X21, K4, K5                        // 62f355043fe951
    //VPCMPEQB (BX), Y0, K0, K1

    // SETEQ in Go assembly (amd64) sets a byte register to 1 if the zero flag (ZF) is set (i.e., operands are equal) after a comparison, and 0 otherwise
    // SETEQ AL

