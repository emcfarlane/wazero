package amd64

// CompileStackGrowCallSequence implements backend.Machine.
func (m *machine) CompileStackGrowCallSequence() []byte {
	// TODO
	ud2 := m.allocateInstr().asUD2()
	m.encodeWithoutRelResolution(ud2)
	return m.c.Buf()
}

// SetupPrologue implements backend.Machine.
func (m *machine) SetupPrologue() {
	cur := m.ectx.RootInstr
	prevInitInst := cur.next

	// At this point, we have the stack layout as follows:
	//
	//                   (high address)
	//                 +-----------------+ <----- RBP (somewhere in the middle of the stack)
	//                 |     .......     |
	//                 |      ret Y      |
	//                 |     .......     |
	//                 |      ret 0      |
	//                 |      arg X      |
	//                 |     .......     |
	//                 |      arg 1      |
	//                 |      arg 0      |
	//                 |   Return Addr   |
	//       RSP ----> +-----------------+
	//                    (low address)

	// First, we push the RBP, and update the RBP to the current RSP.
	// 		push %rbp
	// 		mov %rsp, %rbp
	cur = linkInstr(cur, m.allocateInstr().asPush64(newOperandReg(rbpVReg)))
	cur = linkInstr(cur, m.allocateInstr().asMovRR(rspVReg, rbpVReg, true))

	//                   (high address)                     (high address)
	//       RBP ----> +-----------------+                +-----------------+
	//                 |     .......     |                |     .......     |
	//                 |      ret Y      |                |      ret Y      |
	//                 |     .......     |                |     .......     |
	//                 |      ret 0      |                |      ret 0      |
	//                 |      arg X      |                |      arg X      |
	//                 |     .......     |     ====>      |     .......     |
	//                 |      arg 1      |                |      arg 1      |
	//                 |      arg 0      |                |      arg 0      |
	//                 |   Return Addr   |                |   Return Addr   |
	//       RSP ----> +-----------------+                |    Frame Addr   |
	//                    (low address)                   +-----------------+ <----- RSP, RBP
	//
	if !m.stackBoundsCheckDisabled { //nolint
		// TODO: stack bounds check
	}

	if regs := m.clobberedRegs; len(regs) > 0 {
		panic("TODO: save clobbered registers")
	}

	if size := m.spillSlotSize; size > 0 {
		panic("TODO: allocate spill slots")
	}

	linkInstr(cur, prevInitInst)
}

// SetupEpilogue implements backend.Machine.
func (m *machine) SetupEpilogue() {
	ectx := m.ectx
	for cur := ectx.RootInstr; cur != nil; cur = cur.next {
		if cur.kind == ret {
			m.setupEpilogueAfter(cur.prev)
			continue
		}

		// Removes the redundant copy instruction.
		// TODO: doing this in `SetupEpilogue` seems weird. Find a better home.
		if cur.IsCopy() && cur.op1.r.RealReg() == cur.op2.r.RealReg() {
			prev, next := cur.prev, cur.next
			// Remove the copy instruction.
			prev.next = next
			if next != nil {
				next.prev = prev
			}
		}
	}
}

func (m *machine) setupEpilogueAfter(cur *instruction) {
	prevNext := cur.next

	// At this point, we have the stack layout as follows:
	//
	//            (high address)
	//          +-----------------+
	//          |     .......     |
	//          |      ret Y      |
	//          |     .......     |
	//          |      ret 0      |
	//          |      arg X      |
	//          |     .......     |
	//          |      arg 1      |
	//          |      arg 0      |
	//          |   ReturnAddress |
	//          |   Caller_RBP    |
	//          +-----------------+ <--- RBP
	//          |    clobbered M  |
	//          |   ............  |
	//          |    clobbered 1  |
	//          |    clobbered 0  |
	//          |   spill slot N  |
	//          |   ............  |
	//          |   spill slot 0  |
	//          +-----------------+ <--- RSP
	//             (low address)

	if size := m.spillSlotSize; size > 0 {
		panic("TODO: deallocate spill slots")
	}
	if regs := m.clobberedRegs; len(regs) > 0 {
		panic("TODO: restore clobbered registers")
	}

	// Now roll back the RSP to the return address, and pop the RBP.
	// 		mov  %rbp, %rsp
	// 		pop  %rbp
	cur = linkInstr(cur, m.allocateInstr().asMovRR(rbpVReg, rspVReg, true))
	cur = linkInstr(cur, m.allocateInstr().asPop64(rbpVReg))

	linkInstr(cur, prevNext)
}
