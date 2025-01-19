package utils

import (
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
)

// UUID 是一个字节切片，用于表示UUID。
type UUID []byte

// rander 是全局的随机数生成器，默认使用 crypto/rand.Reader。
var rander = rand.Reader

// randomBits 为给定的字节切片填充随机位。
// 如果填充失败，则会触发 panic，因为随机数生成不应该失败。
func randomBits(b []byte) {
	if _, err := io.ReadFull(rander, b); err != nil {
		panic(err.Error()) // 随机数生成不应失败
	}
}

// NewRandom 生成一个新的随机UUID。
// 它首先创建一个16字节的随机数组，然后根据UUID版本4的规范设置特定的字节位。
// 返回值是一个符合UUID版本4规范的UUID。
func NewRandom() UUID {
	uuid := make([]byte, 16)
	randomBits(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40  // 设置版本号为4
	uuid[8] = (uuid[8] &^ 0x40) | 0x80 // 设置变体为RFC 4122
	return uuid
}

// WriteKeyFile 将内容写入指定文件，并确保目录存在且权限正确。
// 参数：
//   - file: 目标文件路径
//   - content: 要写入的内容
//
// 返回值：
//   - error: 如果操作过程中发生错误，则返回错误信息；否则返回nil
//
// 该函数的工作流程如下：
// 1. 确保目标文件所在目录存在，如果不存在则创建，权限为0700。
// 2. 创建一个临时文件并写入内容。
// 3. 写入成功后，将临时文件重命名为目标文件名。
func WriteKeyFile(file string, content []byte) error {
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()
	return os.Rename(f.Name(), file)
}
