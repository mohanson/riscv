package rv64

type isaI struct{}

func (_ *isaI) lui(c *CPU, rd uint64, imm uint64) (uint64, error) {
	c.SetRegister(rd, imm)
	c.SetPC(c.GetPC() + 4)
	return 1, nil
}

func (_ *isaI) aupic(c *CPU, rd uint64, imm uint64) (uint64, error) {
	c.SetRegister(rd, c.GetPC()+imm)
	c.SetPC(c.GetPC() + 4)
	return 1, nil
}

func (_ *isaI) jal(c *CPU, rd uint64, imm uint64) (uint64, error) {
	c.SetRegister(rd, c.GetPC()+4)
	r := c.GetPC() + imm
	if r%2 != 0x00 {
		return 0, ErrMisalignedInstructionFetch
	}
	c.SetPC(r)
	return 1, nil
}

func (_ *isaI) jalr(c *CPU, rd uint64, rs1 uint64, imm uint64) (uint64, error) {
	c.SetRegister(rd, c.GetPC()+4)
	r := (c.GetRegister(rs1) + imm) & 0xfffffffffffffffe
	if r%2 != 0x00 {
		return 0, ErrMisalignedInstructionFetch
	}
	c.SetPC(r)
	return 1, nil
}

func (_ *isaI) beq(c *CPU, rs1 uint64, rs2 uint64, imm uint64) (uint64, error) {
	if imm%2 != 0x00 {
		return 0, ErrMisalignedInstructionFetch
	}
	if c.GetRegister(rs1) == c.GetRegister(rs2) {
		c.SetPC(c.GetPC() + imm)
	} else {
		c.SetPC(c.GetPC() + 4)
	}
	return 1, nil
}

func (_ *isaI) bne(c *CPU, rs1 uint64, rs2 uint64, imm uint64) (uint64, error) {
	if imm%2 != 0x00 {
		return 0, ErrMisalignedInstructionFetch
	}
	if c.GetRegister(rs1) != c.GetRegister(rs2) {
		c.SetPC(c.GetPC() + imm)
	} else {
		c.SetPC(c.GetPC() + 4)
	}
	return 1, nil
}

func (_ *isaI) blt(c *CPU, rs1 uint64, rs2 uint64, imm uint64) (uint64, error) {
	if imm%2 != 0x00 {
		return 0, ErrMisalignedInstructionFetch
	}
	if int64(c.GetRegister(rs1)) < int64(c.GetRegister(rs2)) {
		c.SetPC(c.GetPC() + imm)
	} else {
		c.SetPC(c.GetPC() + 4)
	}
	return 1, nil
}

func (_ *isaI) bge(c *CPU, rs1 uint64, rs2 uint64, imm uint64) (uint64, error) {
	if imm%2 != 0x00 {
		return 0, ErrMisalignedInstructionFetch
	}
	if int64(c.GetRegister(rs1)) >= int64(c.GetRegister(rs2)) {
		c.SetPC(c.GetPC() + imm)
	} else {
		c.SetPC(c.GetPC() + 4)
	}
	return 1, nil
}

func (_ *isaI) bltu(c *CPU, rs1 uint64, rs2 uint64, imm uint64) (uint64, error) {
	if imm%2 != 0x00 {
		return 0, ErrMisalignedInstructionFetch
	}
	if c.GetRegister(rs1) < c.GetRegister(rs2) {
		c.SetPC(c.GetPC() + imm)
	} else {
		c.SetPC(c.GetPC() + 4)
	}
	return 1, nil
}

func (_ *isaI) bgeu(c *CPU, rs1 uint64, rs2 uint64, imm uint64) (uint64, error) {
	if imm%2 != 0x00 {
		return 0, ErrMisalignedInstructionFetch
	}
	if c.GetRegister(rs1) >= c.GetRegister(rs2) {
		c.SetPC(c.GetPC() + imm)
	} else {
		c.SetPC(c.GetPC() + 4)
	}
	return 1, nil
}

func (_ *isaI) lb(c *CPU, rd uint64, rs1 uint64, imm uint64) (uint64, error) {
	a := c.GetRegister(rs1) + imm
	b, err := c.GetMemory().GetUint8(a)
	if err != nil {
		return 0, err
	}
	v := SignExtend(uint64(b), 7)
	c.SetRegister(rd, v)
	c.SetPC(c.GetPC() + 4)
	return 1, nil
}

func (_ *isaI) lh(c *CPU, rd uint64, rs1 uint64, imm uint64) (uint64, error) {
	a := c.GetRegister(rs1) + imm
	b, err := c.GetMemory().GetUint16(a)
	if err != nil {
		return 0, err
	}
	v := SignExtend(uint64(b), 15)
	c.SetRegister(rd, v)
	c.SetPC(c.GetPC() + 4)
	return 1, nil
}

func (_ *isaI) lw(c *CPU, rd uint64, rs1 uint64, imm uint64) (uint64, error) {
	a := c.GetRegister(rs1) + imm
	b, err := c.GetMemory().GetUint32(a)
	if err != nil {
		return 0, err
	}
	v := SignExtend(uint64(b), 31)
	c.SetRegister(rd, v)
	c.SetPC(c.GetPC() + 4)
	return 1, nil
}

