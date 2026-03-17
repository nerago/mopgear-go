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

// active version
TEXT ·StatBlock_AddAndSubtract_Into(SB), NOSPLIT|NOFRAME, $0-32
    MOVQ        add1+0(FP), AX
    MOVQ        add2+8(FP), BX
    MOVQ   subtract+16(FP), CX
    MOVQ        out+24(FP), DX
 
    VMOVDQU           (AX), Y0
    VPADDD            (BX), Y0, Y0
    VPSUBD            (CX), Y0, Y0
    VMOVDQU             Y0, (DX)

    VMOVDQU         32(AX), X1
    VPADDD          32(BX), X1, X1
    VPSUBD          32(CX), X1, X1
    VMOVDQU             X1, 32(DX)

	RET

TEXT ·StatBlock_Equals(SB), NOSPLIT|NOFRAME, $0-17
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

TEXT ·StatBlock_MultiplyForTotalSum_Float(SB), NOSPLIT|NOFRAME, $0-20
    MOVQ          a+0(FP), AX
    MOVQ          b+8(FP), BX

    // first 8 values load and convert to float32
    VCVTDQ2PS        (AX), Y1
    VCVTDQ2PS        (BX), Y2 
    // next 4 values load and convert to float32
    VCVTDQ2PS      32(AX), X3
    VCVTDQ2PS      32(BX), X4

    // main dot products
    // immediate value means read all inputs, write into lowest item in output only
    VDPPS           $0xF1, Y1, Y2, Y5
    VDPPS           $0xF1, X3, X4, X6

    // add subtotals 
    VADDSS             X5, X6, X0 // add X6, Y5(low half)
    VEXTRACTF128       $1, Y5, X7 // grab Y5(top half)
    VADDSS             X0, X7, X0 // add Y5(top half)

    // result
    MOVSS              X0, ret+16(FP)

    RET

DATA permuteVectorPairs<>+0x00(SB)/4, $0x00000001
DATA permuteVectorPairs<>+0x04(SB)/4, $0x00000000
DATA permuteVectorPairs<>+0x08(SB)/4, $0x00000003
DATA permuteVectorPairs<>+0x0C(SB)/4, $0x00000002
DATA permuteVectorPairs<>+0x10(SB)/4, $0x00000005
DATA permuteVectorPairs<>+0x14(SB)/4, $0x00000004
DATA permuteVectorPairs<>+0x18(SB)/4, $0x00000007
DATA permuteVectorPairs<>+0x1C(SB)/4, $0x00000006
GLOBL permuteVectorPairs<>(SB), RODATA, $32

TEXT ·StatBlock_MultiplyForTotalSum_Int(SB), NOSPLIT|NOFRAME, $0-24
    MOVQ       a+0(FP), AX
    MOVQ       b+8(FP), BX

    VMOVDQU     permuteVectorPairs<>+0x00(SB), Y0  // load global vector for permute

    VMOVDQU       (AX), Y1       // Y1=ABCDEFGH load
    VMOVDQU       (BX), Y2       // Y2=abcdefgh load
    VPMULUDQ        Y1, Y2, Y3   // Y3=AaCcEeGg int32*int32->int64, skips every second
    VPERMD          Y1, Y0, Y1   // Y1=BADCFEHG swap adjacent pairs
    VPERMD          Y2, Y0, Y2   // Y2=badcfehg swap adjacent pairs
    VPMULUDQ        Y1, Y2, Y1   // Y1=BbDdFfHh int32*int32->int64, after swap hits other one second
    VPADDQ          Y3, Y1, Y3   // Y3=4 subtotals

    VMOVDQU     32(AX), X4       // X4 load
    VMOVDQU     32(BX), X5       // X5 load
    VPMULUDQ        X4, X5, X6   // X6=mult
    VPERMD          Y4, Y0, Y4   // Y4=swapped, note uses Y since X version doesn't exist
    VPERMD          Y5, Y0, Y5   // Y5=swapped
    VPMULUDQ        X4, X5, X7   // X7=mult
    VPADDQ          X6, X7, X7   // X7=2 subtotals

    VPADDQ          Y3, Y7, Y7   // Y7=4 finished subtotals

    // lower subtotal values
    VPEXTRQ           $0, X7, AX   // pull out extra subtotal
    VPEXTRQ           $1, X7, BX
    VEXTRACTI128      $1, Y7, X6
    VPEXTRQ           $0, X6, CX
    VPEXTRQ           $1, X6, DX

    ADDQ              DX, CX
    ADDQ              BX, AX
    ADDQ              CX, AX
    MOVQ              AX, ret+16(FP)

    RET
