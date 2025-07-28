package gerrors

import (
	"errors"
	"fmt"
)

type BizErrorIface interface {
	BizStatusCode() int32
	BizMessage() string
	BizExtra() map[string]string
	Error() string
}

type BizError struct {
	code  int32
	msg   string
	extra map[string]string
}

// FromBizStatusError converts err to BizErrorIface.
func FromBizStatusError(err error) (bizErr BizErrorIface, ok bool) {
	if err == nil {
		return
	}
	ok = errors.As(err, &bizErr)
	return
}

// NewBizError returns BizErrorIface by passing code and msg.
func NewBizError(code int32, msg string) *BizError {
	return &BizError{
		code:  code,
		msg:   msg,
		extra: make(map[string]string),
	}
}

// NewBizErrorWithExtra returns BizErrorIface which contains extra info.
func NewBizErrorWithExtra(code int32, msg string, extra map[string]string) *BizError {
	return &BizError{code: code, msg: msg, extra: extra}
}

func (b *BizError) Error() string {
	return fmt.Sprintf("biz error: code=%d, err=%s", b.code, b.msg)
}

func (b *BizError) BizStatusCode() int32 {
	return b.code
}

func (b *BizError) BizMessage() string {
	return b.msg
}

func (b *BizError) BizExtra() map[string]string {
	return b.extra
}

func (b *BizError) SetBizExtra(key, val string) *BizError {
	newErr := *b

	newExtra := make(map[string]string)
	for k, v := range b.extra {
		newExtra[k] = v
	}
	newExtra[key] = val

	newErr.extra = newExtra
	return &newErr
}

func (b *BizError) AppendBizMessage(err error) *BizError {
	if err == nil {
		return b
	}

	return &BizError{
		code:  b.code,
		msg:   fmt.Sprintf("%s, %s", b.msg, err.Error()),
		extra: b.extra,
	}
}
