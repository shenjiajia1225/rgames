package antnet

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

func AddStopCheck(cs string) uint64 {
	id := atomic.AddUint64(&stopCheckIndex, 1)
	if id == 0 {
		id = atomic.AddUint64(&stopCheckIndex, 1)
	}
	stopCheckMap.Lock()
	stopCheckMap.M[id] = cs
	stopCheckMap.Unlock()
	return id
}

func RemoveStopCheck(id uint64) {
	stopCheckMap.Lock()
	delete(stopCheckMap.M, id)
	stopCheckMap.Unlock()
}

func AtExit(fun func()) {
	id := atomic.AddUint32(&atexitId, 1)
	if id == 0 {
		id = atomic.AddUint32(&atexitId, 1)
	}

	atexitMapSync.Lock()
	atexitMap[id] = fun
	atexitMapSync.Unlock()
}

func Stop() {
	if !atomic.CompareAndSwapInt32(&stop, 0, 1) {
		return
	}

	close(stopChanForGo)
	for sc := 0; !waitAll.TryWait(); sc++ {
		Sleep(1)
		if sc >= Config.StopTimeout {
			LogError("Server Stop Timeout")
			stopCheckMap.Lock()
			for _, v := range stopCheckMap.M {
				LogError("Server Stop Timeout:%v", v)
			}
			stopCheckMap.Unlock()
			break
		}
	}

	LogInfo("Server Stop")
	close(stopChanForSys)
}

func IsStop() bool {
	return stop == 1
}

func IsRuning() bool {
	return stop == 0
}

func CmdAct(cmd, act uint8) int {
	return int(cmd)<<8 + int(act)
}

func Tag(cmd, act uint8, index uint16) int {
	return int(cmd)<<16 + int(act)<<8 + int(index)
}

func MD5Str(s string) string {
	return MD5Bytes([]byte(s))
}

func MD5Bytes(s []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(s)
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func MD5File(path string) string {
	data, err := ReadFile(path)
	if err != nil {
		LogError("calc md5 failed path:%v", path)
		return ""
	}
	return MD5Bytes(data)
}

func WaitForSystemExit(atexit ...func()) {
	statis.StartTime = time.Now()
	signal.Notify(stopChanForSys, os.Interrupt, os.Kill, syscall.SIGTERM)
	select {
	case <-stopChanForSys:
		Stop()
	}
	for _, v := range atexit {
		if v != nil {
			v()
		}
	}
	atexitMapSync.Lock()
	for _, v := range atexitMap {
		v()
	}
	atexitMapSync.Unlock()
	//for _, v := range redisManagers {
	//	v.close()
	//}
	waitAllForRedis.Wait()
	if !atomic.CompareAndSwapInt32(&stopForLog, 0, 1) {
		return
	}
	close(stopChanForLog)
	waitAllForLog.Wait()
}

func Daemon(skip ...string) {
	if os.Getppid() != 1 {
		filePath, _ := filepath.Abs(os.Args[0])
		newCmd := []string{os.Args[0]}
		add := 0
		for _, v := range os.Args[1:] {
			if add == 1 {
				add = 0
				continue
			} else {
				add = 0
			}
			for _, s := range skip {
				if strings.Contains(v, s) {
					if strings.Contains(v, "--") {
						add = 2
					} else {
						add = 1
					}
					break
				}
			}
			if add == 0 {
				newCmd = append(newCmd, v)
			}
		}
		Println("go deam args:", newCmd)
		cmd := exec.Command(filePath)
		cmd.Args = newCmd
		cmd.Start()
	}
}

func GetStatis() *Statis {
	statis.GoCount = int(gocount)
	statis.MsgqueCount = len(msgqueMap)
	statis.PoolGoCount = poolGoCount
	return statis
}

func Atoi(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

func Atoi64(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		LogError("str to int64 failed err:%v", err)
		return 0
	}
	return i
}

func Atof(str string) float32 {
	i, err := strconv.ParseFloat(str, 32)
	if err != nil {
		LogError("str to int64 failed err:%v", err)
		return 0
	}
	return float32(i)
}

func Atof64(str string) float64 {
	i, err := strconv.ParseFloat(str, 64)
	if err != nil {
		LogError("str to int64 failed err:%v", err)
		return 0
	}
	return i
}

func Itoa(num interface{}) string {
	switch n := num.(type) {
	case int8:
		return strconv.FormatInt(int64(n), 10)
	case int16:
		return strconv.FormatInt(int64(n), 10)
	case int32:
		return strconv.FormatInt(int64(n), 10)
	case int:
		return strconv.FormatInt(int64(n), 10)
	case int64:
		return strconv.FormatInt(int64(n), 10)
	case uint8:
		return strconv.FormatUint(uint64(n), 10)
	case uint16:
		return strconv.FormatUint(uint64(n), 10)
	case uint32:
		return strconv.FormatUint(uint64(n), 10)
	case uint:
		return strconv.FormatUint(uint64(n), 10)
	case uint64:
		return strconv.FormatUint(uint64(n), 10)
	}
	return ""
}

func Try(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			if handler == nil {
				LogStack()
				LogError("error catch:%v", err)
			} else {
				handler(err)
			}
			atomic.AddInt32(&statis.PanicCount, 1)
			statis.LastPanic = int(Timestamp)
		}
	}()
	fun()
}

