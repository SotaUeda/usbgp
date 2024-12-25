package message

import "testing"

func TestHeaderMarshalAndUnmarshal(t *testing.T) {
	h, err := newHeader(19, Open)
	if err != nil {
		t.Error(err)
	}
	b, err := h.marshalBytes()
	if err != nil {
		t.Error(err)
	}
	h2 := &Header{}
	err = h2.unMarshalBytes(b)
	if err != nil {
		t.Error(err)
	}
	if !headerEqual(h, h2, t) {
		t.Errorf("header not equal: %v, %v", h, h2)
	}
}

func headerEqual(h1, h2 *Header, t *testing.T) bool {
	if h1.len != h2.len {
		t.Errorf("header length not equal: %v, %v", h1.len, h2.len)
		return false
	}
	if h1.msgType != h2.msgType {
		t.Errorf("header message type not equal: %v, %v", h1.msgType, h2.msgType)
		return false
	}
	return true
}
