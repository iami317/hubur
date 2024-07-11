package hubur

import (
	"bufio"
	"github.com/thoas/go-funk"
	"io/ioutil"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsDir(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}
	return f.Mode().IsDir()
}

var (
	FileExistsMsgTemplate       = "file %s already exists, please backup and remove it at first"
	FileDoesNotExistMsgTemplate = "file %s does not exist"
)

var DynamicFileExts = []string{"php", "jsp", "asp", "asa", "cer", "cdx", "php3"}

func IsDynamicFileExt(ext string) bool {
	if ext == "" {
		return false
	}
	if ext[0] == '.' {
		ext = ext[1:]
	}
	return funk.Contains(DynamicFileExts, ext)
}

func GetTempFilePath() (string, error) {
	f, err := ioutil.TempFile("", "*")
	if err != nil {
		return "", err
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
	return f.Name(), nil
}

func GetExeRelativePath(ref string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), ref), nil
}

//读取指定path文件内容

func ReadFile(filePath string) []byte {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return []byte{}
	}
	return content
}

func IsFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

func WriteFile(filePath string, content string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	//及时关闭file句柄
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)

	_, err = write.WriteString(content)
	if err != nil {
		return err
	}

	//Flush将缓存的文件真正写入到文件中
	err = write.Flush()
	if err != nil {
		return err
	}
	return nil
}