func Try2(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			LogStack()
			LogError("error catch:%v", err)
			if handler != nil {
				handler(err)
			}
			atomic.AddInt32(&statis.PanicCount, 1)
			statis.LastPanic = int(Timestamp)
		}
	}()
	fun()
}

func ParseBaseKind(kind reflect.Kind, data string) (interface{}, error) {
	switch kind {
	case reflect.String:
		return data, nil
	case reflect.Bool:
		v := data == "1" || data == "true"
		return v, nil
	case reflect.Int:
		x, err := strconv.ParseInt(data, 0, 64)
		return int(x), err
	case reflect.Int8:
		x, err := strconv.ParseInt(data, 0, 8)
		return int8(x), err
	case reflect.Int16:
		x, err := strconv.ParseInt(data, 0, 16)
		return int16(x), err
	case reflect.Int32:
		x, err := strconv.ParseInt(data, 0, 32)
		return int32(x), err
	case reflect.Int64:
		x, err := strconv.ParseInt(data, 0, 64)
		return int64(x), err
	case reflect.Float32:
		x, err := strconv.ParseFloat(data, 32)
		return float32(x), err
	case reflect.Float64:
		x, err := strconv.ParseFloat(data, 64)
		return float64(x), err
	case reflect.Uint:
		x, err := strconv.ParseUint(data, 10, 64)
		return uint(x), err
	case reflect.Uint8:
		x, err := strconv.ParseUint(data, 10, 8)
		return uint8(x), err
	case reflect.Uint16:
		x, err := strconv.ParseUint(data, 10, 16)
		return uint16(x), err
	case reflect.Uint32:
		x, err := strconv.ParseUint(data, 10, 32)
		return uint32(x), err
	case reflect.Uint64:
		x, err := strconv.ParseUint(data, 10, 64)
		return uint64(x), err
	default:
		LogError("parse failed type not found type:%v data:%v", kind, data)
		return nil, errors.New("type not found")
	}
}

func GobPack(e interface{}) ([]byte, error) {
	var bio bytes.Buffer
	enc := gob.NewEncoder(&bio)
	err := enc.Encode(e)
	if err != nil {
		return nil, ErrGobPack
	}
	return bio.Bytes(), nil
}

func GobUnPack(data []byte, msg interface{}) error {
	bio := bytes.NewBuffer(data)
	enc := gob.NewDecoder(bio)
	err := enc.Decode(msg)
	if err != nil {
		return ErrGobUnPack
	}
	return nil
}

func ParseBool(s string) bool {
	if s == "1" || s == "true" {
		return true
	}
	return false
}

func ParseUint32(s string) uint32 {
	value, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(value)
}
func ParseUint64(s string) uint64 {
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return uint64(value)
}

func ParseInt32(s string) int32 {
	value, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int32(value)
}
func ParseInt64(s string) int64 {
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return int64(value)
}

