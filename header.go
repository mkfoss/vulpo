package vulpo

import "C"
import "time"

type Header struct {
	Recordcount uint
	LastUpdated time.Time
	HasIndex    bool
	HasFpt      bool
	DbfCodepage Codepage
}

//func (h *Header) RecordCount() uint {
//	return h.Recordcount
//}
//
//func (h *Header) LastUpdated() time.Time {
//	return h.LastUpdated
//}
//
//func (h *Header) HasIndex() bool {
//	return h.HasIndex
//}
//
//func (h *Header) HasFpt() bool {
//	return h.HasFpt
//}
//
//func (h *Header) DbfCodepage() DbfCodepage {
//	return h.DbfCodepage
//}
