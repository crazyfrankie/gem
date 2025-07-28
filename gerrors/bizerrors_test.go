package gerrors

import (
	"errors"
	"testing"
)

var (
	ErrNotFound         = NewBizError(40004, "resources not found")
	ErrInternalError    = NewBizError(50000, "internal server error")
	ErrInvalidParameter = NewBizError(40001, "invalid parameter")
)

func TestGlobalBizError_SetBizExtra_DoesNotModifyOriginal(t *testing.T) {
	// 记录原始状态
	originalCode := ErrNotFound.BizStatusCode()
	originalMessage := ErrNotFound.BizMessage()
	originalExtra := ErrNotFound.BizExtra()
	originalExtraLen := len(originalExtra)

	// 调用SetBizExtra创建新的错误实例
	newErr := ErrNotFound.SetBizExtra("request_id", "12345")

	// 验证原始全局变量没有被修改
	if ErrNotFound.BizStatusCode() != originalCode {
		t.Errorf("Original error code was modified. Expected: %d, Got: %d",
			originalCode, ErrNotFound.BizStatusCode())
	}

	if ErrNotFound.BizMessage() != originalMessage {
		t.Errorf("Original error message was modified. Expected: %s, Got: %s",
			originalMessage, ErrNotFound.BizMessage())
	}

	if len(ErrNotFound.BizExtra()) != originalExtraLen {
		t.Errorf("Original error extra was modified. Expected length: %d, Got length: %d",
			originalExtraLen, len(ErrNotFound.BizExtra()))
	}

	// 验证新错误实例包含了额外信息
	if newErr.BizStatusCode() != originalCode {
		t.Errorf("New error code is incorrect. Expected: %d, Got: %d",
			originalCode, newErr.BizStatusCode())
	}

	if newErr.BizMessage() != originalMessage {
		t.Errorf("New error message is incorrect. Expected: %s, Got: %s",
			originalMessage, newErr.BizMessage())
	}

	newExtra := newErr.BizExtra()
	if len(newExtra) != originalExtraLen+1 {
		t.Errorf("New error extra length is incorrect. Expected: %d, Got: %d",
			originalExtraLen+1, len(newExtra))
	}

	if newExtra["request_id"] != "12345" {
		t.Errorf("New error extra does not contain expected key-value. Expected: 12345, Got: %s",
			newExtra["request_id"])
	}
}

func TestGlobalBizError_AppendBizMessage_DoesNotModifyOriginal(t *testing.T) {
	// 记录原始状态
	originalCode := ErrInternalError.BizStatusCode()
	originalMessage := ErrInternalError.BizMessage()
	originalExtra := ErrInternalError.BizExtra()

	// 创建一个要追加的错误
	appendErr := errors.New("database connection failed")

	// 调用AppendBizMessage创建新的错误实例
	newErr := ErrInternalError.AppendBizMessage(appendErr)

	// 验证原始全局变量没有被修改
	if ErrInternalError.BizStatusCode() != originalCode {
		t.Errorf("Original error code was modified. Expected: %d, Got: %d",
			originalCode, ErrInternalError.BizStatusCode())
	}

	if ErrInternalError.BizMessage() != originalMessage {
		t.Errorf("Original error message was modified. Expected: %s, Got: %s",
			originalMessage, ErrInternalError.BizMessage())
	}

	if len(ErrInternalError.BizExtra()) != len(originalExtra) {
		t.Errorf("Original error extra was modified. Expected length: %d, Got length: %d",
			len(originalExtra), len(ErrInternalError.BizExtra()))
	}

	// 验证新错误实例包含了追加的消息
	expectedNewMessage := originalMessage + ", " + appendErr.Error()
	if newErr.BizMessage() != expectedNewMessage {
		t.Errorf("New error message is incorrect. Expected: %s, Got: %s",
			expectedNewMessage, newErr.BizMessage())
	}

	if newErr.BizStatusCode() != originalCode {
		t.Errorf("New error code is incorrect. Expected: %d, Got: %d",
			originalCode, newErr.BizStatusCode())
	}
}

