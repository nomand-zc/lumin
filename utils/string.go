package utils

import "unsafe"

// Bytes2Str 零拷贝：[]byte 转 string（只读场景使用）
func Bytes2Str(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// unsafe.String 是 Go 1.20+ 新增的安全零拷贝函数（替代旧的 unsafe.Pointer 方式）
	return unsafe.String(&b[0], len(b))
}

// Str2Bytes 零拷贝：string 转 []byte（只读场景使用）
func Str2Bytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	// unsafe.Slice 配合 unsafe.StringData 是 Go 1.20+ 推荐的零拷贝方式
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
