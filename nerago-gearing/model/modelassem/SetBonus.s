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

TEXT ·CalcBonusSolveAssem(SB), NOSPLIT, $16-60
    MOVQ		             equip+0(FP), AX 				// equip pointer
	MOVQ	        itemToSet_base+8(FP), DX                // itemToSet pointer

	MOVQ                              $0, (SP)              // zero out counts array on stack
	MOVQ                              $0, 8(SP)

equip_head:
    MOVQ          const_Equip_Head*8(AX), BX             	// BX = equip[Equip_Head]
	TESTQ                             BX, BX             	// BX != nil
	JEQ                                   equip_shoulder
	MOVL         SolvableItem_itemId(BX), SI				// SI = item.itemId
	MOVBLZX                   (DX)(SI*1), DI             	// DI entry = itemToSet_base[itemId]
	INCB                      (SP)(DI*1)                 	// counts[DI]++

equip_shoulder:    
    MOVQ      const_Equip_Shoulder*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   equip_chest
	MOVL         SolvableItem_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI
	INCB                      (SP)(DI*1)                  

equip_chest:    
    MOVQ         const_Equip_Chest*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   equip_hand
	MOVL         SolvableItem_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

equip_hand:    
    MOVQ          const_Equip_Hand*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   equip_leg
	MOVL         SolvableItem_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

equip_leg:    
    MOVQ           const_Equip_Leg*8(AX), BX           
	TESTQ                             BX, BX              
	JEQ                                   totals_init
	MOVL         SolvableItem_itemId(BX), SI              
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
	MULSS                     (AX)(SI*4), X0                // value *= activeSets[index][count]
	ADDQ                             $24, AX                // advance activeSets pointer to next entry = sizeof([6]float)
	INCL                              DI                    // index++
	
loop_condition:
	CMPQ                              AX, BX
	JNE                                   loop_body

return:
    MOVSS                             X0, ret+56(FP)
    RET
 
 
TEXT ·CalcBonusSolveAssemAssumeNonNull(SB), NOSPLIT, $16-60
    MOVQ		             equip+0(FP), AX 				// equip pointer
	MOVQ	        itemToSet_base+8(FP), DX                // itemToSet pointer

	MOVQ                              $0, (SP)              // zero out counts array on stack
	MOVQ                              $0, 8(SP)

    MOVQ          const_Equip_Head*8(AX), BX             	// BX = equip[Equip_Head]
	MOVL         SolvableItem_itemId(BX), SI				// SI = item.itemId
	MOVBLZX                   (DX)(SI*1), DI             	// DI entry = itemToSet_base[itemId]
	INCB                      (SP)(DI*1)                 	// counts[DI]++

    MOVQ      const_Equip_Shoulder*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI
	INCB                      (SP)(DI*1)                  

    MOVQ         const_Equip_Chest*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

    MOVQ          const_Equip_Hand*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

    MOVQ           const_Equip_Leg*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
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
	MULSS                     (AX)(SI*4), X0                // value *= activeSets[index][count]
	ADDQ                             $24, AX                // advance activeSets pointer to next entry = sizeof([6]float)
	INCL                              DI                    // index++
	
loop_condition:
	CMPQ                              AX, BX
	JNE                                   loop_body

return:
    MOVSS                             X0, ret+56(FP)
    RET
 


TEXT ·CalcBonusSolveAssemAssumeNonNullWithCases(SB), NOSPLIT, $16-60
	MOVQ		   activeSets_len+40(FP), CX                // CX = activeSets_len
	CMPQ                              CX, $1
	JL                                    return_none
	MOVQ		             equip+0(FP), AX 				// equip pointer
	MOVQ	        itemToSet_base+8(FP), DX                // itemToSet pointer
	JEQ                                   single_checks

multi_checks:
	MOVQ                              $0, (SP)              // zero out counts array on stack
	MOVQ                              $0, 8(SP)
	
    MOVQ          const_Equip_Head*8(AX), BX             	// BX = equip[Equip_Head]
	MOVL         SolvableItem_itemId(BX), SI				// SI = item.itemId
	MOVBLZX                   (DX)(SI*1), DI             	// DI entry = itemToSet_base[itemId]
	INCB                      (SP)(DI*1)                 	// counts[DI]++

    MOVQ      const_Equip_Shoulder*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI
	INCB                      (SP)(DI*1)                  

    MOVQ         const_Equip_Chest*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

    MOVQ          const_Equip_Hand*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

    MOVQ           const_Equip_Leg*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	MOVBLZX                   (DX)(SI*1), DI              
	INCB                      (SP)(DI*1)             

multi_totals_init:
	MOVSS	         float_value_1<>(SB), X0    			// value = 1.0
	XORL                              DI, DI                // DI index = 0
	MOVQ		  activeSets_base+32(FP), AX                // AX = activeSets_base
	LEAQ                      (AX)(CX*4), BX                // BX = activeSets_base + activeSets_len * sizeof(float)
	JMP                                   multi_loop_condition

multi_loop_body:
	MOVBLZX                  1(SP)(DI*1), SI                // SI = counts[index+1]
	MULSS                     (AX)(SI*4), X0                // value *= activeSets_curr[count]
	ADDQ                             $24, AX                // advance activeSets_curr pointer to next entry = sizeof([6]float)
	INCL                              DI                    // index++
	
multi_loop_condition:
	CMPQ                              AX, BX
	JNE                                   multi_loop_body
    MOVSS                             X0, ret+56(FP)
    RET

single_checks:
	XORL                              CX, CX                // repurposed CX = count

    MOVQ          const_Equip_Head*8(AX), BX             	// BX = equip[Equip_Head]
	MOVL         SolvableItem_itemId(BX), SI				// SI = item.itemId
	ADDB                      (DX)(SI*1), CX 				// count += itemToSet_base[itemId]

    MOVQ      const_Equip_Shoulder*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	ADDB                      (DX)(SI*1), CX

    MOVQ         const_Equip_Chest*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	ADDB                      (DX)(SI*1), CX         

    MOVQ          const_Equip_Hand*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	ADDB                      (DX)(SI*1), CX         

    MOVQ           const_Equip_Leg*8(AX), BX           
	MOVL         SolvableItem_itemId(BX), SI              
	ADDB                      (DX)(SI*1), CX
	
	MOVQ		  activeSets_base+32(FP), AX                // AX = activeSets_base
	MOVSS                     (AX)(DI*4), X0                // value = activeSets[count]
    MOVSS                             X0, ret+56(FP)
    RET

return_none:
	MOVSS	         float_value_1<>(SB), X0    			// value = 1.0
    MOVSS                             X0, ret+56(FP)
    RET