//固定形式  x&y&z
func Split1(s string, retSlice *[]uint32) {
	slice := strings.Split(s, "&")
	*retSlice = make([]uint32, 0, len(slice))
	for _, value := range slice {
		*retSlice = append(*retSlice, ParseUint32(value))
	}
}

//固定形式   x&y&z;a&b&c;l_m_n
func Split2(s string, retSlice *[][]uint32) {
	slice1 := strings.Split(s, ";")
	*retSlice = make([][]uint32, 0, len(slice1))
	for _, value := range slice1 {
		var sl1 []uint32
		Split1(value, &sl1)
		*retSlice = append(*retSlice, sl1)
	}
}

//固定形式 x&y&z;a&b&c:x&y&z;a&b&c
func Split3(s string, retSlice *[][][]uint32) {
	slice1 := strings.Split(s, ":")
	*retSlice = make([][][]uint32, 0, len(slice1))
	for _, value := range slice1 {
		var sl2 [][]uint32
		Split2(value, &sl2)
		*retSlice = append(*retSlice, sl2)
	}
}

//固定形式  x&y&z
func SplitString1(s string, retSlice *[]string) {
	*retSlice = strings.Split(s, "&")
}

//固定形式  x&y&z;a&b&c;l_m_n
func SplitString2(s string, retSlice *[][]string) {
	slice1 := strings.Split(s, ";")
	*retSlice = make([][]string, 0, len(slice1))
	for _, value := range slice1 {
		*retSlice = append(*retSlice, strings.Split(value, "&"))
	}
}

//固定形式  x&y&z;a&b&c:x&y&z;a&b&c:
func SplitString3(s string, retSlice *[][][]string) {
	slice1 := strings.Split(s, ":")
	*retSlice = make([][][]string, 0, len(slice1))
	for _, value := range slice1 {
		var sl2 [][]string
		SplitString2(value, &sl2)
		*retSlice = append(*retSlice, sl2)
	}
}

//随机数返回[min,max)
func RandBetween(min, max int) int {
	if min >= max || max == 0 {
		return max
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := r.Intn(max-min) + min
	return random
}

func RandString(count int) string {
	var randomstr string
	for r := 0; r < count; r++ {
		i := RandBetween(65, 90)
		a := rune(i)
		randomstr += string(a)
	}
	return randomstr
}

//随机数返回[min,max)中的count个不重复数值
//一般用来从数组中随机一部分数据的下标
//2个随机数种子保证参数相同，返回值不一定相同，达到伪随机目的
func RandSliceBetween(min, max, count int) []int {
	if min > max {
		min, max = max, min
	}
	if min == max || max == 0 || count <= 0 {
		return []int{max}
	}
	randomRange := max - min
	retSlice := make([]int, 0, count)
	if count >= randomRange {
		for i := min; i < max; i++ {
			retSlice = append(retSlice, i)
		}
		return retSlice
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))
	random := r.Intn(randomRange) + min
	baseRand := RandBetween(random*min, random*max)
	retSlice = append(retSlice, random)
	for i := 1; i < count; i++ {
		random = (i+baseRand*random)%randomRange + min
		isReapeated := false
		for j := 0; j < count; j++ {
			for _, v := range retSlice {
				if random == v {
					isReapeated = true
					break
				}
			}
			if isReapeated {
				random = (random-min+1)%randomRange + min
			} else {
				break
			}
		}
		retSlice = append(retSlice, random)
	}

	return retSlice
}

type valueWeightItem struct {
	weight uint32
	value  uint64
}

// 权值对，根据权重随机一个值出来
type GBValueWeightPair struct {
	allweight uint32
	valuelist []*valueWeightItem
}

func NewValueWeightPair() *GBValueWeightPair {
	vwp := new(GBValueWeightPair)
	vwp.valuelist = make([]*valueWeightItem, 0, 0)
	return vwp
}

func (this *GBValueWeightPair) Add(weight uint32, value uint64) {
	if weight == 0 {
		return
	}
	valueinfo := &valueWeightItem{weight, value}
	this.valuelist = append(this.valuelist, valueinfo)
	this.allweight += weight
}

