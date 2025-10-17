package vulpo

import (
	"strings"
)

type FieldDef struct {
	fieldname string
	fieldtype FieldType
	size      uint8
	decimals  uint8
	system    bool
	nullable  bool
	binary    bool
}

type FieldDefs struct {
	fields   []*FieldDef
	indicies map[string]int
}

func (flds *FieldDefs) Count() int {
	return len(flds.fields)
}

func (flds *FieldDefs) ByIndex(idx int) *FieldDef {
	if idx < 0 || idx >= len(flds.fields) {
		return nil
	}
	return flds.fields[idx]
}

// FieldDef exported getter methods
func (fd *FieldDef) Name() string {
	return fd.fieldname
}

func (fd *FieldDef) Type() FieldType {
	return fd.fieldtype
}

func (fd *FieldDef) Size() uint8 {
	return fd.size
}

func (fd *FieldDef) Decimals() uint8 {
	return fd.decimals
}

func (fd *FieldDef) IsSystem() bool {
	return fd.system
}

func (fd *FieldDef) IsNullable() bool {
	return fd.nullable
}

func (fd *FieldDef) IsBinary() bool {
	return fd.binary
}

func (flds *FieldDefs) checkCreateIndicies() {
	if flds.indicies == nil {
		flds.indicies = make(map[string]int)
	}
}

func (flds *FieldDefs) ByName(name string) *FieldDef {
	flds.checkCreateIndicies()
	idx, ok := flds.indicies[strings.ToLower(name)]
	if !ok {
		return nil
	}
	return flds.fields[idx]
}
