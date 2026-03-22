#include "textflag.h"
#include "go_asm.h"

TEXT ·StatBlock_Equals_Within_Assembly(SB), NOSPLIT|NOFRAME, $0-0
    // args in CX DX
    // reorder operation to not need any more general registers

    // first 256 bits
    VMOVDQU          (CX), Y0
    VPCMPEQD         (DX), Y0, Y2  // 256 bits, 32 bit units, 8 blocks of 32, all set to ones/zeros
    
    // remaining 128 bits
    VMOVDQU        32(CX), X1
    VPCMPEQD       32(DX), X1, X3  // 128 bits, 32 bit units, 4 blocks of 32, all set to ones/zeros
    
    // build result    
    VPMOVMSKB          Y2, CX      // extract top bit from each byte -> 32 bits worth
    VPMOVMSKB          X3, DX      // extract top bit from each byte -> 16 bits worth
    ORQ       $0xFFFF0000, DX      // extend to 32 bits too
    ANDL               AX, DX      // equality is all bits sets
    SUBL      $0xFFFFFFFF, DX      

    // result in flags

    RET



TEXT ·FullItemSet_Equals_Assem(SB), NOSPLIT|NOFRAME, $0-17
    MOVQ        a+0(FP), R8
    MOVQ        b+8(FP), R9

    // should be Nop if offset still zero
    LEAQ        FullItemSet_items(R8), R8
    LEAQ        FullItemSet_items(R9), R9

    // index 0
    XORL        SI, SI

item_loop:
    MOVQ        (R8)(SI*8), AX
    MOVQ        (R9)(SI*8), BX

    // null checks
    TESTQ       AX, AX
    JEQ         first_null
    TESTQ       BX, BX
    JEQ         result_false

    MOVL        FullItem_fullItem_common+fullItem_common_Ref+ItemRef_ItemId(AX), DI
    CMPL        FullItem_fullItem_common+fullItem_common_Ref+ItemRef_ItemId(BX), DI
    JNE         result_false

    MOVW        FullItem_fullItem_common+fullItem_common_Ref+ItemRef_ItemLevel(AX), DI
    CMPW        FullItem_fullItem_common+fullItem_common_Ref+ItemRef_ItemLevel(BX), DI
    JNE         result_false
    
    MOVBLZX     FullItem_fullItem_common+fullItem_common_Slot(AX), DI
    CMPB        FullItem_fullItem_common+fullItem_common_Slot(BX), DI
    JNE         result_false

    LEAQ        FullItem_fullItem_common+fullItem_common_StatBase(AX), CX
    LEAQ        FullItem_fullItem_common+fullItem_common_StatBase(BX), DX
    CALL        ·StatBlock_Equals_Within_Assembly(SB)
    JNE         result_false

    LEAQ        FullItem_fullItem_common+fullItem_common_StatEnchant(AX), CX
    LEAQ        FullItem_fullItem_common+fullItem_common_StatEnchant(BX), DX
    CALL        ·StatBlock_Equals_Within_Assembly(SB)
    JNE         result_false

loop_next:
    INCL        SI
    CMPL        SI, $16
    JEQ         result_true
    JMP         item_loop

first_null:
    TESTQ       BX, BX
    JNE         result_false
    JMP         loop_next

result_true:
    MOVL         $1, AX
    RET

result_false:
    MOVL         $0, AX
    RET



TEXT ·SolvableItemSet_AddItem_Mutating(SB), NOSPLIT|NOFRAME, $0-17
    MOVQ        set+0(FP), AX
    MOVBLZX     slot+8(FP), DI
    MOVQ        item+9(FP), BX

    // set.items[slot] = item
    MOVQ        BX, SolvableItemSet_items(AX)(DI*8)

    // might be zero offsets
    LEAQ        SolvableItemSet_total(AX), AX
    LEAQ        SolvableItem_total(BX), BX

    // ·StatBlock_Increment_Mutating_Within_Assembly
    VMOVDQU      (AX), Y0
    VPADDD       (BX), Y0, Y0
    VMOVDQU        Y0, (AX)

    VMOVDQU    32(AX), X1
    VPADDD     32(BX), X1, X1
    VMOVDQU        X1, 32(AX)

    RET

    

TEXT ·MakeSetFromArraysAndAdvance4(SB), NOSPLIT|NOFRAME, $0-24
    MOVQ        slotOptions+0(FP), AX
    MOVQ        slotIndexes+8(FP), R9
    MOVQ        itemSet+16(FP), R10
    
    // zero out totals
    VMOVDQU     Y15, Y0
    VMOVDQU     X15, X1

    // slot index = 0
    XORL        SI, SI

firstloop:
    // slotOptions is array of slices, 24 each
    MOVQ        (AX), DX                        // start of options array
    MOVQ        8(AX), CX                       // slotSize
    ADDQ        $24, AX

    // if slotSize
    CMPQ        CX, $1
    JL          next_firstloop
    JEQ         add_item_firstloop   // slotSize 1, item is just start of array, already in DX

choose_firstloop:
    MOVL        (R9)(SI*4), DI                  // index := slotIndexes[slot]
    IMUL3Q      $SolvableItem__size, DI, BX
    ADDQ        BX, DX                          // item := &options[index]

    // advance
    INCL        DI                              // index++
    CMPL        DI, CX
    JL          done_advancing
    MOVL        $0, (R9)(SI*4)

add_item_firstloop:
    MOVQ        DX, SolvableItemSet_items(R10)(SI*8)   // set.items[slot] = item
    VPADDD      SolvableItem_total(DX), Y0, Y0         // add to running totals
    VPADDD      SolvableItem_total+32(DX), X1, X1

next_firstloop:
    INCL        SI
    CMPL        SI, $16
    JEQ         save_result
    JMP         firstloop

done_advancing:
    // finish off current
    MOVL        DI, (R9)(SI*4)                         // slotIndexes[slot] = index
    // abbreviated add_item
    MOVQ        DX, SolvableItemSet_items(R10)(SI*8)   // set.items[slot] = item
    VPADDD      SolvableItem_total(DX), Y0, Y0         // add to running totals
    VPADDD      SolvableItem_total+32(DX), X1, X1
    // abbreviated next_loop
    INCL        SI
    CMPL        SI, $16
    JEQ         save_result
    
secondloop:
    MOVQ        (AX), DX                        // start of options array
    MOVQ        8(AX), CX                       // slotSize
    ADDQ        $24, AX

    // if slotSize
    CMPQ        CX, $1
    JL          next_secondloop
    JEQ         add_item_firstloop   // slotSize 1, item is just start of array, already in DX

choose_secondloop:
    MOVL        (R9)(SI*4), DI                  // index := slotIndexes[slot]
    IMUL3Q      $SolvableItem__size, DI, BX
    ADDQ        BX, DX                          // item := &options[index]

add_item_secondloop:
    MOVQ        DX, SolvableItemSet_items(R10)(SI*8)   // set.items[slot] = item
    VPADDD      SolvableItem_total(DX), Y0, Y0
    VPADDD      SolvableItem_total+32(DX), X1, X1

next_secondloop:
    INCL        SI
    CMPL        SI, $16
    JNE         secondloop

save_result:    
    VMOVDQU        Y0, SolvableItemSet_total(R10)
    VMOVDQU        X1, SolvableItemSet_total+32(R10)
    RET