func (this *GBValueWeightPair) Random() uint64 {
	if this.allweight > 0 {
		randvalue := uint32(rand.Intn(int(this.allweight))) + 1 //[1,allweight]
		addweight := uint32(0)
		for i := 0; i < len(this.valuelist); i++ {
			addweight += this.valuelist[i].weight
			if randvalue <= addweight {
				return this.valuelist[i].value
			}
		}
	}
	return 0
}

func SafeSubInt32(a, b int32) int32 {
	if a > b {
		return a - b
	}
	return 0
}

func SafeSub(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return 0
}
func SafeSub64(a, b uint64) uint64 {
	if a > b {
		return a - b
	}
	return 0
}

func SafeSubInt64(a, b int64) int64 {
	if a > b {
		return a - b
	}
	return 0
}

// ---------------------------------------------------------------------------------------------

func Go(fn func()) {
	pc := Config.PoolSize + 1
	select {
	case poolChan <- fn:
		return
	default:
		pc = atomic.AddInt32(&poolGoCount, 1)
		if pc > Config.PoolSize {
			atomic.AddInt32(&poolGoCount, -1)
		}
	}

	waitAll.Add(1)
	var debugStr string
	id := atomic.AddUint32(&goid, 1)
	c := atomic.AddInt32(&gocount, 1)
	if DefLog.Level() <= LogLevelDebug {
		debugStr = LogSimpleStack()
		LogDebug("goroutine start id:%d count:%d from:%s", id, c, debugStr)
	}
	go func() {
		Try(fn, nil)
		for pc <= Config.PoolSize {
			select {
			case <-stopChanForGo:
				pc = Config.PoolSize + 1
			case nfn := <-poolChan:
				Try(nfn, nil)
			}
		}

		waitAll.Done()
		c = atomic.AddInt32(&gocount, -1)

		if DefLog.Level() <= LogLevelDebug {
			LogDebug("goroutine end id:%d count:%d from:%s", id, c, debugStr)
		}
	}()
}

func Go2(fn func(cstop chan struct{})) {
	Go(func() {
		Try(func() { fn(stopChanForGo) }, nil)
	})
}

func GoArgs(fn func(...interface{}), args ...interface{}) {
	waitAll.Add(1)
	var debugStr string
	id := atomic.AddUint32(&goid, 1)
	c := atomic.AddInt32(&gocount, 1)
	if DefLog.Level() <= LogLevelDebug {
		debugStr = LogSimpleStack()
		LogDebug("goroutine start id:%d count:%d from:%s", id, c, debugStr)
	}

	go func() {
		Try(func() { fn(args...) }, nil)

		waitAll.Done()
		c = atomic.AddInt32(&gocount, -1)
		if DefLog.Level() <= LogLevelDebug {
			LogDebug("goroutine end id:%d count:%d from:%s", id, c, debugStr)
		}
	}()
}

func goForRedis(fn func()) {
	waitAllForRedis.Add(1)
	var debugStr string
	id := atomic.AddUint32(&goid, 1)
	c := atomic.AddInt32(&gocount, 1)
	if DefLog.Level() <= LogLevelDebug {
		debugStr = LogSimpleStack()
		LogDebug("goroutine start id:%d count:%d from:%s", id, c, debugStr)
	}
	go func() {
		Try(fn, nil)
		waitAllForRedis.Done()
		c = atomic.AddInt32(&gocount, -1)

		if DefLog.Level() <= LogLevelDebug {
			LogDebug("goroutine end id:%d count:%d from:%s", id, c, debugStr)
		}
	}()
}

func goForLog(fn func(cstop chan struct{})) bool {
	if IsStop() {
		return false
	}
	waitAllForLog.Add(1)

	go func() {
		fn(stopChanForLog)
		waitAllForLog.Done()
	}()
	return true
}

// ---------------------------------------------------------------------------------------------

func Send(msg *Message, fun func(msgque IMsgQue) bool) {
	if msg == nil {
		return
	}
	c := make(chan struct{})
	gmsgMapSync.Lock()
	gmsg := gmsgArray[gmsgId]
	gmsgArray[gmsgId+1] = &gMsg{c: c}
	gmsgId++
	gmsgMapSync.Unlock()
	gmsg.msg = msg
	gmsg.fun = fun
	close(gmsg.c)
}

