package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"strings"
)

func WriteBytesToFile(data []byte, path string) error {
	// 创建文件夹，如果不存在的话
	err := os.MkdirAll(path[:len(path)-len(path[strings.LastIndex(path, "/"):])], 0755)

	if err != nil {
		return err
	}

	// 创建文件，如果不存在的话
	file, err := os.Create(path)

	if err != nil {
		return err
	}

	defer file.Close()

	// 写入数据到文件中
	_, err = file.Write(data)

	return err // 返回写入错误，如果有的话
}

func CreateDir(Dirpath string, perm fs.FileMode) (err error) {
	if Dirpath == "" {
		return nil
	} // 如果Dirpath的路径为空，则不做处理，直接返回nil

	if IsExist(Dirpath) {
		return nil
	} // 如果文件夹已存在，将不需要创建，直接返回nil

	err = os.MkdirAll(Dirpath, perm)

	return err // 返回nil,表示创建成功
}

// 判断路径是否存在函数
func IsExist(filename string) bool {

	// 获取文件信息 _ , err := os.Stat(filename)   // 这里输入你要测试的目录路径即可

	_, err := os.Stat(filename) // 得到的是一个os.FileInfo 接口类型 可以通过其查看文件把信息

	if err != nil { // 文件或者文件夹不存在时 返回false 便会执行上面的mkDIR命令去创建对应目录

		return false

	} else { // 文件存在时 返回true 即! false = ture 直接跳不去MKDIR 目录也就创建完了 teehee~

		return true

	}
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)

	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, in)

	if err != nil {
		return err
	}

	return out.Close()
}

func ObjectMD5(object interface{}) string {
	md5Ctx := md5.New()
	data, _ := json.Marshal(object)
	md5Ctx.Write(data)
	return hex.EncodeToString(md5Ctx.Sum(nil))
}
