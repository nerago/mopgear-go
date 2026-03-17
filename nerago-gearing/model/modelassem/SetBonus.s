#include "textflag.h"
#include "go_asm.h"


DATA float_value_1<>+0x00(SB)/4, $0x3f800000
GLOBL float_value_1<>(SB), RODATA, $4

// NOTE: skips range checks on:
//        * itemToSet lookup < 100000
//        * itemToSet entry fits in counts < 10
//        * set count fits in set bonus array < 6

// NOTE: optimiser maybe skips the item!=nil check

// stack looks like:
//             equip+0
//    itemToSet_base+8(FP)
//    itemToSet_len+16(FP)
//    itemToSet_cap+24(FP)
//  activeSets_base+32(FP)
//   activeSets_len+40(FP)
//   activeSets_cap+48(FP)
//              ret+56(FP)

TEXT ·CalcBonusSolveAssemX(SB), NOSPLIT, $104-60
    // allocate stack for the counts array
    // size needs to be aligned
    PUSHQ                                 BP
    MOVQ                              SP, BP
    SUBQ                             $16, SP

	// locals
    MOVQ		             equip+0(FP), AX 				// equip pointer
	MOVQ	        itemToSet_base+8(FP), DX
	
equip_head:
    MOVQ          const_Equip_Head*8(AX), BX             	// BX = equip[Equip_Head]
	TESTQ                             BX, BX             	// BX != nil
	JEQ                                   return
	MOVL         const_offset_itemId(BX), SI				// SI = item.itemId
	//MOVL         $SolvableItem_itemId*100000, SI
	//MOVL         48(BX), SI				// SI = item.itemId
	//MOVL  $0x12345665,SI
	MOVB                      (DX)(SI*1), DI             	// DI entry = itemToSet_base[itemId]
	//INCB                      (SP)(DI*1)                 	// counts[DI]++

return: 
    MOVSS                             X0, ret+56(FP)
    ADDQ                             $16, SP
    POPQ                                  BP
    RET



TEXT ·CalcBonusSolveAssemY(SB), NOSPLIT, $104-60
    // allocate stack for the counts array
    // size needs to be aligned
   // PUSHQ                                 BP
   // MOVQ                              SP, BP
  //  SUBQ                             $16, SP

	// locals
    MOVQ		             equip+0(FP), AX 				// equip pointer
	MOVQ	        itemToSet_base+8(FP), DX

	//MOVQ $0,AX
	//CALL	runtime·gopanic(SB)

	CMPQ AX,$0
	JEQ mypanic0

equip_head:
    MOVQ          const_Equip_Head*8(AX), BX             	// BX = equip[Equip_Head]
	TESTQ                             BX, BX             	// BX != nil
	JEQ                                   return
	MOVBLZX         const_offset_itemId(BX), SI				// SI = item.itemId
	CMPQ SI,$90000
	JCC mypanic1

	MOVBLZX                      (DX)(SI*1), DI             	// DI entry = itemToSet_base[itemId]
	CMPQ DI,$10
	JCC mypanic2

	INCB                      (SP)(DI*1)                 	// counts[DI]++            				// eoor after change to pointer type

return:
    MOVSS                             X0, ret+56(FP)
 //   ADDQ                             $80, SP
 //   POPQ                                  BP
    RET

mypanic0:
	MOVQ $0,AX
	CALL	runtime·gopanic(SB)
mypanic1:
	MOVQ $0,AX
	CALL	runtime·gopanic(SB)
mypanic2:
	MOVQ $0,AX
	CALL	runtime·gopanic(SB)



TEXT ·CalcBonusSolveAssem(SB), NOSPLIT, $104-60
	// locals
    MOVQ		             equip+0(FP), AX 				// equip pointer
	MOVQ	        itemToSet_base+8(FP), DX

equip_head:
    MOVQ          const_Equip_Head*8(AX), BX             	// BX = equip[Equip_Head]
	TESTQ                             BX, BX             	// BX != nil
	JEQ                                   equip_shoulder
	MOVL         const_offset_itemId(BX), SI				// SI = item.itemId
	MOVBLZX                   (DX)(SI*1), DI             	// DI entry = itemToSet_base[itemId]
	INCB                      (SP)(DI*1)                 	// counts[DI]++

equip_shoulder:    
    MOVQ      const_Equip_Shoulder*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   equip_chest
	MOVL         const_offset_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI
	INCB                      (SP)(DI*1)                  

equip_chest:    
    MOVQ         const_Equip_Chest*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   equip_hand
	MOVL         const_offset_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

equip_hand:    
    MOVQ          const_Equip_Hand*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   equip_leg
	MOVL         const_offset_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

equip_leg:    
    MOVQ           const_Equip_Leg*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   totals_init
	MOVL         const_offset_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

totals_init:
    MOVSS	         float_value_1<>(SB), X0    			// value = 1.0
	XORL                              DI, DI                // DI index = 0
	MOVQ		  activeSets_base+32(FP), AX                // AX = activeSets_base
	MOVQ		   activeSets_len+40(FP), BX                // BX = activeSets_len
	LEAQ                      (AX)(BX*4), BX                // BX = activeSets_base + activeSets_len * sizeof(float)
	JMP                                   loop_condition