func SendGroup(group string, msg *Message) {
	Send(msg, func(msgque IMsgQue) bool {
		return msgque.IsInGroup(group)
	})
}

func HttpGetWithBasicAuth(url, name, passwd string) (string, error, *http.Response) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", ErrHttpRequest, nil
	}
	req.SetBasicAuth(name, passwd)
	resp, err := client.Do(req)
	if err != nil {
		return "", ErrHttpRequest, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", ErrHttpRequest, nil
	}
	resp.Body.Close()
	return string(body), nil, resp
}

func HttpGet(url string) (string, error, *http.Response) {
	resp, err := http.Get(url)
	if err != nil {
		return "", ErrHttpRequest, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", ErrHttpRequest, resp
	}
	resp.Body.Close()
	return string(body), nil, resp
}

func HttpPost(url, form string) (string, error, *http.Response) {
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(form))
	if err != nil {
		return "", ErrHttpRequest, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", ErrHttpRequest, resp
	}
	resp.Body.Close()
	return string(body), nil, resp
}

func HttpUpload(url, field, file string) (*http.Response, error) {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	formFile, err := writer.CreateFormFile(field, file)
	if err != nil {
		LogError("create form file failed:%s\n", err)
		return nil, err
	}

	srcFile, err := os.Open(file)
	if err != nil {
		LogError("%open source file failed:%s\n", err)
		return nil, err
	}
	defer srcFile.Close()
	_, err = io.Copy(formFile, srcFile)
	if err != nil {
		LogError("write to form file falied:%s\n", err)
		return nil, err
	}

	contentType := writer.FormDataContentType()
	writer.Close()
	resp, err := http.Post(url, contentType, buf)
	if err != nil {
		LogError("post failed:%s\n", err)
	}
	resp.Body.Close()
	return resp, err
}

func SendMail(user, password, host, to, subject, body, mailtype string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + user + ">\r\nSubject: " + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, send_to, msg)
	return err
}

var allIp []string

func GetSelfIp(ifnames ...string) []string {
	if allIp != nil {
		return allIp
	}
	inters, _ := net.Interfaces()
	if len(ifnames) == 0 {
		ifnames = []string{"eth", "lo", "eno", "无线网络连接", "本地连接", "以太网"}
	}

	filterFunc := func(name string) bool {
		for _, v := range ifnames {
			if strings.Index(name, v) != -1 {
				return true
			}
		}
		return false
	}

	for _, inter := range inters {
		if !filterFunc(inter.Name) {
			continue
		}
		addrs, _ := inter.Addrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					allIp = append(allIp, ipnet.IP.String())
				}
			}
		}
	}
	return allIp
}

func IsIntraIp(ip string) bool {
	if ip == "127.0.0.1" {
		return true
	}
	ips := strings.Split(ip, ".")
	ipA := ips[0]
	if ipA == "10" {
		return true
	}
	ipB := ips[1]

	if ipA == "192" {
		if ipB == "168" {
			return true
		}
	}

	if ipA == "172" {
		ipb := Atoi(ipB)
		if ipb >= 16 && ipb <= 31 {
			return true
		}
	}

	return false
}

func GetSelfIntraIp(ifnames ...string) (ips []string) {
	all := GetSelfIp(ifnames...)
	for _, v := range all {
		if IsIntraIp(v) {
			ips = append(ips, v)
		}
	}

	return
}

func GetSelfExtraIp(ifnames ...string) (ips []string) {
	all := GetSelfIp(ifnames...)
	for _, v := range all {
		if IsIntraIp(v) {
			continue
		}
		ips = append(ips, v)
	}

	return
}

func IPCanUse(ip string) bool {
	var err error
	for port := 1024; port < 65535; port++ {
		addr := Sprintf("%v:%v", ip, port)
		listen, err := net.Listen("tcp", addr)
		if err == nil {
			listen.Close()
			break
		} else if StrFind(err.Error(), "address is not valid") != -1 {
			return false
		}
	}
	return err == nil
}
