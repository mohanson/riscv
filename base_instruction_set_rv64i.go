package riscv

type RegisterRV64I struct {
	RG [32]uint64
	PC uint64
}

func ExecuterRV64I(r *RegisterRV64I, i uint64) int {
	switch {
	case i&0b_0000_0000_0000_0000_0000_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0011_0111: // LUI
	case i&0b_0000_0000_0000_0000_0000_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_00010_111: // AUIPC
		_, rd, imm := UType(i)
		DebuglnUType("AUIPC", rd, imm)
		r.RG[rd] = r.PC + imm
		r.PC += 4
		return 1
	case i&0b_0000_0000_0000_0000_0000_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0110_1111: // JAL
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0110_0111: // JALR
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0110_0011: // BEQ
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0110_0011: // BNE
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0100_0000_0110_0011: // BLT
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0101_0000_0110_0011: // BGE
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0110_0000_0110_0011: // BLTU
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0111_0000_0110_0011: // BGEU
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0000_0011: // LB
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0000_0011: // LH
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0010_0000_0000_0011: // LW
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0100_0000_0000_0011: // LBU
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0101_0000_0000_0011: // LHU
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0010_0011: // SB
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0010_0011: // SH
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0010_0000_0010_0011: // SW
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0001_0011: // ADDI
		_, rd, _, rs1, imm := IType(i)
		DebuglnIType("ADDI", rd, rs1, SignExtend(imm, 11))
		r.RG[rd] = r.RG[rs1] + SignExtend(imm, 11)
		r.PC += 4
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0010_0000_0001_0011: // SLTI
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0011_0000_0001_0011: // SLTIU
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0100_0000_0001_0011: // XORI
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0110_0000_0001_0011: // ORI
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0111_0000_0001_0011: // ANDI
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0001_0011: // SLLI
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0101_0000_0001_0011: // SRLI
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0100_0000_0000_0000_0101_0000_0001_0011: // SRAI
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0011_0011: // ADD
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0100_0000_0000_0000_0000_0000_0011_0011: // SUB
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0011_0011: // SLL
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0010_0000_0011_0011: // SLT
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0011_0000_0011_0011: // SLTU
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0100_0000_0011_0011: // XOR
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0101_0000_0011_0011: // SRL
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0100_0000_0000_0000_0101_0000_0011_0011: // SRA
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0110_0000_0011_0011: // OR
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0111_0000_0011_0011: // AND
	case i&0b_1111_0000_0000_1111_1111_1111_1111_1111 == 0b_0000_0000_0000_0000_0000_0000_0000_1111: // FENCE
	case i&0b_1111_1111_1111_1111_1111_1111_1111_1111 == 0b_0000_0000_0000_0000_0001_0000_0000_1111: // FENCE.I
	case i&0b_1111_1111_1111_1111_1111_1111_1111_1111 == 0b_0000_0000_0000_0000_0000_0000_0111_0011: // ECALL
	case i&0b_1111_1111_1111_1111_1111_1111_1111_1111 == 0b_0000_0000_0001_0000_0000_0000_0111_0011: // EBREAK
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0111_0011: // CSRRW
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0010_0000_0111_0011: // CSRRS
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0011_0000_0111_0011: // CSRRC
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0101_0000_0111_0011: // CSRRWI
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0110_0000_0111_0011: // CSRRSI
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0111_0000_0111_0011: // CSRRCI
	}
	return 0
}
