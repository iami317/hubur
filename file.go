package hubur

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"github.com/thoas/go-funk"
	"io"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	FileExistsMsgTemplate       = "file %s already exists, please backup and remove it at first"
	FileDoesNotExistMsgTemplate = "file %s does not exist"
	DynamicFileExts             = []string{"php", "jsp", "asp", "asa", "cer", "cdx", "php3"}
)

// DirExists 判断目录或文件是否存在
func DirExists(dir string) bool {
	info, err := os.Stat(dir)
	return (err == nil || os.IsExist(err)) && info.IsDir()
}

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

// ReadFile 读取指定path文件内容
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

// CreateFolder path
func CreateFolder(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// TarFile 对单个文件进行打包
func TarFile(sourceFullFile string, writer *tar.Writer) error {
	info, err := os.Stat(sourceFullFile)
	if err != nil {
		return err
	}
	// 创建头信息
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	// 头信息写入
	err = writer.WriteHeader(header)
	if err != nil {
		return err
	}
	// 读取源文件，将内容拷贝到tar.Writer中
	fr, err := os.Open(sourceFullFile)
	if err != nil {
		return err
	}
	defer func() {
		// 如果主程序的err为空nil，而文件句柄关闭err，则将关闭句柄的err返回
		if err2 := fr.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()
	if _, err = io.Copy(writer, fr); err != nil {
		return err
	}
	return nil
}

// TarFolder sourceFullPath为待打包目录，baseName为待打包目录的根目录名称
func TarFolder(sourceFullPath string, baseName string, writer *tar.Writer) error {
	// 保留最开始的原始目录，用于目录遍历过程中将文件由绝对路径改为相对路径
	baseFullPath := sourceFullPath
	return filepath.Walk(sourceFullPath, func(fileName string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 创建头信息
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		// 修改header的name，这里需要按照相对路径来
		// 说明这里是根目录，直接将目录名写入header即可
		if fileName == baseFullPath {
			header.Name = baseName
		} else {
			// 非根目录，需要对路径做处理：去掉绝对路径的前半部分，然后构造基于根目录的相对路径
			header.Name = filepath.Join(baseName, strings.TrimPrefix(fileName, baseFullPath))
		}

		if err = writer.WriteHeader(header); err != nil {
			return err
		}
		// linux文件有很多类型，这里仅处理普通文件，如业务需要处理其他类型的文件，这里添加相应的处理逻辑即可
		if !info.Mode().IsRegular() {
			return nil
		}
		// 普通文件，则创建读句柄，将内容拷贝到tarWriter中
		fr, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer fr.Close()
		if _, err := io.Copy(writer, fr); err != nil {
			return err
		}
		return nil
	})
}

// RemoveTargetFile 打包时如果目标的tar文件已经存在，则删除掉
func RemoveTargetFile(fileName string) (err error) {
	// 判断是否存在同名目标文件
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(fileName)
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()
	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
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

func WriteJson(filename, param string) error {
	fileHandle, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer fileHandle.Close()
	// NewWriter 默认缓冲区大小是 4096
	// 需要使用自定义缓冲区的writer 使用 NewWriterSize()方法
	buf := bufio.NewWriter(fileHandle)
	// 字节写入
	//buf.Write([]byte("buffer Write : " + param))
	// 字符串写入
	buf.WriteString(param + "\n")
	// 将缓冲中的数据写入
	return buf.Flush()
}

// WriteFileByte 按字节写入文件函数封装
func WriteFileByte(fileName string, content []byte) error {
	fileObj, err := os.Create(fileName)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(fileObj)
	defer writer.Flush()
	writer.Write(content)
	return nil
}

// FilePutContents - Write data to a file
func FilePutContents(filename string, data []byte) error {
	if dir := filepath.Dir(filename); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return ioutil.WriteFile(filename, data, 0644)
}

const (
	RandCharsetLength int    = 13 //随机字符串长度
	RandCharsetString string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_0123456789"
)

// RandCharset 生成随机字符串 长度13位
func RandCharset() string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, RandCharsetLength)
	for i := range b {
		b[i] = RandCharsetString[seededRand.Intn(len(RandCharsetString))]
	}
	return string(b)
}

// ChMod 改变文件权限
func ChMod(filename string, mode os.FileMode) bool {
	return os.Chmod(filename, mode) == nil
}