func TestGlobalBizError_ChainedOperations_DoesNotModifyOriginal(t *testing.T) {
	// 记录原始状态
	originalCode := ErrInvalidParameter.BizStatusCode()
	originalMessage := ErrInvalidParameter.BizMessage()
	originalExtraLen := len(ErrInvalidParameter.BizExtra())

	// 链式调用多个操作
	appendErr := errors.New("validation failed")
	newErr := ErrInvalidParameter.
		SetBizExtra("field", "username").
		SetBizExtra("value", "invalid_user").
		AppendBizMessage(appendErr)

	// 验证原始全局变量完全没有被修改
	if ErrInvalidParameter.BizStatusCode() != originalCode {
		t.Errorf("Original error code was modified after chained operations. Expected: %d, Got: %d",
			originalCode, ErrInvalidParameter.BizStatusCode())
	}

	if ErrInvalidParameter.BizMessage() != originalMessage {
		t.Errorf("Original error message was modified after chained operations. Expected: %s, Got: %s",
			originalMessage, ErrInvalidParameter.BizMessage())
	}

	if len(ErrInvalidParameter.BizExtra()) != originalExtraLen {
		t.Errorf("Original error extra was modified after chained operations. Expected length: %d, Got length: %d",
			originalExtraLen, len(ErrInvalidParameter.BizExtra()))
	}

	// 验证新错误实例包含所有修改
	newExtra := newErr.BizExtra()
	if len(newExtra) != originalExtraLen+2 {
		t.Errorf("New error extra length is incorrect. Expected: %d, Got: %d",
			originalExtraLen+2, len(newExtra))
	}

	if newExtra["field"] != "username" {
		t.Errorf("New error missing expected extra field. Expected: username, Got: %s",
			newExtra["field"])
	}

	if newExtra["value"] != "invalid_user" {
		t.Errorf("New error missing expected extra value. Expected: invalid_user, Got: %s",
			newExtra["value"])
	}

	expectedMessage := originalMessage + ", " + appendErr.Error()
	if newErr.BizMessage() != expectedMessage {
		t.Errorf("New error message is incorrect. Expected: %s, Got: %s",
			expectedMessage, newErr.BizMessage())
	}
}

func TestGlobalBizError_AppendBizMessage_WithNilError(t *testing.T) {
	// 测试传入nil错误的情况
	originalMessage := ErrNotFound.BizMessage()

	newErr := ErrNotFound.AppendBizMessage(nil)

	// 当传入nil时，应该返回原错误实例
	if newErr != ErrNotFound {
		t.Error("AppendBizMessage with nil should return the original error instance")
	}

	if newErr.BizMessage() != originalMessage {
		t.Errorf("Message should remain unchanged when appending nil. Expected: %s, Got: %s",
			originalMessage, newErr.BizMessage())
	}
}

func TestGlobalBizError_SetBizExtra_MultipleKeys(t *testing.T) {
	// 测试多次设置不同的key不会相互影响
	originalExtraLen := len(ErrNotFound.BizExtra())

	err1 := ErrNotFound.SetBizExtra("key1", "value1")
	err2 := ErrNotFound.SetBizExtra("key2", "value2")

	// 验证原始错误没有被修改
	if len(ErrNotFound.BizExtra()) != originalExtraLen {
		t.Errorf("Original error was modified. Expected extra length: %d, Got: %d",
			originalExtraLen, len(ErrNotFound.BizExtra()))
	}

	// 验证两个新错误实例是独立的
	extra1 := err1.BizExtra()
	extra2 := err2.BizExtra()

	if extra1["key1"] != "value1" {
		t.Errorf("err1 missing expected key. Expected: value1, Got: %s", extra1["key1"])
	}

	if extra1["key2"] != "" {
		t.Errorf("err1 should not contain key2. Got: %s", extra1["key2"])
	}

	if extra2["key2"] != "value2" {
		t.Errorf("err2 missing expected key. Expected: value2, Got: %s", extra2["key2"])
	}

	if extra2["key1"] != "" {
		t.Errorf("err2 should not contain key1. Got: %s", extra2["key1"])
	}
}
