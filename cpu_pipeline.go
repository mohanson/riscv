package rv64

import (
	"fmt"
	"math"
	"math/big"
)

// The classic RISC pipeline comprises:
//   1. Instruction fetch
//   2. Instruction decode and register fetch
//   3. Execute
//   4. Memory access
//   5. Register write back
//
// | 1   | 2   | 3   | 4   | 5   | 6   | 7   | 8   | 9   |
// | --- | --- | --- | --- | --- | --- | --- | --- | --- |
// | IF  | ID  | EX  | MEM | WB  |     |     |     |     |
// |     | IF  | ID  | EX  | MEM | WB  |     |     |     |
// |     |     | IF  | ID  | EX  | MEM | WB  |     |     |
// |     |     |     | IF  | ID  | EX  | MEM | WB  |     |
// |     |     |     |     | IF  | ID  | EX  | MEM | WB  |

func (c *CPU) PipelineInstructionFetch() ([]byte, error) {
	a, err := c.GetMemory().GetByte(c.GetPC(), 2)
	if err != nil {
		return nil, err
	}
	b := InstructionLengthEncoding(a)
	r, err := c.GetMemory().GetByte(c.GetPC(), uint64(b))
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *CPU) PipelineExecute(data []byte) (uint64, error) {
	switch len(data) {
	case 2:
		Panicln("Unreachable")
	case 4:
		var s uint64 = 0
		for i := len(data) - 1; i >= 0; i-- {
			s += uint64(data[i]) << (8 * i)
		}
		switch InstructionPart(s, 0, 6) {
		case 0b0110111:
			rd, imm := UType(s)
			Debugln(fmt.Sprintf("%#08x % 10s rd : %s imm: ____(%#016x)", c.GetPC(), "lui", c.LogI(rd), imm))
			c.SetRegister(rd, imm)
			c.SetPC(c.GetPC() + 4)
			return 1, nil
		case 0b0010111:
			rd, imm := UType(s)
			Debugln(fmt.Sprintf("%#08x % 10s rd : %s imm: ____(%#016x)", c.GetPC(), "auipc", c.LogI(rd), imm))
			c.SetRegister(rd, c.GetPC()+imm)
			c.SetPC(c.GetPC() + 4)
			return 1, nil
		case 0b1101111:
			rd, imm := JType(s)
			Debugln(fmt.Sprintf("%#08x % 10s rd : %s imm: ____(%#016x)", c.GetPC(), "jal", c.LogI(rd), imm))
			c.SetRegister(rd, c.GetPC()+4)
			r := c.GetPC() + imm
			if r%4 != 0x00 {
				return 0, ErrMisalignedInstructionFetch
			}
			c.SetPC(r)
			return 1, nil
		case 0b1100111:
			rd, rs1, imm := IType(s)
			Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "jalr", c.LogI(rd), c.LogI(rs1), imm))
			c.SetRegister(rd, c.GetPC()+4)
			r := (c.GetRegister(rs1) + imm) & 0xfffffffffffffffe
			if r%4 != 0x00 {
				return 0, ErrMisalignedInstructionFetch
			}
			c.SetPC(r)
			return 1, nil
		case 0b1100011:
			rs1, rs2, imm := BType(s)
			if imm%2 != 0x00 {
				return 0, ErrMisalignedInstructionFetch
			}
			var cond bool
			switch InstructionPart(s, 12, 14) {
			case 0b000:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "beq", c.LogI(rs1), c.LogI(rs2), imm))
				cond = c.GetRegister(rs1) == c.GetRegister(rs2)
			case 0b001:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "bne", c.LogI(rs1), c.LogI(rs2), imm))
				cond = c.GetRegister(rs1) != c.GetRegister(rs2)
			case 0b100:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "blt", c.LogI(rs1), c.LogI(rs2), imm))
				cond = int64(c.GetRegister(rs1)) < int64(c.GetRegister(rs2))
			case 0b101:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "bge", c.LogI(rs1), c.LogI(rs2), imm))
				cond = int64(c.GetRegister(rs1)) >= int64(c.GetRegister(rs2))
			case 0b110:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "bltu", c.LogI(rs1), c.LogI(rs2), imm))
				cond = c.GetRegister(rs1) < c.GetRegister(rs2)
			case 0b111:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "bgeu", c.LogI(rs1), c.LogI(rs2), imm))
				cond = c.GetRegister(rs1) >= c.GetRegister(rs2)
			}
			if cond {
				c.SetPC(c.GetPC() + imm)
			} else {
				c.SetPC(c.GetPC() + 4)
			}
			return 1, nil
		case 0b0000011:
			rd, rs1, imm := IType(s)
			a := c.GetRegister(rs1) + imm
			var v uint64
			switch InstructionPart(s, 12, 14) {
			case 0b000:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "lb", c.LogI(rd), c.LogI(rs1), imm))
				b, err := c.GetMemory().GetUint8(a)
				if err != nil {
					return 0, err
				}
				v = SignExtend(uint64(b), 7)
			case 0b001:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "lh", c.LogI(rd), c.LogI(rs1), imm))
				b, err := c.GetMemory().GetUint16(a)
				if err != nil {
					return 0, err
				}
				v = SignExtend(uint64(b), 15)
			case 0b010:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "lw", c.LogI(rd), c.LogI(rs1), imm))
				b, err := c.GetMemory().GetUint32(a)
				if err != nil {
					return 0, err
				}
				v = SignExtend(uint64(b), 31)
			case 0b011:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "ld", c.LogI(rd), c.LogI(rs1), imm))
				b, err := c.GetMemory().GetUint64(a)
				if err != nil {
					return 0, err
				}
				v = b
			case 0b100:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "lbu", c.LogI(rd), c.LogI(rs1), imm))
				b, err := c.GetMemory().GetUint8(a)
				if err != nil {
					return 0, err
				}
				v = uint64(b)
			case 0b101:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "lhu", c.LogI(rd), c.LogI(rs1), imm))
				b, err := c.GetMemory().GetUint16(a)
				if err != nil {
					return 0, err
				}
				v = uint64(b)
			case 0b110:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "lwu", c.LogI(rd), c.LogI(rs1), imm))
				b, err := c.GetMemory().GetUint32(a)
				if err != nil {
					return 0, err
				}
				v = uint64(b)
			}
			c.SetRegister(rd, v)
			c.SetPC(c.GetPC() + 4)
			return 1, nil
		case 0b0100011:
			rs1, rs2, imm := SType(s)
			a := c.GetRegister(rs1) + imm
			var err error
			switch InstructionPart(s, 12, 14) {
			case 0b000:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "sb", c.LogI(rs1), c.LogI(rs2), imm))
				err = c.GetMemory().SetUint8(a, uint8(c.GetRegister(rs2)))
			case 0b001:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "sh", c.LogI(rs1), c.LogI(rs2), imm))
				err = c.GetMemory().SetUint16(a, uint16(c.GetRegister(rs2)))
			case 0b010:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "sw", c.LogI(rs1), c.LogI(rs2), imm))
				err = c.GetMemory().SetUint32(a, uint32(c.GetRegister(rs2)))
			case 0b011:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "sd", c.LogI(rs1), c.LogI(rs2), imm))
				err = c.GetMemory().SetUint64(a, c.GetRegister(rs2))
			}
			if err != nil {
				return 0, err
			}
			c.SetPC(c.GetPC() + 4)
			return 1, nil
		case 0b0010011:
			rd, rs1, imm := IType(s)
			switch InstructionPart(s, 12, 14) {
			case 0b000:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "addi", c.LogI(rd), c.LogI(rs1), imm))
				c.SetRegister(rd, c.GetRegister(rs1)+imm)
			case 0b010:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "slti", c.LogI(rd), c.LogI(rs1), imm))
				if int64(c.GetRegister(rs1)) < int64(imm) {
					c.SetRegister(rd, 1)
				} else {
					c.SetRegister(rd, 0)
				}
			case 0b011:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "sltiu", c.LogI(rd), c.LogI(rs1), imm))
				if c.GetRegister(rs1) < imm {
					c.SetRegister(rd, 1)
				} else {
					c.SetRegister(rd, 0)
				}
			case 0b100:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "xori", c.LogI(rd), c.LogI(rs1), imm))
				c.SetRegister(rd, c.GetRegister(rs1)^imm)
			case 0b110:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "ori", c.LogI(rd), c.LogI(rs1), imm))
				c.SetRegister(rd, c.GetRegister(rs1)|imm)
			case 0b111:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "andi", c.LogI(rd), c.LogI(rs1), imm))
				c.SetRegister(rd, c.GetRegister(rs1)&imm)
			case 0b001:
				shamt := InstructionPart(imm, 0, 5)
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "slli", c.LogI(rd), c.LogI(rs1), imm))
				c.SetRegister(rd, c.GetRegister(rs1)<<shamt)
			case 0b101:
				shamt := InstructionPart(imm, 0, 5)
				switch InstructionPart(s, 26, 31) {
				case 0b000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "srli", c.LogI(rd), c.LogI(rs1), imm))
					c.SetRegister(rd, c.GetRegister(rs1)>>shamt)
				case 0b010000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "srai", c.LogI(rd), c.LogI(rs1), imm))
					c.SetRegister(rd, uint64(int64(c.GetRegister(rs1))>>shamt))
				}
			}
			c.SetPC(c.GetPC() + 4)
			return 1, nil
		case 0b0110011:
			rd, rs1, rs2 := RType(s)
			switch InstructionPart(s, 12, 14) {
			case 0b000:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "add", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, c.GetRegister(rs1)+c.GetRegister(rs2))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0000001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "mul", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, uint64(int64(c.GetRegister(rs1))*int64(c.GetRegister(rs2))))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0100000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "sub", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, c.GetRegister(rs1)-c.GetRegister(rs2))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b001:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "sll", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, c.GetRegister(rs1)<<(c.GetRegister(rs2)&0x3f))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0000001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "mulh", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v := func() uint64 {
						ag1 := big.NewInt(int64(c.GetRegister(rs1)))
						ag2 := big.NewInt(int64(c.GetRegister(rs2)))
						tmp := big.NewInt(0)
						tmp.Mul(ag1, ag2)
						tmp.Rsh(tmp, 64)
						return uint64(tmp.Int64())
					}()
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b010:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "slt", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					if int64(c.GetRegister(rs1)) < int64(c.GetRegister(rs2)) {
						c.SetRegister(rd, 1)
					} else {
						c.SetRegister(rd, 0)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0000001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "mulhsu", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v := func() uint64 {
						ag1 := big.NewInt(int64(c.GetRegister(rs1)))
						ag2 := big.NewInt(int64(c.GetRegister(rs2)))
						if ag2.Cmp(big.NewInt(0)) == -1 {
							tmp := big.NewInt(0)
							tmp.Add(big.NewInt(math.MaxInt64), big.NewInt(math.MaxInt64))
							tmp.Add(tmp, big.NewInt(2))
							ag2 = tmp.Add(tmp, ag2)
						}
						tmp := big.NewInt(0)
						tmp.Mul(ag1, ag2)
						tmp.Rsh(tmp, 64)
						return uint64(tmp.Int64())
					}()
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b011:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "sltu", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					if c.GetRegister(rs1) < c.GetRegister(rs2) {
						c.SetRegister(rd, 1)
					} else {
						c.SetRegister(rd, 0)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0000001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "mulhu", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v := func() uint64 {
						ag1 := big.NewInt(int64(c.GetRegister(rs1)))
						ag2 := big.NewInt(int64(c.GetRegister(rs2)))
						if ag1.Cmp(big.NewInt(0)) == -1 {
							tmp := big.NewInt(0)
							tmp.Add(big.NewInt(math.MaxInt64), big.NewInt(math.MaxInt64))
							tmp.Add(tmp, big.NewInt(2))
							ag1 = tmp.Add(tmp, ag1)
						}
						if ag2.Cmp(big.NewInt(0)) == -1 {
							tmp := big.NewInt(0)
							tmp.Add(big.NewInt(math.MaxInt64), big.NewInt(math.MaxInt64))
							tmp.Add(tmp, big.NewInt(2))
							ag2 = tmp.Add(tmp, ag2)
						}
						tmp := big.NewInt(0)
						tmp.Mul(ag1, ag2)
						tmp.Rsh(tmp, 64)
						return tmp.Uint64()
					}()
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b100:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "xor", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, c.GetRegister(rs1)^c.GetRegister(rs2))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0000001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "div", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					if c.GetRegister(rs2) == 0 {
						c.SetRegister(rd, math.MaxUint64)
					} else {
						c.SetRegister(rd, uint64(int64(c.GetRegister(rs1))/int64(c.GetRegister(rs2))))
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b101:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "srl", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, c.GetRegister(rs1)>>(c.GetRegister(rs2)&0x3f))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0000001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "divu", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					if c.GetRegister(rs2) == 0 {
						c.SetRegister(rd, math.MaxUint64)
					} else {
						c.SetRegister(rd, c.GetRegister(rs1)/c.GetRegister(rs2))
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0100000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "sra", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, uint64(int64(c.GetRegister(rs1))>>(c.GetRegister(rs2)&0x3f)))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b110:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "or", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, c.GetRegister(rs1)|c.GetRegister(rs2))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0000001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "rem", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					if c.GetRegister(rs2) == 0 {
						c.SetRegister(rd, c.GetRegister(rs1))
					} else {
						c.SetRegister(rd, uint64(int64(c.GetRegister(rs1))%int64(c.GetRegister(rs2))))
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b111:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "and", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, c.GetRegister(rs1)&c.GetRegister(rs2))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0000001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "remu", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					if c.GetRegister(rs2) == 0 {
						c.SetRegister(rd, c.GetRegister(rs1))
					} else {
						c.SetRegister(rd, c.GetRegister(rs1)%c.GetRegister(rs2))
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			}
		case 0b0001111:
			switch InstructionPart(s, 12, 14) {
			case 0b000:
				Debugln(fmt.Sprintf("%#08x % 10s", c.GetPC(), "fence"))
			case 0b001:
				Debugln(fmt.Sprintf("%#08x % 10s", c.GetPC(), "fence.i"))
			}
			c.SetPC(c.GetPC() + 4)
			return 1, nil
		case 0b1110011:
			rd, rs1, csr := IType(s)
			switch InstructionPart(s, 12, 14) {
			case 0b000:
				switch InstructionPart(s, 20, 31) {
				case 0b000000000000:
					Debugln(fmt.Sprintf("%#08x % 10s", c.GetPC(), "ecall"))
					return c.GetSystem().HandleCall(c)
				case 0b000000000001:
					Debugln(fmt.Sprintf("%#08x % 10s", c.GetPC(), "ebreak"))
					return 1, nil
				}
			case 0b001:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s csr: %#016x", c.GetPC(), "csrrw", c.LogI(rd), c.LogI(rs1), csr))
				if rd != Rzero {
					c.SetRegister(rd, c.GetCSR().Get(csr))
				}
				c.GetCSR().Set(csr, c.GetRegister(rs1))
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b010:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s csr: %#016x", c.GetPC(), "csrrs", c.LogI(rd), c.LogI(rs1), csr))
				c.SetRegister(rd, c.GetCSR().Get(csr))
				if rs1 != Rzero {
					c.GetCSR().Set(csr, c.GetCSR().Get(csr)|c.GetRegister(rs1))
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b011:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s csr: %#016x", c.GetPC(), "csrrc", c.LogI(rd), c.LogI(rs1), csr))
				c.SetRegister(rd, c.GetCSR().Get(csr))
				if rs1 != Rzero {
					c.GetCSR().Set(csr, c.GetCSR().Get(csr)&(math.MaxUint64-c.GetRegister(rs1)))
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b101:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s csr: %#016x", c.GetPC(), "csrrwi", c.LogI(rd), c.LogI(rs1), csr))
				if rd != Rzero {
					c.SetRegister(rd, c.GetCSR().Get(csr))
				}
				c.GetCSR().Set(csr, rs1)
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b110:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s csr: %#016x", c.GetPC(), "csrrsi", c.LogI(rd), c.LogI(rs1), csr))
				c.SetRegister(rd, c.GetCSR().Get(csr))
				if csr != 0x00 {
					c.GetCSR().Set(csr, c.GetCSR().Get(csr)|rs1)
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b111:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s csr: %#016x", c.GetPC(), "csrrci", c.LogI(rd), c.LogI(rs1), csr))
				c.SetRegister(rd, c.GetCSR().Get(csr))
				if csr != 0x00 {
					c.GetCSR().Set(csr, c.GetCSR().Get(csr)&(math.MaxUint64-rs1))
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			}
		case 0b0011011:
			rd, rs1, imm := IType(s)
			switch InstructionPart(s, 12, 14) {
			case 0b000:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "addiw", c.LogI(rd), c.LogI(rs1), imm))
				c.SetRegister(rd, uint64(int32(c.GetRegister(rs1))+int32(imm)))
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b001:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "slliw", c.LogI(rd), c.LogI(rs1), imm))
				if InstructionPart(imm, 5, 5) != 0x00 {
					return 0, ErrAbnormalInstruction
				}
				c.SetRegister(rd, SignExtend(uint64(uint32(c.GetRegister(rs1))<<imm), 31))
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b101:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "srliw", c.LogI(rd), c.LogI(rs1), imm))
					if InstructionPart(imm, 5, 5) != 0x00 {
						return 0, ErrAbnormalInstruction
					}
					shamt := InstructionPart(imm, 0, 4)
					c.SetRegister(rd, SignExtend(uint64(uint32(c.GetRegister(rs1))>>shamt), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0100000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "sraiw", c.LogI(rd), c.LogI(rs1), imm))
					if InstructionPart(imm, 5, 5) != 0x00 {
						return 0, ErrAbnormalInstruction
					}
					shamt := InstructionPart(imm, 0, 4)
					c.SetRegister(rd, uint64(int32(c.GetRegister(rs1))>>shamt))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			}
		case 0b0111011:
			rd, rs1, rs2 := RType(s)
			switch InstructionPart(s, 12, 14) {
			case 0b000:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "addw", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, uint64(int32(c.GetRegister(rs1))+int32(c.GetRegister(rs2))))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0000001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "mulw", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, uint64(int32(c.GetRegister(rs1))*int32(c.GetRegister(rs2))))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0100000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "subw", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, uint64(int32(c.GetRegister(rs1))-int32(c.GetRegister(rs2))))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b001:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "sllw", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
				s := c.GetRegister(rs2) & 0x1f
				c.SetRegister(rd, SignExtend(uint64(uint32(c.GetRegister(rs1))<<s), 31))
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b100:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "divw", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
				if c.GetRegister(rs2) == 0 {
					c.SetRegister(rd, math.MaxUint64)
				} else {
					c.SetRegister(rd, SignExtend(uint64(int32(c.GetRegister(rs1))/int32(c.GetRegister(rs2))), 31))
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b101:
				switch InstructionPart(s, 25, 31) {
				case 0b0000000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "srlw", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					s := c.GetRegister(rs2) & 0x1f
					c.SetRegister(rd, SignExtend(uint64(uint32(c.GetRegister(rs1))>>s), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0000001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "divuw", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					if c.GetRegister(rs2) == 0 {
						c.SetRegister(rd, math.MaxUint64)
					} else {
						c.SetRegister(rd, SignExtend(uint64(uint32(c.GetRegister(rs1))/uint32(c.GetRegister(rs2))), 31))
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b0100000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "sraw", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					c.SetRegister(rd, uint64(int32(c.GetRegister(rs1))>>InstructionPart(c.GetRegister(rs2), 0, 4)))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b110:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "remw", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
				if c.GetRegister(rs2) == 0 {
					c.SetRegister(rd, c.GetRegister(rs1))
				} else {
					c.SetRegister(rd, uint64(int32(c.GetRegister(rs1))%int32(c.GetRegister(rs2))))
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b111:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "remuw", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
				if c.GetRegister(rs2) == 0 {
					c.SetRegister(rd, c.GetRegister(rs1))
				} else {
					c.SetRegister(rd, SignExtend(uint64(uint32(c.GetRegister(rs1))%uint32(c.GetRegister(rs2))), 31))
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			}
		case 0b0101111:
			rd, rs1, rs2 := RType(s)
			switch InstructionPart(s, 12, 14) {
			case 0b010:
				a := SignExtend(c.GetRegister(rs1), 31)
				switch InstructionPart(s, 27, 31) {
				case 0b00010:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "lr.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint32(a)
					if err != nil {
						return 0, err
					}
					c.SetRegister(rd, SignExtend(uint64(v), 31))
					c.SetLoadReservation(a)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00011:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "sc.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					if a == c.GetLoadReservation() {
						c.GetMemory().SetUint32(a, uint32(c.GetRegister(rs2)))
						c.SetRegister(rd, 0)
					} else {
						c.SetRegister(rd, 1)
					}
					c.SetLoadReservation(0)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amoswap.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint32(a)
					if err != nil {
						return 0, err
					}
					c.GetMemory().SetUint32(a, uint32(c.GetRegister(rs2)))
					c.SetRegister(rd, SignExtend(uint64(v), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amoadd.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint32(a)
					if err != nil {
						return 0, err
					}
					c.GetMemory().SetUint32(a, v+uint32(c.GetRegister(rs2)))
					c.SetRegister(rd, SignExtend(uint64(v), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00100:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amoxor.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint32(a)
					if err != nil {
						return 0, err
					}
					c.GetMemory().SetUint32(a, v^uint32(c.GetRegister(rs2)))
					c.SetRegister(rd, SignExtend(uint64(v), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b01100:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amoand.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint32(a)
					if err != nil {
						return 0, err
					}
					c.GetMemory().SetUint32(a, v&uint32(c.GetRegister(rs2)))
					c.SetRegister(rd, SignExtend(uint64(v), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b01000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amoor.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint32(a)
					if err != nil {
						return 0, err
					}
					c.GetMemory().SetUint32(a, v|uint32(c.GetRegister(rs2)))
					c.SetRegister(rd, SignExtend(uint64(v), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b10000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amomin.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint32(a)
					if err != nil {
						return 0, err
					}
					var r uint32
					if int32(v) < int32(uint32(c.GetRegister(rs2))) {
						r = v
					} else {
						r = uint32(c.GetRegister(rs2))
					}
					c.GetMemory().SetUint32(a, r)
					c.SetRegister(rd, SignExtend(uint64(v), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b10100:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amomax.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint32(a)
					if err != nil {
						return 0, err
					}
					var r uint32
					if int32(v) > int32(uint32(c.GetRegister(rs2))) {
						r = v
					} else {
						r = uint32(c.GetRegister(rs2))
					}
					c.GetMemory().SetUint32(a, r)
					c.SetRegister(rd, SignExtend(uint64(v), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b11000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amominu.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint32(a)
					if err != nil {
						return 0, err
					}
					var r uint32
					if v < uint32(c.GetRegister(rs2)) {
						r = v
					} else {
						r = uint32(c.GetRegister(rs2))
					}
					c.GetMemory().SetUint32(a, r)
					c.SetRegister(rd, SignExtend(uint64(v), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b11100:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amomaxu.w", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint32(a)
					if err != nil {
						return 0, err
					}
					var r uint32
					if v > uint32(c.GetRegister(rs2)) {
						r = v
					} else {
						r = uint32(c.GetRegister(rs2))
					}
					c.GetMemory().SetUint32(a, r)
					c.SetRegister(rd, SignExtend(uint64(v), 31))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b011:
				a := c.GetRegister(rs1)
				switch InstructionPart(s, 27, 31) {
				case 0b00010:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "lr.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint64(a)
					if err != nil {
						return 0, err
					}
					c.SetRegister(rd, v)
					c.SetLoadReservation(a)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00011:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "sc.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					if a == c.GetLoadReservation() {
						c.GetMemory().SetUint64(a, c.GetRegister(rs2))
						c.SetRegister(rd, 0)
					} else {
						c.SetRegister(rd, 1)
					}
					c.SetLoadReservation(0)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amoswap.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint64(a)
					if err != nil {
						return 0, err
					}
					c.GetMemory().SetUint64(a, c.GetRegister(rs2))
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amoadd.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint64(a)
					if err != nil {
						return 0, err
					}
					c.GetMemory().SetUint64(a, v+c.GetRegister(rs2))
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00100:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amoxor.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint64(a)
					if err != nil {
						return 0, err
					}
					c.GetMemory().SetUint64(a, v^c.GetRegister(rs2))
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b01100:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amoand.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint64(a)
					if err != nil {
						return 0, err
					}
					c.GetMemory().SetUint64(a, v&c.GetRegister(rs2))
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b01000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amoor.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					a := c.GetRegister(rs1)
					v, err := c.GetMemory().GetUint64(a)
					if err != nil {
						return 0, err
					}
					c.GetMemory().SetUint64(a, v|c.GetRegister(rs2))
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b10000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amomin.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint64(a)
					if err != nil {
						return 0, err
					}
					var r uint64 = 0
					if int64(v) < int64(c.GetRegister(rs2)) {
						r = v
					} else {
						r = c.GetRegister(rs2)
					}
					c.GetMemory().SetUint64(a, r)
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b10100:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amomax.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint64(a)
					if err != nil {
						return 0, err
					}
					var r uint64 = 0
					if int64(v) > int64(c.GetRegister(rs2)) {
						r = v
					} else {
						r = c.GetRegister(rs2)
					}
					c.GetMemory().SetUint64(a, r)
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b11000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amominu.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint64(a)
					if err != nil {
						return 0, err
					}
					var r uint64 = 0
					if v < c.GetRegister(rs2) {
						r = v
					} else {
						r = c.GetRegister(rs2)
					}
					c.GetMemory().SetUint64(a, r)
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b11100:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "amomaxu.d", c.LogI(rd), c.LogI(rs1), c.LogI(rs2)))
					v, err := c.GetMemory().GetUint64(a)
					if err != nil {
						return 0, err
					}
					var r uint64 = 0
					if v > c.GetRegister(rs2) {
						r = v
					} else {
						r = c.GetRegister(rs2)
					}
					c.GetMemory().SetUint64(a, r)
					c.SetRegister(rd, v)
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			}
		case 0b0000111:
			rd, rs1, imm := IType(s)
			a := c.GetRegister(rs1) + imm
			switch InstructionPart(s, 12, 14) {
			case 0b010:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "flw", c.LogF(rd), c.LogI(rs1), imm))
				v, err := c.GetMemory().GetUint32(a)
				if err != nil {
					return 0, err
				}
				c.SetRegisterFloatAsFloat32(rd, math.Float32frombits(v))
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b011:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s imm: ____(%#016x)", c.GetPC(), "fld", c.LogF(rd), c.LogI(rs1), imm))
				v, err := c.GetMemory().GetUint64(a)
				if err != nil {
					return 0, err
				}
				c.SetRegisterFloat(rd, v)
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			}
		case 0b0100111:
			rs1, rs2, imm := SType(s)
			a := c.GetRegister(rs1) + imm
			switch InstructionPart(s, 12, 14) {
			case 0b010:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "fsw", c.LogI(rs1), c.LogF(rs2), imm))
				err := c.GetMemory().SetUint32(a, uint32(c.GetRegisterFloat(rs2)))
				if err != nil {
					return 0, err
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b011:
				Debugln(fmt.Sprintf("%#08x % 10s rs1: %s rs2: %s imm: ____(%#016x)", c.GetPC(), "fsd", c.LogI(rs1), c.LogF(rs2), imm))
				err := c.GetMemory().SetUint64(a, c.GetRegisterFloat(rs2))
				if err != nil {
					return 0, err
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			}
		case 0b1000011:
			rd, rs1, rs2, rs3 := R4Type(s)
			switch InstructionPart(s, 25, 26) {
			case 0b00:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s rs3: %s", c.GetPC(), "fmadd.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2), c.LogF(rs3)))
				c.ClrFloatFlag()
				a := c.GetRegisterFloatAsFloat32(rs1)
				b := c.GetRegisterFloatAsFloat32(rs2)
				d := c.GetRegisterFloatAsFloat32(rs3)
				r := a*b + d
				c.SetRegisterFloatAsFloat32(rd, r)
				if r-d != a*b || r-a*b != d || (r-d)/a != b || (r-d)/b != a {
					c.SetFloatFlag(FFlagsNX, 1)
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b01:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s rs3: %s", c.GetPC(), "fmadd.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2), c.LogF(rs3)))
				c.ClrFloatFlag()
				a := c.GetRegisterFloatAsFloat64(rs1)
				b := c.GetRegisterFloatAsFloat64(rs2)
				d := c.GetRegisterFloatAsFloat64(rs3)
				r := a*b + d
				c.SetRegisterFloatAsFloat64(rd, r)
				if r-d != a*b || r-a*b != d || (r-d)/a != b || (r-d)/b != a {
					c.SetFloatFlag(FFlagsNX, 1)
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			}
		case 0b1000111:
			rd, rs1, rs2, rs3 := R4Type(s)
			switch InstructionPart(s, 25, 26) {
			case 0b00:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s rs3: %s", c.GetPC(), "fmsub.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2), c.LogF(rs3)))
				c.ClrFloatFlag()
				a := c.GetRegisterFloatAsFloat32(rs1)
				b := c.GetRegisterFloatAsFloat32(rs2)
				d := c.GetRegisterFloatAsFloat32(rs3)
				r := a*b - d
				c.SetRegisterFloatAsFloat32(rd, r)
				if r+d != a*b || a*b-r != d || (r+d)/a != b || (r+d)/b != a {
					c.SetFloatFlag(FFlagsNX, 1)
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b01:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s rs3: %s", c.GetPC(), "fmsub.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2), c.LogF(rs3)))
				c.ClrFloatFlag()
				a := c.GetRegisterFloatAsFloat64(rs1)
				b := c.GetRegisterFloatAsFloat64(rs2)
				d := c.GetRegisterFloatAsFloat64(rs3)
				r := a*b - d
				c.SetRegisterFloatAsFloat64(rd, r)
				if r+d != a*b || a*b-r != d || (r+d)/a != b || (r+d)/b != a {
					c.SetFloatFlag(FFlagsNX, 1)
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			}
		case 0b1001011:
			rd, rs1, rs2, rs3 := R4Type(s)
			switch InstructionPart(s, 25, 26) {
			case 0b00:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s rs3: %s", c.GetPC(), "fnmsub.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2), c.LogF(rs3)))
				c.ClrFloatFlag()
				a := c.GetRegisterFloatAsFloat32(rs1)
				b := c.GetRegisterFloatAsFloat32(rs2)
				d := c.GetRegisterFloatAsFloat32(rs3)
				r := a*b - d
				c.SetRegisterFloatAsFloat32(rd, -r)
				if r+d != a*b || a*b-r != d || (r+d)/a != b || (r+d)/b != a {
					c.SetFloatFlag(FFlagsNX, 1)
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b01:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s rs3: %s", c.GetPC(), "fnmsub.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2), c.LogF(rs3)))
				c.ClrFloatFlag()
				a := c.GetRegisterFloatAsFloat64(rs1)
				b := c.GetRegisterFloatAsFloat64(rs2)
				d := c.GetRegisterFloatAsFloat64(rs3)
				r := a*b - d
				c.SetRegisterFloatAsFloat64(rd, -r)
				if r+d != a*b || a*b-r != d || (r+d)/a != b || (r+d)/b != a {
					c.SetFloatFlag(FFlagsNX, 1)
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			}
		case 0b1001111:
			rd, rs1, rs2, rs3 := R4Type(s)
			switch InstructionPart(s, 25, 26) {
			case 0b00:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s rs3: %s", c.GetPC(), "fnmadd.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2), c.LogF(rs3)))
				c.ClrFloatFlag()
				a := c.GetRegisterFloatAsFloat32(rs1)
				b := c.GetRegisterFloatAsFloat32(rs2)
				d := c.GetRegisterFloatAsFloat32(rs3)
				r := a*b + d
				c.SetRegisterFloatAsFloat32(rd, -r)
				if r-d != a*b || r-a*b != d || (r-d)/a != b || (r-d)/b != a {
					c.SetFloatFlag(FFlagsNX, 1)
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			case 0b01:
				Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s rs3: %s", c.GetPC(), "fnmadd.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2), c.LogF(rs3)))
				c.ClrFloatFlag()
				a := c.GetRegisterFloatAsFloat64(rs1)
				b := c.GetRegisterFloatAsFloat64(rs2)
				d := c.GetRegisterFloatAsFloat64(rs3)
				r := a*b + d
				c.SetRegisterFloatAsFloat64(rd, -r)
				if r-d != a*b || r-a*b != d || (r-d)/a != b || (r-d)/b != a {
					c.SetFloatFlag(FFlagsNX, 1)
				}
				c.SetPC(c.GetPC() + 4)
				return 1, nil
			}
		case 0b1010011:
			rd, rs1, rs2 := RType(s)
			switch InstructionPart(s, 25, 26) {
			case 0b00:
				a := c.GetRegisterFloatAsFloat32(rs1)
				b := c.GetRegisterFloatAsFloat32(rs2)
				switch InstructionPart(s, 27, 31) {
				case 0b00000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fadd.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.ClrFloatFlag()
					d := a + b
					c.SetRegisterFloatAsFloat32(rd, d)
					if d-a != b || d-b != a {
						c.SetFloatFlag(FFlagsNX, 1)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fsub.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.ClrFloatFlag()
					if (math.Signbit(float64(a)) == math.Signbit(float64(b))) && math.IsInf(float64(a), 0) && math.IsInf(float64(b), 0) {
						c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(NaN32))
						c.SetFloatFlag(FFlagsNV, 1)
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					d := a - b
					c.SetRegisterFloatAsFloat32(rd, d)
					if a-d != b || b+d != a {
						c.SetFloatFlag(FFlagsNX, 1)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00010:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fmul.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.ClrFloatFlag()
					d := a * b
					c.SetRegisterFloatAsFloat32(rd, d)
					if d/a != b || d/b != a || float64(a)*float64(b) != float64(d) {
						c.SetFloatFlag(FFlagsNX, 1)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00011:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fdiv.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.ClrFloatFlag()
					if b == 0 {
						c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(NaN32))
						c.SetFloatFlag(FFlagsDZ, 1)
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					d := a / b
					c.SetRegisterFloatAsFloat32(rd, d)
					if a/d != b || b*d != a || float64(b)*float64(d) != float64(a) {
						c.SetFloatFlag(FFlagsNX, 1)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b01011:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fsqrt.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.ClrFloatFlag()
					if a < 0 {
						c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(NaN32))
						c.SetFloatFlag(FFlagsNV, 1)
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					d := float32(math.Sqrt(float64(a)))
					c.SetRegisterFloatAsFloat32(rd, d)
					if a/d != d || d*d != a || float64(d)*float64(d) != float64(a) {
						c.SetFloatFlag(FFlagsNX, 1)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00100:
					switch InstructionPart(s, 12, 14) {
					case 0b000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fsgnj.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if math.Signbit(float64(b)) {
							c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(math.Float32bits(a)|0x80000000))
						} else {
							c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(math.Float32bits(a)&0x7fffffff))
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fsgnjn.s.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if math.Signbit(float64(b)) {
							c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(math.Float32bits(a)&0x7fffffff))
						} else {
							c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(math.Float32bits(a)|0x80000000))
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b010:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fsgnjx.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if math.Signbit(float64(a)) != math.Signbit(float64(b)) {
							c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(math.Float32bits(a)|0x80000000))
						} else {
							c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(math.Float32bits(a)&0x7fffffff))
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
				case 0b00101:
					switch InstructionPart(s, 12, 14) {
					case 0b000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fmin.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					case 0b001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fmax.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					}
					c.ClrFloatFlag()
					if math.IsNaN(float64(a)) && math.IsNaN(float64(b)) {
						c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(NaN32))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					if math.IsNaN(float64(a)) {
						c.SetRegisterFloatAsFloat32(rd, b)
						if IsSNaN32(a) {
							c.SetFloatFlag(FFlagsNV, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					if math.IsNaN(float64(b)) {
						c.SetRegisterFloatAsFloat32(rd, a)
						if IsSNaN32(b) {
							c.SetFloatFlag(FFlagsNV, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					switch InstructionPart(s, 12, 14) {
					case 0b000:
						if (math.Signbit(float64(a)) && !math.Signbit(float64(b))) || a < b {
							c.SetRegisterFloatAsFloat32(rd, a)
						} else {
							c.SetRegisterFloatAsFloat32(rd, b)
						}
					case 0b001:
						if (!math.Signbit(float64(a)) && math.Signbit(float64(b))) || a > b {
							c.SetRegisterFloatAsFloat32(rd, a)
						} else {
							c.SetRegisterFloatAsFloat32(rd, b)
						}
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b11000:
					switch InstructionPart(s, 20, 24) {
					case 0b00000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.w.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						d := c.GetRegisterFloatAsFloat32(rs1)
						if math.IsNaN(float64(d)) {
							c.SetRegister(rd, 0x7fffffff)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d > float32(math.MaxInt32) {
							c.SetRegister(rd, SignExtend(0x7fffffff, 31))
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d < float32(math.MinInt32) {
							c.SetRegister(rd, SignExtend(0x80000000, 31))
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						c.SetRegister(rd, SignExtend(uint64(int32(d)), 31))
						if math.Ceil(float64(d)) != float64(d) {
							c.SetFloatFlag(FFlagsNX, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.wu.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						d := c.GetRegisterFloatAsFloat32(rs1)
						if math.IsNaN(float64(d)) {
							c.SetRegister(rd, 0xffffffffffffffff)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d > float32(math.MaxUint32) {
							c.SetRegister(rd, SignExtend(0xffffffff, 31))
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d <= float32(-1) {
							c.SetRegister(rd, SignExtend(0x00000000, 31))
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						c.SetRegister(rd, SignExtend(uint64(uint32(d)), 31))
						if math.Ceil(float64(d)) != float64(d) {
							c.SetFloatFlag(FFlagsNX, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00010:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.l.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						d := c.GetRegisterFloatAsFloat32(rs1)
						if math.IsNaN(float64(d)) {
							c.SetRegister(rd, 0x7fffffffffffffff)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d > float32(math.MaxInt64) {
							c.SetRegister(rd, 0x7fffffffffffffff)
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d < float32(math.MinInt64) {
							c.SetRegister(rd, 0x8000000000000000)
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						c.SetRegister(rd, uint64(int64(d)))
						if math.Ceil(float64(d)) != float64(d) {
							c.SetFloatFlag(FFlagsNX, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00011:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.lu.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						d := c.GetRegisterFloatAsFloat32(rs1)
						if math.IsNaN(float64(d)) {
							c.SetRegister(rd, 0xffffffffffffffff)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d > float32(math.MaxUint64) {
							c.SetRegister(rd, 0xffffffffffffffff)
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d <= float32(-1) {
							c.SetRegister(rd, 0x0000000000000000)
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						c.SetRegister(rd, uint64(d))
						if math.Ceil(float64(d)) != float64(d) {
							c.SetFloatFlag(FFlagsNX, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
				case 0b01000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.s.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					d := c.GetRegisterFloatAsFloat64(rs1)
					if math.IsNaN(d) {
						c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(NaN32))
					} else {
						c.SetRegisterFloatAsFloat32(rd, float32(d))
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b11100:
					switch InstructionPart(s, 12, 14) {
					case 0b000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fmv.x.w", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						c.SetRegister(rd, SignExtend(uint64(uint32(c.GetRegisterFloat(rs1))), 31))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fclass.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						a := c.GetRegisterFloatAsFloat32(rs1)
						c.SetRegister(rd, FClassS(a))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
				case 0b10100:
					var cond bool
					switch InstructionPart(s, 12, 14) {
					case 0b010:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "feq.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if IsSNaN32(a) || IsSNaN32(b) {
							c.SetFloatFlag(FFlagsNV, 1)
						} else {
							cond = a == b
						}
					case 0b001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "flt.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if math.IsNaN(float64(a)) || math.IsNaN(float64(b)) {
							c.SetFloatFlag(FFlagsNV, 1)
						} else {
							cond = a < b
						}
					case 0b000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fle.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if math.IsNaN(float64(a)) || math.IsNaN(float64(b)) {
							c.SetFloatFlag(FFlagsNV, 1)
						} else {
							cond = a <= b
						}
					}
					if cond {
						c.SetRegister(rd, 1)
					} else {
						c.SetRegister(rd, 0)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b11010:
					switch InstructionPart(s, 20, 24) {
					case 0b00000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.s.w", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						c.SetRegisterFloatAsFloat32(rd, float32(int32(c.GetRegister(rs1))))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.s.wu", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						c.SetRegisterFloatAsFloat32(rd, float32(uint32(c.GetRegister(rs1))))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00010:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.s.l", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						c.SetRegisterFloatAsFloat32(rd, float32(int64(c.GetRegister(rs1))))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00011:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.s.lu", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						c.SetRegisterFloatAsFloat32(rd, float32(uint64(c.GetRegister(rs1))))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
				case 0b11110:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fmv.w.x", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.SetRegisterFloat(rd, 0xffffffff00000000|uint64(uint32(c.GetRegister(rs1))))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			case 0b01:
				a := c.GetRegisterFloatAsFloat64(rs1)
				b := c.GetRegisterFloatAsFloat64(rs2)
				switch InstructionPart(s, 27, 31) {
				case 0b00000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fadd.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.ClrFloatFlag()
					c.SetRegisterFloatAsFloat64(rd, a+b)
					if big.NewFloat(0).Add(big.NewFloat(a), big.NewFloat(b)).Acc() != big.Exact {
						c.SetFloatFlag(FFlagsNX, 1)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00001:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fsub.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.ClrFloatFlag()
					if (math.Signbit(a) == math.Signbit(b)) && math.IsInf(a, 0) && math.IsInf(b, 0) {
						c.SetRegisterFloat(rd, NaN64)
						c.SetFloatFlag(FFlagsNV, 1)
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					c.SetRegisterFloatAsFloat64(rd, a-b)
					if big.NewFloat(0).Sub(big.NewFloat(a), big.NewFloat(b)).Acc() != big.Exact {
						c.SetFloatFlag(FFlagsNX, 1)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00010:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fmul.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.ClrFloatFlag()
					c.SetRegisterFloatAsFloat64(rd, a*b)
					if big.NewFloat(0).Add(big.NewFloat(a), big.NewFloat(b)).Acc() != big.Exact {
						c.SetFloatFlag(FFlagsNX, 1)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00011:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fdiv.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.ClrFloatFlag()
					if b == 0 {
						c.SetRegisterFloat(rd, NaN64)
						c.SetFloatFlag(FFlagsDZ, 1)
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					c.SetRegisterFloatAsFloat64(rd, a/b)
					if big.NewFloat(0).Quo(big.NewFloat(a), big.NewFloat(b)).Acc() != big.Exact {
						c.SetFloatFlag(FFlagsNX, 1)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b01011:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fsqrt.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.ClrFloatFlag()
					if a < 0 {
						c.SetRegisterFloat(rd, NaN64)
						c.SetFloatFlag(FFlagsNV, 1)
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					c.SetRegisterFloatAsFloat64(rd, math.Sqrt(a))
					d := big.NewFloat(0).Sqrt(big.NewFloat(a))
					if big.NewFloat(0).Mul(d, d).Acc() != big.Exact {
						c.SetFloatFlag(FFlagsNX, 1)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b00100:
					switch InstructionPart(s, 12, 14) {
					case 0b000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fsgnj.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if math.Signbit(b) {
							c.SetRegisterFloat(rd, math.Float64bits(a)|0x8000000000000000)
						} else {
							c.SetRegisterFloat(rd, math.Float64bits(a)&0x7fffffffffffffff)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fsgnjn.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if math.Signbit(b) {
							c.SetRegisterFloat(rd, math.Float64bits(a)&0x7fffffffffffffff)
						} else {
							c.SetRegisterFloat(rd, math.Float64bits(a)|0x8000000000000000)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b010:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fsgnjx.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if math.Signbit(a) != math.Signbit(b) {
							c.SetRegisterFloat(rd, math.Float64bits(a)|0x8000000000000000)
						} else {
							c.SetRegisterFloat(rd, math.Float64bits(a)&0x7fffffffffffffff)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
				case 0b00101:
					switch InstructionPart(s, 12, 14) {
					case 0b000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fmin.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					case 0b001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fmax.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					}
					c.ClrFloatFlag()
					if math.IsNaN(a) && math.IsNaN(b) {
						c.SetRegisterFloat(rd, NaN64)
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					if math.IsNaN(a) {
						c.SetRegisterFloatAsFloat64(rd, b)
						if IsSNaN64(a) {
							c.SetFloatFlag(FFlagsNV, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					if math.IsNaN(b) {
						c.SetRegisterFloatAsFloat64(rd, a)
						if IsSNaN64(b) {
							c.SetFloatFlag(FFlagsNV, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
					switch InstructionPart(s, 12, 14) {
					case 0b000:
						if (math.Signbit(a) && !math.Signbit(b)) || a < b {
							c.SetRegisterFloatAsFloat64(rd, a)
						} else {
							c.SetRegisterFloatAsFloat64(rd, b)
						}
					case 0b001:
						if (!math.Signbit(a) && math.Signbit(b)) || a > b {
							c.SetRegisterFloatAsFloat64(rd, a)
						} else {
							c.SetRegisterFloatAsFloat64(rd, b)
						}
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b11000:
					switch InstructionPart(s, 20, 24) {
					case 0b00000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.w.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						d := c.GetRegisterFloatAsFloat64(rs1)
						if math.IsNaN(d) {
							c.SetRegister(rd, 0x7fffffff)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d > float64(math.MaxInt32) {
							c.SetRegister(rd, SignExtend(0x7fffffff, 31))
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d < float64(math.MinInt32) {
							c.SetRegister(rd, SignExtend(0x80000000, 31))
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						c.SetRegister(rd, SignExtend(uint64(int32(d)), 31))
						if math.Ceil(d) != d {
							c.SetFloatFlag(FFlagsNX, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.wu.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						d := c.GetRegisterFloatAsFloat64(rs1)
						if math.IsNaN(d) {
							c.SetRegister(rd, 0xffffffffffffffff)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d > float64(math.MaxUint32) {
							c.SetRegister(rd, SignExtend(0xffffffff, 31))
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d <= float64(-1) {
							c.SetRegister(rd, SignExtend(0x00000000, 31))
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						c.SetRegister(rd, SignExtend(uint64(uint32(d)), 31))
						if math.Ceil(d) != d {
							c.SetFloatFlag(FFlagsNX, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00010:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.l.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						d := c.GetRegisterFloatAsFloat64(rs1)
						if math.IsNaN(d) {
							c.SetRegister(rd, 0x7fffffffffffffff)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d > float64(math.MaxInt64) {
							c.SetRegister(rd, 0x7fffffffffffffff)
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d < float64(math.MinInt64) {
							c.SetRegister(rd, 0x8000000000000000)
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						c.SetRegister(rd, uint64(int64(d)))
						if math.Ceil(d) != d {
							c.SetFloatFlag(FFlagsNX, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00011:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.lu.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						d := c.GetRegisterFloatAsFloat64(rs1)
						if math.IsNaN(d) {
							c.SetRegister(rd, 0xffffffffffffffff)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d > float64(math.MaxUint64) {
							c.SetRegister(rd, 0xffffffffffffffff)
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						if d <= float64(-1) {
							c.SetRegister(rd, 0x0000000000000000)
							c.SetFloatFlag(FFlagsNV, 1)
							c.SetPC(c.GetPC() + 4)
							return 1, nil
						}
						c.SetRegister(rd, uint64(d))
						if math.Ceil(d) != d {
							c.SetFloatFlag(FFlagsNX, 1)
						}
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
				case 0b01000:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.d.s", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					d := c.GetRegisterFloatAsFloat32(rs1)
					if math.IsNaN(float64(d)) {
						c.SetRegisterFloat(rd, NaN64)
					} else {
						c.SetRegisterFloatAsFloat64(rd, float64(d))
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b10100:
					var cond bool
					switch InstructionPart(s, 12, 14) {
					case 0b010:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "feq.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if IsSNaN64(a) || IsSNaN64(b) {
							c.SetFloatFlag(FFlagsNV, 1)
						} else {
							cond = a == b
						}
					case 0b001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "flt.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if math.IsNaN(a) || math.IsNaN(b) {
							c.SetFloatFlag(FFlagsNV, 1)
						} else {
							cond = a < b
						}
					case 0b000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fle.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						if math.IsNaN(a) || math.IsNaN(b) {
							c.SetFloatFlag(FFlagsNV, 1)
						} else {
							cond = a <= b
						}
					}
					if cond {
						c.SetRegister(rd, 1)
					} else {
						c.SetRegister(rd, 0)
					}
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				case 0b11100:
					switch InstructionPart(s, 12, 14) {
					case 0b000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fmv.x.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						c.SetRegister(rd, c.GetRegisterFloat(rs1))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fclass.d", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						a := c.GetRegisterFloatAsFloat64(rs1)
						c.SetRegister(rd, FClassD(a))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
				case 0b11010:
					switch InstructionPart(s, 20, 24) {
					case 0b00000:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.d.w", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						c.SetRegisterFloatAsFloat64(rd, float64(int32(c.GetRegister(rs1))))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00001:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.d.wu", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						c.SetRegisterFloatAsFloat64(rd, float64(uint32(c.GetRegister(rs1))))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00010:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.d.l", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						c.SetRegisterFloatAsFloat64(rd, float64(int64(c.GetRegister(rs1))))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					case 0b00011:
						Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fcvt.d.lu", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
						c.SetRegisterFloatAsFloat64(rd, float64(uint64(c.GetRegister(rs1))))
						c.SetPC(c.GetPC() + 4)
						return 1, nil
					}
				case 0b11110:
					Debugln(fmt.Sprintf("%#08x % 10s rd : %s rs1: %s rs2: %s", c.GetPC(), "fmv.d.x", c.LogF(rd), c.LogF(rs1), c.LogF(rs2)))
					c.SetRegisterFloat(rd, c.GetRegister(rs1))
					c.SetPC(c.GetPC() + 4)
					return 1, nil
				}
			}
		}
	}
	return 0, nil
}

func IsQNaN32(f float32) bool {
	return math.IsNaN(float64(f)) && math.Float32bits(f)&0x00400000 != 0x00
}

func IsSNaN32(f float32) bool {
	return math.IsNaN(float64(f)) && math.Float32bits(f)&0x00400000 == 0x00
}

func IsSubmoduleFloat32(f float32) bool {
	b := math.Float32bits(f)
	return b&0x7f800000 == 0 && b&0x000fffff != 0
}

func IsQNaN64(f float64) bool {
	return math.IsNaN(f) && math.Float64bits(f)&0x0008000000000000 != 0x00
}

func IsSNaN64(f float64) bool {
	return math.IsNaN(f) && math.Float64bits(f)&0x0008000000000000 == 0x00
}

func IsSubmoduleFloat64(f float64) bool {
	b := math.Float64bits(f)
	return b&0x7ff0000000000000 == 0 && b&0x000fffffffffffff != 0
}

func FClassS(f float32) uint64 {
	s := math.Float32bits(f)&(1<<31) != 0
	if IsSNaN32(f) {
		return 0b01_00000000
	}
	if IsQNaN32(f) {
		return 0b10_00000000
	}
	if s {
		if f < -math.MaxFloat32 {
			return 0b00_00000001
		} else if f == 0 {
			return 0b00_00001000
		} else if IsSubmoduleFloat32(f) {
			return 0b00_00000100
		} else {
			return 0b00_00000010
		}
	}
	if f > math.MaxFloat32 {
		return 0b00_10000000
	} else if f == 0 {
		return 0b00_00010000
	} else if IsSubmoduleFloat32(f) {
		return 0b00_00100000
	} else {
		return 0b00_01000000
	}
}

func FClassD(f float64) uint64 {
	s := math.Signbit(f)
	if IsSNaN64(f) {
		return 0b01_00000000
	}
	if IsQNaN64(f) {
		return 0b10_00000000
	}
	if s {
		if f < -math.MaxFloat64 {
			return 0b00_00000001
		} else if f == 0 {
			return 0b00_00001000
		} else if IsSubmoduleFloat64(f) {
			return 0b00_00000100
		} else {
			return 0b00_00000010
		}
	}
	if f > math.MaxFloat64 {
		return 0b00_10000000
	} else if f == 0 {
		return 0b00_00010000
	} else if IsSubmoduleFloat64(f) {
		return 0b00_00100000
	} else {
		return 0b00_01000000
	}
}