func (_ *isaI) ld(c *CPU, rd uint64, rs1 uint64, imm uint64) (uint64, error) {
	a := c.GetRegister(rs1) + imm
	b, err := c.GetMemory().GetUint64(a)
	if err != nil {
		return 0, err
	}
	v := b
	c.SetRegister(rd, v)
	c.SetPC(c.GetPC() + 4)
	return 1, nil
}

func (_ *isaI) lbu(c *CPU, rd uint64, rs1 uint64, imm uint64) (uint64, error) {
	a := c.GetRegister(rs1) + imm
	b, err := c.GetMemory().GetUint8(a)
	if err != nil {
		return 0, err
	}
	v := uint64(b)
	c.SetRegister(rd, v)
	c.SetPC(c.GetPC() + 4)
	return 1, nil
}

func (_ *isaI) lhu(c *CPU, rd uint64, rs1 uint64, imm uint64) (uint64, error) {
	a := c.GetRegister(rs1) + imm
	b, err := c.GetMemory().GetUint16(a)
	if err != nil {
		return 0, err
	}
	v := uint64(b)
	c.SetRegister(rd, v)
	c.SetPC(c.GetPC() + 4)
	return 1, nil
}

func (_ *isaI) lwu(c *CPU, rd uint64, rs1 uint64, imm uint64) (uint64, error) {
	a := c.GetRegister(rs1) + imm
	b, err := c.GetMemory().GetUint32(a)
	if err != nil {
		return 0, err
	}
	v := uint64(b)
	c.SetRegister(rd, v)
	c.SetPC(c.GetPC() + 4)
	return 1, nil
}

func (_ *isaI) sb()    {}
func (_ *isaI) sh()    {}
func (_ *isaI) addi()  {}
func (_ *isaI) slti()  {}
func (_ *isaI) sltiu() {}
func (_ *isaI) xori()  {}
func (_ *isaI) andi()  {}
func (_ *isaI) slli()  {}
func (_ *isaI) srli()  {}
func (_ *isaI) srai()  {}
func (_ *isaI) add()   {}

func (_ *isaI) sub(c *CPU, rd uint64, rs1 uint64, rs2 uint64) (uint64, error) {
	c.SetRegister(rd, c.GetRegister(rs1)-c.GetRegister(rs2))
	c.SetPC(c.GetPC() + 4)
	return 1, nil
}

func (_ *isaI) sll()    {}
func (_ *isaI) slt()    {}
func (_ *isaI) sltu()   {}
func (_ *isaI) xor()    {}
func (_ *isaI) srl()    {}
func (_ *isaI) sra()    {}
func (_ *isaI) or()     {}
func (_ *isaI) and()    {}
func (_ *isaI) fenci()  {}
func (_ *isaI) ecall()  {}
func (_ *isaI) ebreak() {}

func (_ *isaI) sd()    {}
func (_ *isaI) addiw() {}
func (_ *isaI) slliw() {}
func (_ *isaI) srliw() {}
func (_ *isaI) sraiw() {}
func (_ *isaI) addw()  {}
func (_ *isaI) subw()  {}
func (_ *isaI) sllw()  {}
func (_ *isaI) srlw()  {}
func (_ *isaI) sraw()  {}

type isaZifencei struct{}

func (_ *isaZifencei) fencei() {}

type isaZicsr struct{}

func (_ *isaZicsr) csrrw()  {}
func (_ *isaZicsr) csrrs()  {}
func (_ *isaZicsr) csrrc()  {}
func (_ *isaZicsr) csrrwi() {}
func (_ *isaZicsr) csrrsi() {}
func (_ *isaZicsr) csrrci() {}

type isaM struct{}

func (_ *isaM) mul()    {}
func (_ *isaM) mulh()   {}
func (_ *isaM) mulhsu() {}
func (_ *isaM) mulhu()  {}
func (_ *isaM) div()    {}
func (_ *isaM) divu()   {}
func (_ *isaM) rem()    {}
func (_ *isaM) remu()   {}
func (_ *isaM) mulw()   {}
func (_ *isaM) divw()   {}
func (_ *isaM) divuw()  {}
func (_ *isaM) remw()   {}
func (_ *isaM) remuw()  {}

type isaA struct{}

func (_ *isaA) lrw()      {}
func (_ *isaA) scw()      {}
func (_ *isaA) amoswapw() {}
func (_ *isaA) amoaddw()  {}
func (_ *isaA) amoxorw()  {}
func (_ *isaA) amoandw()  {}
func (_ *isaA) amoorw()   {}
func (_ *isaA) amominw()  {}
func (_ *isaA) amomaxw()  {}
func (_ *isaA) amominuw() {}
func (_ *isaA) amomaxuw() {}
func (_ *isaA) lrd()      {}
func (_ *isaA) scd()      {}
func (_ *isaA) amoswapd() {}
func (_ *isaA) amoaddd()  {}
func (_ *isaA) amoxord()  {}
func (_ *isaA) amoandd()  {}
func (_ *isaA) amoord()   {}
func (_ *isaA) amomind()  {}
func (_ *isaA) amomaxd()  {}
func (_ *isaA) amominud() {}
func (_ *isaA) amomaxud() {}

