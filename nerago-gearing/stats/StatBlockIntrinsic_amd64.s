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

TEXT ·StatBlock_MultiplyForTotalSum(SB), NOSPLIT|NOFRAME, $0-24
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

    // convert subtotals to double
    VEXTRACTF128       $1, Y5, X7 // grab Y5(top half)
    VCVTSS2SD          X7, X7, X7 // convert top half to double
    VCVTSS2SD          X5, X5, X5 // convert lower half to double, old junk may still remain in upper part of Y5, ignore
    VCVTSS2SD          X6, X6, X6 // convert the single float to double

    // add subtotals 
    VADDSD             X5, X6, X0 // add X6, Y5(low half)
    VADDSD             X0, X7, X0 // add Y5(top half)

    // result
    MOVSD              X0, ret+16(FP)

    RET

TEXT ·StatBlock_StatBlockFloat_MultiplyForTotalSum(SB), NOSPLIT|NOFRAME, $0-24
    MOVQ          a+0(FP), AX
    MOVQ          b+8(FP), BX

    // load the float64 vector in parts as float32
    VCVTPD2PSY       (AX), X0
    VCVTPD2PSY     32(AX), X1
    VINSERTF128  $1, X0, Y1, Y1  // not sure about endian lineup
    VCVTPD2PSY      64(AX), X3

    // load the uint32 vector
    VCVTDQ2PS        (BX), Y2
    VCVTDQ2PS      32(BX), X4

    // main dot products
    // immediate value means read all inputs, write into lowest item in output only
    VDPPS           $0xF1, Y1, Y2, Y5
    VDPPS           $0xF1, X3, X4, X6

    // convert subtotals to double
    VEXTRACTF128       $1, Y5, X7 // grab Y5(top half)
    VCVTSS2SD          X7, X7, X7 // convert top half to double
    VCVTSS2SD          X5, X5, X5 // convert lower half to double, old junk may still remain in upper part of Y5, ignore
    VCVTSS2SD          X6, X6, X6 // convert the single float to double

    // add subtotals
    VADDSD             X5, X6, X0 // add X6, Y5(low half)
    VADDSD             X0, X7, X0 // add Y5(top half)

    // result
    MOVSD              X0, ret+16(FP)

    RET
