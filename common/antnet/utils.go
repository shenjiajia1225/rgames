package antnet

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

func PathBase(p string) string {
	return path.Base(p)
}

func PathAbs(p string) string {
	f, err := filepath.Abs(p)
	if err != nil {
		LogError("get abs path failed path:%v err:%v", p, err)
		return ""
	}
	return f
}

func GetEXEDir() string {
	return PathDir(GetEXEPath())
}

func GetEXEPath() string {
	return PathAbs(os.Args[0])
}

func GetExeName() string {
	return PathBase(GetEXEPath())
}

func GetExeDir() string {
	return PathDir(GetEXEPath())
}

func GetExePath() string {
	return PathAbs(os.Args[0])
}

func GetEXEName() string {
	return PathBase(GetEXEPath())
}

func PathDir(p string) string {
	return path.Dir(p)
}

func PathExt(p string) string {
	return path.Ext(p)
}

func PathClean(p string) string {
	return path.Clean(p)
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func NewDir(path string) error {
	return os.MkdirAll(path, 0777)
}

func ReadFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		LogError("read file filed path:%v err:%v", path, err)
		return nil, ErrFileRead
	}
	return data, nil
}

func WriteFile(path string, data []byte) {
	dir := PathDir(path)
	if !PathExists(dir) {
		NewDir(dir)
	}
	ioutil.WriteFile(path, data, 0777)
}

func GetFiles(path string) []string {
	files := []string{}
	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files
}

func DelFile(path string) {
	os.Remove(path)
}

func DelDir(path string) {
	os.RemoveAll(path)
}

func CreateFile(path string) (*os.File, error) {
	dir := PathDir(path)
	if !PathExists(dir) {
		NewDir(dir)
	}
	return os.Create(path)
}

func CopyFile(dst io.Writer, src io.Reader) (written int64, err error) {
	return io.Copy(dst, src)
}

func walkDirTrue(dir string, wg *sync.WaitGroup, fun func(dir string, info os.FileInfo)) {
	wg.Add(1)
	defer wg.Done()
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		LogError("walk dir failed dir:%v err:%v", dir, err)
		return
	}
	for _, info := range infos {
		if info.IsDir() {
			fun(dir, info)
			subDir := filepath.Join(dir, info.Name())
			go walkDirTrue(subDir, wg, fun)
		} else {
			fun(dir, info)
		}
	}
}

func WalkDir(dir string, fun func(dir string, info os.FileInfo)) {
	if fun == nil {
		return
	}
	wg := &sync.WaitGroup{}
	walkDirTrue(dir, wg, fun)
	wg.Wait()
}

func FileCount(dir string) int32 {
	var count int32 = 0
	WalkDir(dir, func(dir string, info os.FileInfo) {
		if !info.IsDir() {
			atomic.AddInt32(&count, 1)
		}
	})
	return count
}

func DirCount(dir string) int32 {
	var count int32 = 0
	WalkDir(dir, func(dir string, info os.FileInfo) {
		if info.IsDir() {
			atomic.AddInt32(&count, 1)
		}
	})
	return count
}

func DirSize(dir string) int64 {
	var size int64 = 0
	WalkDir(dir, func(dir string, info os.FileInfo) {
		if !info.IsDir() {
			atomic.AddInt64(&size, info.Size())
		}
	})
	return size
}

// ---------------------------------------------------------------------------------------------

func ZlibCompress(data []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(data)
	w.Close()
	return in.Bytes()
}

func ZlibUnCompress(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r, _ := zlib.NewReader(b)
	defer r.Close()
	undatas, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return undatas, nil
}

func GZipCompress(data []byte) []byte {
	var in bytes.Buffer
	w := gzip.NewWriter(&in)
	w.Write(data)
	w.Close()
	return in.Bytes()
}

func GZipUnCompress(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r, _ := gzip.NewReader(b)
	defer r.Close()
	undatas, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return undatas, nil
}

// ---------------------------------------------------------------------------------------------

func Print(a ...interface{}) (int, error) {
	return fmt.Print(a...)
}
func Println(a ...interface{}) (int, error) {
	return fmt.Println(a...)
}
func Printf(format string, a ...interface{}) (int, error) {
	return fmt.Printf(format, a...)
}
func Sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

// ---------------------------------------------------------------------------------------------

func SplitStr(s string, sep string) []string {
	return strings.Split(s, sep)
}

func StrSplit(s string, sep string) []string {
	return strings.Split(s, sep)
}

func SplitStrN(s string, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

func StrSplitN(s string, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

func StrFind(s string, f string) int {
	return strings.Index(s, f)
}

func FindStr(s string, f string) int {
	return strings.Index(s, f)
}

func ReplaceStr(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

func StrReplace(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

func ReplaceMultStr(s string, oldnew ...string) string {
	r := strings.NewReplacer(oldnew...)
	return r.Replace(s)
}

func StrReplaceMult(s string, oldnew ...string) string {
	r := strings.NewReplacer(oldnew...)
	return r.Replace(s)
}

func TrimStrSpace(s string) string {
	return strings.TrimSpace(s)
}

func StrTrimSpace(s string) string {
	return strings.TrimSpace(s)
}

func TrimStr(s string, cutset []string) string {
	for _, v := range cutset {
		s = strings.Trim(s, v)
	}
	return s
}

func StrTrim(s string, cutset []string) string {
	return TrimStr(s, cutset)
}

func StrContains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func ContainsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}

func JoinStr(a []string, sep string) string {
	return strings.Join(a, sep)
}

func StrJoin(a []string, sep string) string {
	return strings.Join(a, sep)
}

func StrToLower(s string) string {
	return strings.ToLower(s)
}

func ToLowerStr(s string) string {
	return strings.ToLower(s)
}

func StrToUpper(s string) string {
	return strings.ToUpper(s)
}

func ToUpperStr(s string) string {
	return strings.ToUpper(s)
}

func StrTrimRight(s, cutset string) string {
	return strings.TrimRight(s, cutset)
}

func TrimRightStr(s, cutset string) string {
	return strings.TrimRight(s, cutset)
}
