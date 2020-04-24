// +build windows

package utils

//import "C"
import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yanzongzhen/Logger/logger"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//var h hash.Hash
//func init() {
//	h = md5.New()
//}

var charSetRegStr string
var regMap map[string]*regexp.Regexp

func init() {
	charSetRegStr = `[=][\S][^;]+`
	regMap = make(map[string]*regexp.Regexp)
	charSetReg := regexp.MustCompile(charSetRegStr)
	regMap[charSetRegStr] = charSetReg
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func EncodeObject(object interface{}) ([]byte, error) {
	return JSONMarshal(object)
}

func IsEmpty(s ...string) bool {
	for _, str := range s {
		if !(len(str) > 0) {
			return true
		}
	}
	return false
}

func DecodeObject(object interface{}, data []byte) error {
	//buffer := bytes.NewBuffer(data)
	//decoder := gob.NewDecoder(buffer)
	//err := decoder.Decode(object)
	return json.Unmarshal(data, object)
}

func DigestMessage(msg []byte) (string, error) {
	h := md5.New()
	h.Write(msg)
	digest := h.Sum(nil)
	return hex.EncodeToString(digest), nil
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		//log.Fatal(err)
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func CheckPanicError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func ToJsonStr(data interface{}, defaultValue string) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return defaultValue
	}

	return string(jsonData)
}

func DeleteFile(path string) error {
	//if f, err := os.Open(path); os.IsNotExist(err) {
	//	return nil
	//} else {
	//}
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err

}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func GetCommonPrefix(s1 string, s2 string) string {
	//minLen := math.Min(len(s1), len(s2))

	minLen := Min(len(s1), len(s2))
	//res := ""
	index := 0
	for i := 0; i < minLen && s1[i] == s2[i]; i++ {
		//res = append(res, s1[i])
		index = i
	}
	if index > 0 {
		return s1[0 : index+1]
	} else {
		return ""
	}
}

func GetChineseCommonPrefix(s1 string, s2 string) (string, int) {

	s1Rune := make([]rune, 0, 1)
	s2Rune := make([]rune, 0, 1)
	for _, r := range s1 {
		s1Rune = append(s1Rune, r)
	}

	for _, r := range s2 {
		s2Rune = append(s2Rune, r)
	}

	minLen := Min(len(s1Rune), len(s2Rune))

	//res := ""
	index := 0
	for i := 0; i < minLen && s1Rune[i] == s2Rune[i]; i++ {
		//res = append(res, s1[i])
		index = i
	}
	if index > 0 {
		return string(s1Rune[0 : index+1]), index + 1
	} else {
		return "", -1
	}
}

func GetCharSetFromContentType(contentType string) string {
	if regMap[charSetRegStr].MatchString(contentType) {
		charsetStr := regMap[charSetRegStr].FindString(contentType)
		return charsetStr[1:]
	}
	return ""
}

func ConvertToUTF8(src []byte, sourceCharset string) ([]byte, error) {
	srcReader := bytes.NewReader(src)
	sourceCharset = strings.ToLower(sourceCharset)
	switch sourceCharset {
	case "gbk":
		out := transform.NewReader(srcReader, simplifiedchinese.GBK.NewDecoder())
		return ioutil.ReadAll(out)
	case "gb18030":
		out := transform.NewReader(srcReader, simplifiedchinese.GB18030.NewDecoder())
		return ioutil.ReadAll(out)
	case "gb2312":
		out := transform.NewReader(srcReader, simplifiedchinese.HZGB2312.NewDecoder())
		return ioutil.ReadAll(out)
	case "utf-8":
		return src, nil
	case "utf8":
		return src, nil
	case "iso-8859-1", "iso8859-1":
		return src, nil
		//return []byte(iso8859ToUtf(src)), nil
	default:
		logger.Errorln("不支持的编码格式")
		return src, errors.New("不支持的编码格式")

	}
}

func iso8859ToUtf(data []byte) string {
	buf := make([]rune, len(data))
	for i, b := range data {
		buf[i] = rune(b)
	}
	return string(buf)
}

func ConvertUTF8To(src []byte, toCharset string) ([]byte, error) {
	switch toCharset {
	case "gbk":
		return convertTo(src, simplifiedchinese.GBK.NewEncoder())
	case "gb2312":
		return convertTo(src, simplifiedchinese.HZGB2312.NewEncoder())
	case "utf-8":
		return src, nil
	default:
		return src, nil
	}
}

