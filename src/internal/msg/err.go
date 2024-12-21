package msg

import "fmt"

type ConvMsgErr struct {
	Err error
}

func (e ConvMsgErr) Error() string {
	return e.Err.Error()
}

func NewConvMsgErr(s string) ConvMsgErr {
	err := fmt.Errorf("BytesからMessageへの変換に失敗しました: %s", s)
	return ConvMsgErr{Err: err}
}

type ConvBytesErr struct {
	Err error
}

func (e ConvBytesErr) Error() string {
	return e.Err.Error()
}

func NewConvBytesErr(s string) ConvBytesErr {
	err := fmt.Errorf("MessageからBytesへの変換に失敗しました: %s", s)
	return ConvBytesErr{Err: err}
}
