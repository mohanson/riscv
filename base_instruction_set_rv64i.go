package riscv

import "log"

type RegisterRV64I struct {
	RG [32]uint64
	PC uint64
}

func ExecuterRV64I(r *RegisterRV64I, m []byte, i uint64) int {
	switch {
	case i&0b_0000_0000_0000_0000_0000_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0011_0111: // LUI
		log.Println("LUI")
	case i&0b_0000_0000_0000_0000_0000_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_00010_111: // AUIPC
		rd, imm := UType(i)
		DebuglnUType("AUIPC", rd, imm)
		r.RG[rd] = r.PC + imm
		r.PC += 4
		return 1
	case i&0b_0000_0000_0000_0000_0000_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0110_1111: // JAL
		rd, imm := JType(i)
		imm = SignExtend(imm, 19)
		DebuglnJType("JAL", rd, imm)
		r.RG[rd] = r.PC + 4
		r.PC += imm
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0110_0111: // JALR
		log.Println("JALR")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0110_0011: // BEQ
		rs1, rs2, imm := BType(i)
		imm = SignExtend(imm, 12)
		DebuglnBType("BEQ", rs1, rs2, imm)
		if r.RG[rs1] == r.RG[rs2] {
			r.PC += imm
		} else {
			r.PC += 4
		}
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0110_0011: // BNE
		rs1, rs2, imm := BType(i)
		imm = SignExtend(imm, 12)
		DebuglnBType("BNE", rs1, rs2, imm)
		if r.RG[rs1] != r.RG[rs2] {
			r.PC += imm
		} else {
			r.PC += 4
		}
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0100_0000_0110_0011: // BLT
		rs1, rs2, imm := BType(i)
		imm = SignExtend(imm, 12)
		DebuglnBType("BLT", rs1, rs2, imm)
		if int64(r.RG[rs1]) < int64(r.RG[rs2]) {
			r.PC += imm
		} else {
			r.PC += 4
		}
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0101_0000_0110_0011: // BGE
		rs1, rs2, imm := BType(i)
		imm = SignExtend(imm, 12)
		DebuglnBType("BGE", rs1, rs2, imm)
		if int64(r.RG[rs1]) >= int64(r.RG[rs2]) {
			r.PC += imm
		} else {
			r.PC += 4
		}
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0110_0000_0110_0011: // BLTU
		rs1, rs2, imm := BType(i)
		imm = SignExtend(imm, 12)
		DebuglnBType("BLTU", rs1, rs2, imm)
		if r.RG[rs1] < r.RG[rs2] {
			r.PC += imm
		} else {
			r.PC += 4
		}
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0111_0000_0110_0011: // BGEU
		rs1, rs2, imm := BType(i)
		imm = SignExtend(imm, 12)
		DebuglnBType("BGEU", rs1, rs2, imm)
		if r.RG[rs1] >= r.RG[rs2] {
			r.PC += imm
		} else {
			r.PC += 4
		}
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0000_0011: // LB
		log.Println("LB")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0000_0011: // LH
		log.Println("LH")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0010_0000_0000_0011: // LW
		log.Println("LW")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0100_0000_0000_0011: // LBU
		log.Println("LBU")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0101_0000_0000_0011: // LHU
		log.Println("LHU")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0010_0011: // SB
		log.Println("SB")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0010_0011: // SH
		log.Println("SH")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0010_0000_0010_0011: // SW
		log.Println("SW")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0001_0011: // ADDI
		rd, rs1, imm := IType(i)
		imm = SignExtend(imm, 11)
		DebuglnIType("ADDI", rd, rs1, imm)
		r.RG[rd] = r.RG[rs1] + imm
		r.PC += 4
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0010_0000_0001_0011: // SLTI
		rd, rs1, imm := IType(i)
		imm = SignExtend(imm, 11)
		DebuglnIType("SLTI", rd, rs1, imm)
		if int64(r.RG[rs1]) < int64(imm) {
			r.RG[rd] = 1
		} else {
			r.RG[rd] = 0
		}
		r.PC += 4
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0011_0000_0001_0011: // SLTIU
		rd, rs1, imm := IType(i)
		DebuglnIType("SLTIU", rd, rs1, imm)
		if r.RG[rs1] < imm {
			r.RG[rd] = 1
		} else {
			r.RG[rd] = 0
		}
		r.PC += 4
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0100_0000_0001_0011: // XORI
		rd, rs1, imm := IType(i)
		imm = SignExtend(imm, 11)
		DebuglnIType("XORI", rd, rs1, imm)
		r.RG[rd] = r.RG[rs1] ^ imm
		r.PC += 4
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0110_0000_0001_0011: // ORI
		rd, rs1, imm := IType(i)
		imm = SignExtend(imm, 11)
		DebuglnIType("ORI", rd, rs1, imm)
		r.RG[rd] = r.RG[rs1] | imm
		r.PC += 4
		return 1
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0111_0000_0001_0011: // ANDI
		rd, rs1, imm := IType(i)
		imm = SignExtend(imm, 11)
		DebuglnIType("ANDI", rd, rs1, imm)
		r.RG[rd] = r.RG[rs1] & imm
		r.PC += 4
		return 1
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0001_0011: // SLLI
		log.Println("SLLI")
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0101_0000_0001_0011: // SRLI
		log.Println("SRLI")
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0100_0000_0000_0000_0101_0000_0001_0011: // SRAI
		log.Println("SRAI")
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0000_0000_0011_0011: // ADD
		rd, rs1, rs2 := RType(i)
		DebuglnRType("ADD", rd, rs1, rs2)
		r.RG[rd] = r.RG[rs1] + r.RG[rs2]
		r.PC += 4
		return 1
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0100_0000_0000_0000_0000_0000_0011_0011: // SUB
		rd, rs1, rs2 := RType(i)
		DebuglnRType("SUB", rd, rs1, rs2)
		r.RG[rd] = r.RG[rs1] - r.RG[rs2]
		r.PC += 4
		return 1
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0011_0011: // SLL
		log.Println("SLL")
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0010_0000_0011_0011: // SLT
		log.Println("SLT")
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0011_0000_0011_0011: // SLTU
		log.Println("SLTU")
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0100_0000_0011_0011: // XOR
		log.Println("XOR")
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0101_0000_0011_0011: // SRL
		log.Println("SRL")
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0100_0000_0000_0000_0101_0000_0011_0011: // SRA
		log.Println("SRA")
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0110_0000_0011_0011: // OR
		log.Println("OR")
	case i&0b_1111_1110_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0111_0000_0011_0011: // AND
		log.Println("AND")
	case i&0b_1111_0000_0000_1111_1111_1111_1111_1111 == 0b_0000_0000_0000_0000_0000_0000_0000_1111: // FENCE
		log.Println("FENCE")
	case i&0b_1111_1111_1111_1111_1111_1111_1111_1111 == 0b_0000_0000_0000_0000_0001_0000_0000_1111: // FENCE.I
		log.Println("FENCE.I")
	case i&0b_1111_1111_1111_1111_1111_1111_1111_1111 == 0b_0000_0000_0000_0000_0000_0000_0111_0011: // ECALL
		log.Println("ECALL")
	case i&0b_1111_1111_1111_1111_1111_1111_1111_1111 == 0b_0000_0000_0001_0000_0000_0000_0111_0011: // EBREAK
		log.Println("EBREAK")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0001_0000_0111_0011: // CSRRW
		log.Println("CSRRW")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0010_0000_0111_0011: // CSRRS
		log.Println("CSRRS")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0011_0000_0111_0011: // CSRRC
		log.Println("CSRRC")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0101_0000_0111_0011: // CSRRWI
		log.Println("CSRRWI")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0110_0000_0111_0011: // CSRRSI
		log.Println("CSRRSI")
	case i&0b_0000_0000_0000_0000_0111_0000_0111_1111 == 0b_0000_0000_0000_0000_0111_0000_0111_0011: // CSRRCI
		log.Println("CSRRCI")
	}
	return 0
}