func convertTo(data []byte, encoder *encoding.Encoder) ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	writer := transform.NewWriter(buffer, encoder)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Traverse(list []interface{}, condition func(item interface{}) bool) bool {
	for _, item := range list {
		if condition(item) {
			return true
		}
	}
	return false
}

func RegxMatchAll(regxStr string, src string) []string {
	if reg, ok := regMap[regxStr]; ok {
		return reg.FindAllString(src, -1)
	} else {
		reg, _ = regexp.Compile(regxStr)
		return reg.FindAllString(src, -1)
	}
}

func RegxMatchOne(regxStr string, src string) string {
	if reg, ok := regMap[regxStr]; ok {
		return reg.FindString(src)
	} else {
		reg, _ = regexp.Compile(regxStr)
		return reg.FindString(src)
	}
}

func ExecCmd(name string, cmdStr []string) string {
	cmd := exec.Command(name, cmdStr...)
	var out bytes.Buffer
	var err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	errRun := cmd.Run()

	if errRun != nil {
		logger.Error(errRun)
	}
	return out.String()
}

func IsExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

//func BytesToCChar(msg []byte) *C.char {
//	logger.Debugf("BytesToCChar, msg : %s", string(msg))
//	return (*C.char)(unsafe.Pointer(&msg[0]))
//}
//
//func BytesListToCStringList(msg [][]byte) **C.char {
//	argv := make([]*C.char, len(msg))
//	for _, s := range msg {
//		cs := (*C.char)(unsafe.Pointer(&s[0]))
//		logger.Debugf("BytesListToCStringList, msg : %s", string(s))
//		argv = append(argv, cs)
//	}
//	return (**C.char)(unsafe.Pointer(&argv[0]))
//}

func GetLocalAddr() (string, error) {
	var ip string
	addrSlice, err := net.InterfaceAddrs()
	if err != nil {
		logger.Errorf("获取本地IP地址失败:%v", err)
		return "", err
	}
	for _, addr := range addrSlice {
		if iPNet, ok := addr.(*net.IPNet); ok && !iPNet.IP.IsLoopback() {
			if nil != iPNet.IP.To4() {
				ip = iPNet.IP.String()
				logger.Debugf("本机IP为:%v", ip)
				return ip, nil
			}
		}
	}
	return "", errors.New("获取本地IP地址失败")
}

func SendMail(msg string, subject string, toMail ...string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "jishuo1213@126.com")
	m.SetHeader("To", toMail...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", msg)

	d := gomail.NewPlainDialer("smtp.126.com", 465, "jishuo1213@126.com", "6629589y")

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {

		logger.Errorf("send mail failed")
	}
}
func SendMailWithParam(msg string, subject string, host string, port int, email string, password string, toMail ...string) {
	m := gomail.NewMessage()
	m.SetHeader("From", email)
	m.SetHeader("To", toMail...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", msg)
	d := gomail.NewPlainDialer(host, 465, email, password)
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		logger.Errorf("send mail failed")
	}
}

func Retry(f func() error, s string) {
	go func() {
		times := 0
		for {
			times++
			err := f()
			if err != nil {
				time.Sleep(time.Second * 5)
				logger.Errorf("retry %s 失败 次数 %d", s, times)

				if times > 10 {
					logger.Errorf("retry %s 失败 次数 %d", s, times)
					SendMail(fmt.Sprintf("重试十次 %s 失败", s), "警告！！！！", "fanjsh@inspur.com")
				}
			} else {
				logger.Infof("retry %s 成功", s)
				break
			}
		}
	}()
}