type isaF struct{}

func (_ *isaF) flw()     {}
func (_ *isaF) fsw()     {}
func (_ *isaF) fmadds()  {}
func (_ *isaF) fmsubs()  {}
func (_ *isaF) fnmsubs() {}
func (_ *isaF) fnmadds() {}
func (_ *isaF) fadds()   {}
func (_ *isaF) fsubs()   {}
func (_ *isaF) fmuls()   {}
func (_ *isaF) fdivs()   {}
func (_ *isaF) fsqrts()  {}
func (_ *isaF) fsgnjs()  {}
func (_ *isaF) fsgnjns() {}
func (_ *isaF) fsgnjxs() {}
func (_ *isaF) fmins()   {}
func (_ *isaF) fmaxs()   {}
func (_ *isaF) fcvtws()  {}
func (_ *isaF) fcvtwus() {}
func (_ *isaF) fmvxw()   {}
func (_ *isaF) feqs()    {}
func (_ *isaF) flts()    {}
func (_ *isaF) fles()    {}
func (_ *isaF) fclasss() {}
func (_ *isaF) fcvtsw()  {}
func (_ *isaF) fcvtswu() {}
func (_ *isaF) fmvwx()   {}
func (_ *isaF) fcvtls()  {}
func (_ *isaF) fcvtlus() {}
func (_ *isaF) fcvtsl()  {}
func (_ *isaF) fcvtslu() {}

type isaD struct{}

func (_ *isaD) fld()     {}
func (_ *isaD) fsd()     {}
func (_ *isaD) fmaddd()  {}
func (_ *isaD) fmsubd()  {}
func (_ *isaD) fnmsubd() {}
func (_ *isaD) fnmaddd() {}
func (_ *isaD) faddd()   {}
func (_ *isaD) fsubd()   {}
func (_ *isaD) fmuld()   {}
func (_ *isaD) fdivd()   {}
func (_ *isaD) fsqrtd()  {}
func (_ *isaD) fsgnjd()  {}
func (_ *isaD) fsgnjnd() {}
func (_ *isaD) fsgnjxd() {}
func (_ *isaD) fmind()   {}
func (_ *isaD) fmaxd()   {}
func (_ *isaD) fcvtsd()  {}
func (_ *isaD) fcvtds()  {}
func (_ *isaD) feqd()    {}
func (_ *isaD) fltd()    {}
func (_ *isaD) fled()    {}
func (_ *isaD) fclassd() {}
func (_ *isaD) fcvtwd()  {}
func (_ *isaD) fcvtwud() {}
func (_ *isaD) fcvtdw()  {}
func (_ *isaD) fcvtdwu() {}
func (_ *isaD) fcvtld()  {}
func (_ *isaD) fcvtlud() {}
func (_ *isaD) fmvxd()   {}
func (_ *isaD) fcvtdl()  {}
func (_ *isaD) fcvtdlu() {}
func (_ *isaD) fmvdx()   {}

type isaC struct{}

func (_ *isaC) addi4spn() {}
func (_ *isaC) fld()      {}
func (_ *isaC) lw()       {}
func (_ *isaC) ld()       {}
func (_ *isaC) fsd()      {}
func (_ *isaC) sw()       {}
func (_ *isaC) sd()       {}
func (_ *isaC) nop()      {}
func (_ *isaC) addi()     {}
func (_ *isaC) addiw()    {}
func (_ *isaC) li()       {}
func (_ *isaC) addi16sp() {}
func (_ *isaC) lui()      {}
func (_ *isaC) srli64()   {}
func (_ *isaC) srai64()   {}
func (_ *isaC) andi()     {}
func (_ *isaC) sub()      {}
func (_ *isaC) xor()      {}
func (_ *isaC) or()       {}
func (_ *isaC) and()      {}
func (_ *isaC) subw()     {}
func (_ *isaC) addw()     {}
func (_ *isaC) j()        {}
func (_ *isaC) beqz()     {}
func (_ *isaC) bnez()     {}
func (_ *isaC) slli64()   {}
func (_ *isaC) fldsp()    {}
func (_ *isaC) lwsp()     {}
func (_ *isaC) ldsp()     {}
func (_ *isaC) jr()       {}
func (_ *isaC) mv()       {}
func (_ *isaC) ebreak()   {}
func (_ *isaC) jalr()     {}
func (_ *isaC) add()      {}
func (_ *isaC) fsdsp()    {}
func (_ *isaC) sqsp()     {}
func (_ *isaC) swsp()     {}
func (_ *isaC) sdsp()     {}

var (
	aluI        = &isaI{}
	aluZifencei = &isaZifencei{}
	aluZicsr    = &isaZicsr{}
	aluM        = &isaM{}
	aluA        = &isaA{}
	aluF        = &isaF{}
	aluD        = &isaD{}
	aluC        = &isaC{}
)