loop_body:
	MOVBLZX                  1(SP)(DI*1), SI                // SI = counts[index+1]
    MOVSS                     (AX)(SI*4), X1                // value *= activeSets[index][count]
	MULSS                     (AX)(SI*4), X0                // value *= activeSets[index][count]
	ADDQ                             $24, AX                // advance activeSets pointer to next entry = sizeof([6]float)
	INCL                              DI                    // index++
	
loop_condition:
	CMPQ                              AX, BX
	JNE                                   loop_body

return:
    MOVSS                             X0, ret+56(FP)
    RET


TEXT ·CalcBonusSolveAssem0(SB), 0, $104-60
    // allocate stack for the counts array
    // size needs to be aligned
    PUSHQ                                 BP
    MOVQ                              SP, BP
    SUBQ                             $96, SP

	// locals
    MOVQ		             equip+0(FP), AX 				// equip pointer
	MOVQ	        itemToSet_base+8(FP), DX

equip_head:
    MOVQ          const_Equip_Head*8(AX), BX             	// BX = equip[Equip_Head]
	TESTQ                             BX, BX             	// BX != nil
	JEQ                                   equip_shoulder
	MOVL         SolvableItem_itemId(BX), SI				// SI = item.itemId
	MOVB                      (DX)(SI*1), DI             	// DI entry = itemToSet_base[itemId]
	INCB                      (SP)(DI*1)                 	// counts[DI]++            				// eoor after change to pointer type

equip_shoulder:    
    MOVQ      const_Equip_Shoulder*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   equip_chest
	MOVL         SolvableItem_itemId(BX), SI              
	MOVB                      (DX)(SI*1), DI             	//err 
	INCB                      (SP)(DI*1)                  

equip_chest:    
    MOVQ         const_Equip_Chest*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   equip_hand
	MOVL         SolvableItem_itemId(BX), SI              
	MOVB                      (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

equip_hand:    
    MOVQ          const_Equip_Hand*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   equip_leg
	MOVL         SolvableItem_itemId(BX), SI              
	MOVB                      (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

equip_leg:    
    MOVQ           const_Equip_Leg*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   totals_init
	MOVL         SolvableItem_itemId(BX), SI              
	MOVB                      (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

totals_init:
    MOVSS	         float_value_1<>(SB), X0    			// value = 1.0
	MOVQ		 activeSets_base+32(FP), AX                // AX = activeSets_base
	MOVQ		   activeSets_len+40(FP), BX                // BX = activeSets_len
	XORL                              DI, DI                // DI index = 0

totals_loop:
	MOVB                     1(SP)(DI*1), SI                // SI = counts[index+1]
    MULSS                     (AX)(SI*1), X0                // value *= activeSets[index][count]
	ADDQ                             $24, AX                // advance activeSets pointer to next entry
	INCL                              DI                    // index++
	CMPQ                              DI, BX
	JL                                    totals_loop

return:
    MOVSS                             X0, ret+56(FP)
    ADDQ                             $96, SP
    POPQ                                  BP
    RET

 
TEXT ·CalcBonusSolveAssemAssumeNonNull(SB), NOSPLIT, $16-60
    PUSHQ                                 BP
    MOVQ                              SP, BP
    SUBQ                             $16, SP

	// locals
    MOVQ	                 equip+0(FP), AX 				// equip pointer
	MOVQ		      itemToSet_base(FP), DX

equip_head:
    MOVQ          const_Equip_Head*8(AX), BX             	// BX = equip[Equip_Head]
	MOVL         SolvableItem_itemId(BX), SI				// SI = item.itemId
	MOVB                      (DX)(SI*1), DI             	// DI entry = itemToSet_base[itemId]
	INCB                      (SP)(DI*1)                 	// counts[DI]++

equip_shoulder:    
    MOVQ      const_Equip_Shoulder*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVB                      (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)                  

equip_chest:    
    MOVQ         const_Equip_Chest*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVB                      (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

equip_hand:    
    MOVQ          const_Equip_Hand*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVB                      (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

equip_leg:    
    MOVQ           const_Equip_Leg*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVB                      (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

totals_init:
    MOVSS	    const_float32_value1(SB), X0    			// value = 1.0
	MOVQ		 activeSets_base+32(FP), AX                // AX = activeSets_base
	MOVQ		   activeSets_len+40(FP), BX                // BX = activeSets_len
	XORL                              DI, DI                // DI index = 0

totals_loop:
	MOVB                     1(SP)(DI*1), SI                // SI = counts[index+1]
    MULSS                     (AX)(SI*1), X0                // value *= activeSets[index][count]
	ADDQ                             $24, AX                // advance activeSets pointer to next entry
	INCL                              DI                    // index++
	CMPQ                              DI, BX
	JL                                    totals_loop

return:
    MOVSS                             X0, ret+56(FP)
    ADDQ                             $16, SP
    POPQ                                  BP
    RET

 