func JudgeTime(begDate string, endDate string) bool {
	beginTime, err := time.ParseInLocation("2006-01-02 15:04:05", begDate, time.Local)
	if err != nil {
		return false
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", endDate, time.Local)
	if err != nil {
		return false
	}
	if beginTime.Before(endTime) {
		return false
	} else {
		return true
	}
}
func ContainsHan(param ...string) bool {
	reg, _ := regexp.Compile("[\u4e00-\u9fa5]")
	for _, p := range param {
		if len(reg.Find([]byte(p))) > 0 {
			return true
		}
	}
	return false
}

//type DiskStatus struct {
//	All  uint64 `json:"all"`
//	Used uint64 `json:"used"`
//	Free uint64 `json:"free"`
//}
//
//func DiskUsage(path string) (disk DiskStatus, err error) {
//	fs := syscall.Statfs_t{}
//	err = syscall.Statfs(path, &fs)
//	if err != nil {
//		return
//	}
//	disk.All = fs.Blocks * uint64(fs.Bsize)
//	disk.Free = fs.Bfree * uint64(fs.Bsize)
//	disk.Used = disk.All - disk.Free
//	return
//}

var weight = [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
var valid_value = [11]byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}
var valid_province = map[string]string{
	"11": "北京市",
	"12": "天津市",
	"13": "河北省",
	"14": "山西省",
	"15": "内蒙古自治区",
	"21": "辽宁省",
	"22": "吉林省",
	"23": "黑龙江省",
	"31": "上海市",
	"32": "江苏省",
	"33": "浙江省",
	"34": "安徽省",
	"35": "福建省",
	"36": "山西省",
	"37": "山东省",
	"41": "河南省",
	"42": "湖北省",
	"43": "湖南省",
	"44": "广东省",
	"45": "广西壮族自治区",
	"46": "海南省",
	"50": "重庆市",
	"51": "四川省",
	"52": "贵州省",
	"53": "云南省",
	"54": "西藏自治区",
	"61": "陕西省",
	"62": "甘肃省",
	"63": "青海省",
	"64": "宁夏回族自治区",
	"65": "新疆维吾尔自治区",
	"71": "台湾省",
	"81": "香港特别行政区",
	"91": "澳门特别行政区",
}

// 效验18位身份证
func isValidCitizenNo18(citizenNo18 []byte) bool {
	nLen := len(citizenNo18)
	if nLen != 18 {
		return false

	}

	nSum := 0
	for i := 0; i < nLen-1; i++ {
		n, _ := strconv.Atoi(string((citizenNo18)[i]))
		nSum += n * weight[i]
	}
	mod := nSum % 11
	if valid_value[mod] == (citizenNo18)[17] {
		return true

	}
	return false
}
func IsLeapYear(nYear int) bool {
	if nYear <= 0 {
		return false

	}
	if (nYear%4 == 0 && nYear%100 != 0) || nYear%400 == 0 {
		return true

	}
	return false
}

// 生日日期格式效验
func checkBirthdayValid(nYear, nMonth, nDay int) bool {
	if nYear < 1900 || nMonth <= 0 || nMonth > 12 || nDay <= 0 || nDay > 31 {
		return false

	}

	curYear, curMonth, curDay := time.Now().Date()
	if nYear == curYear {
		if nMonth > int(curMonth) {
			return false

		} else if nMonth == int(curMonth) && nDay > curDay {
			return false

		}

	}

	if 2 == nMonth {
		if IsLeapYear(nYear) {
			if nDay > 29 {
				return false
			}
		} else {
			if nDay > 28 {
				return false
			}
		}

	} else if 4 == nMonth || 6 == nMonth || 9 == nMonth || 11 == nMonth {
		if nDay > 30 {
			return false

		}

	}

	return true
}

// 省份号码效验
func checkProvinceValid(citizenNo []byte) bool {
	provinceCode := make([]byte, 0)
	provinceCode = append(provinceCode, citizenNo[:2]...)
	provinceStr := string(provinceCode)

	for i := range valid_province {
		if provinceStr == i {
			return true

		}

	}

	return false
}

// 效验有效地身份证号码
func IsValidCitizenNo(citizenNo []byte) bool {
	if !isValidCitizenNo18(citizenNo) {
		return false
	}

	for i, v := range citizenNo {
		n, _ := strconv.Atoi(string(v))
		if n >= 0 && n <= 9 {
			continue

		}
		if v == 'X' && i == 16 {
			continue

		}
		return false

	}
	if !checkProvinceValid(citizenNo) {
		return false

	}
	nYear, _ := strconv.Atoi(string((citizenNo)[6:10]))
	nMonth, _ := strconv.Atoi(string((citizenNo)[10:12]))
	nDay, _ := strconv.Atoi(string((citizenNo)[12:14]))
	if !checkBirthdayValid(nYear, nMonth, nDay) {
		return false

	}
	return true
}
