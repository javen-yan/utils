package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/yanzongzhen/DataFormatUtils/json"
	"github.com/yanzongzhen/utils"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const (
	NotEmpty     = "notEmpty"    //字符串不能为空
	IntMax       = "int-max"     //int最大值
	IntMin       = "int-min"     //int最小值
	Type         = "type"        //类型
	StrMaxLength = "str-max-len" //字符串最大长度
	StrMinLength = "str-min-len" //字符串最小长度
	StrLength    = "str-len"     //字符串长度
	Range        = "range"       //元素必须在合适的范围内 例:1-100
	Func         = "func"        //函数校验
)

const (
	ApplicationJson = "application/json"
)

type Verifier interface {
	Valid(param string) error
	CallName() string
}

type Option func(param string) error

func (o Option) Valid(param string) error {
	return o(param)
}

func (o Option) CallName() string {
	funcName := strings.Split(runtime.FuncForPC(reflect.ValueOf(o).Pointer()).Name(), ".")
	return funcName[len(funcName)-1]
}

func NewContext(r *http.Request, w http.ResponseWriter, params httprouter.Params) *Context {
	return &Context{
		Request:   r,
		Writer:    w,
		Ctx:       context.Background(),
		Params:    params,
		FuncChain: make(map[string]Verifier),
	}
}

type Context struct {
	Request   *http.Request
	Writer    http.ResponseWriter
	Params    httprouter.Params
	Ctx       context.Context
	FuncChain map[string]Verifier
}

func (c *Context) GetRequestBody() ([]byte, error) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

func (c *Context) Bind(ptr interface{}) error {
	//解析form表单
	if err := c.Request.ParseForm(); err != nil {
		return err
	}
	//验证必须为结构体指针
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr || reflect.ValueOf(ptr).Elem().Kind() != reflect.Struct {
		return errors.New("param must be struct pointer")
	}
	v := reflect.ValueOf(ptr).Elem()
	//遍历结构体
	for i := 0; i < v.NumField(); i++ {
		//结构体的字段
		fieldInfo := v.Type().Field(i)
		tag := fieldInfo.Tag
		name := tag.Get("param")
		//取json中的字段
		if strings.Index(name, ApplicationJson) != -1 {
			err := c.setJsonParam(name, v.Field(i), tag, fieldInfo.Name)
			if err != nil {
				return err
			}
			continue
		}
		//取form中的字段
		if name == "" {
			//默认使用结构体字段名称
			name = strings.ToLower(fieldInfo.Name)
		}
		//如果为空或未传，设置为默认值
		defaultValue := tag.Get("default")
		var value string
		if values, ok := c.Request.Form[name]; !ok {
			value = defaultValue
		} else {
			value = values[0]
			if utils.IsEmpty(value) {
				value = defaultValue
			}
		}
		//非空验证提前,解决Int在空情况下初始化为0的问题
		valid := tag.Get("valid")
		if strings.Index(valid, NotEmpty) != -1 {
			if value == "" {
				return errors.New(fieldInfo.Name + " value not empty")
			}
		}
		//绑定参数
		if err := c.populate(v.Field(i), value); err != nil {
			return fmt.Errorf("%s: %v", name, err)
		}

	}
	return nil
}

func (c *Context) setJsonParam(name string, field reflect.Value, tag reflect.StructTag, filedName string) error {
	valid := tag.Get("valid")
	paramType := tag.Get("path")
	defaultValue := tag.Get("default")
	if paramType == "all" {
		requestBody, _ := c.GetRequestBody()
		value := string(requestBody)
		if strings.Index(valid, NotEmpty) != -1 {
			if value == "" {
				return errors.New(filedName + " value not empty")
			}
		}
		err := c.populate(field, value)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(paramType, "$") {
		var value interface{}
		requestBody, _ := c.GetRequestBody()
		_ = json.TravelJsonData(requestBody, func(path string, v interface{}) bool {
			if path == paramType {
				value = v
				return true
			}
			return false
		})
		//非空校验
		if strings.Index(valid, NotEmpty) != -1 {
			//说明json中没有接口体对应的字段
			if value == nil {
				return errors.New(filedName + " value not empty")
			}
			//非空验证前提是json中该字段为字符串形式，其他形式必不为空
			switch reflect.TypeOf(value).Kind() {
			case reflect.String:
				v := value.(string)
				if utils.IsEmpty(v) {
					return errors.New(filedName + " value not empty")
				}
			}
		}
		if !utils.IsEmpty(defaultValue) {
			if value == nil {
				value = reflect.ValueOf(defaultValue).Interface()
			}
			if v, ok := value.(string); ok {
				if utils.IsEmpty(v) {
					value = reflect.ValueOf(defaultValue).Interface()
				}
			}
		}
		err := c.populateJson(field, value, valid, filedName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) populate(field reflect.Value, value string) error {
	var err error
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int:
		var i int64
		i, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			err = errors.New("param type not match,require:int")
		}
		field.SetInt(i)
	case reflect.Bool:
		b := false
		b, err = strconv.ParseBool(value)
		if err != nil {
			err = errors.New("param type not match,require:bool")
		}
		field.SetBool(b)
	case reflect.Float64:
		var floatV float64
		floatV, err = strconv.ParseFloat(value, 64)
		if err != nil {
			err = errors.New("param type not match,require:float")
		}
		field.SetFloat(floatV)
	default:
		return fmt.Errorf("unsupported kind %s", field.Type())
	}
	return err
}

func (c *Context) populateJson(field reflect.Value, value interface{}, valid string, filedName string) error {
	switch field.Kind() {
	case reflect.String:
		switch reflect.ValueOf(value).Kind() {
		case reflect.Float64:
			field.SetString(strconv.FormatFloat(value.(float64), 'E', -1, 64))
		case reflect.String:
			field.SetString(value.(string))
		case reflect.Bool:
			field.SetString(strconv.FormatBool(value.(bool)))
		default:
			return fmt.Errorf("unsupported convert type,expect type:%v,actual type:%v", field.Kind(), reflect.ValueOf(value).Kind())
		}
	case reflect.Float64:
		switch reflect.ValueOf(value).Kind() {
		case reflect.Float64:
			field.SetFloat(value.(float64))
		case reflect.String:
			floatV, err := strconv.ParseFloat(value.(string), 64)
			if err != nil {
				return fmt.Errorf("string convert to float64 error:%v", err)
			}
			field.SetFloat(floatV)
		default:
			return fmt.Errorf("unsupported convert type,expect type:%v,actual type:%v", field.Kind(), reflect.ValueOf(value).Kind())
		}
	case reflect.Int:
		switch reflect.ValueOf(value).Kind() {
		case reflect.Float64:
			field.SetInt(int64(value.(float64)))
		case reflect.String:
			intV, err := strconv.ParseInt(value.(string), 10, 64)
			if err != nil {
				return fmt.Errorf("int convert to string error:%v", err)
			}
			field.SetInt(intV)
		default:
			return fmt.Errorf("unsupported convert type,expect type:%v,actual type:%v", field.Kind(), reflect.ValueOf(value).Kind())
		}
	case reflect.Bool:
		switch reflect.ValueOf(value).Kind() {
		case reflect.Bool:
			field.SetBool(value.(bool))
		case reflect.String:
			b, err := strconv.ParseBool(value.(string))
			if err != nil {
				return fmt.Errorf("string convert to bool error:%v", err)
			}
			field.SetBool(b)
		default:
			return fmt.Errorf("unsupported convert type,expect type:%v,actual type:%v", field.Kind(), reflect.ValueOf(value).Kind())
		}
	default:
		return fmt.Errorf("unsupported kind %s", field.Type())
	}
	return nil
}

//对外暴露结构体验证函数
func (c *Context) Validate(ptr interface{}, option ...Option) error {
	for _, o := range option {
		c.Register(o)
	}
	//验证必须为结构体指针
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr || reflect.ValueOf(ptr).Elem().Kind() != reflect.Struct {
		return errors.New("param must be struct pointer")
	}
	fields := reflect.ValueOf(ptr).Elem()
	//NumField 遍历结构体每个字段
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Type().Field(i)
		//获取tag为valid的字符串
		valid := field.Tag.Get("valid")
		if valid == "" {
			continue
		}
		//field.Name 字段名称 value 字段值
		value := fields.FieldByName(field.Name)
		err := c.fieldValidate(field.Name, valid, value)
		if err != nil {
			return err
		}
	}
	return nil
}

//属性验证
func (c *Context) fieldValidate(fieldName, valid string, value reflect.Value) error {
	valids := strings.Split(valid, " ")
	for _, valid := range valids {
		if strings.Index(valid, Type) != -1 {
			v := value.Type().Name()
			split := strings.Split(valid, "=")
			t := split[1]
			if v != t {
				return errors.New(fieldName + " type must is " + t)
			}
		}

		if strings.Index(valid, IntMin) != -1 {
			v := value.Int()
			split := strings.Split(valid, "=")
			rule, err := strconv.Atoi(split[1])
			if err != nil {
				return errors.New(fieldName + ":验证规则有误")
			}
			if int(v) < rule {
				return errors.New(fieldName + " value must >= " + strconv.Itoa(rule))
			}
		}

		if strings.Index(valid, IntMax) != -1 {
			v := value.Int()
			split := strings.Split(valid, "=")
			rule, err := strconv.Atoi(split[1])
			if err != nil {
				return errors.New(fieldName + ":验证规则有误")
			}
			if int(v) > rule {
				return errors.New(fieldName + " value must <= " + strconv.Itoa(rule))
			}
		}
		if strings.Index(valid, Func) != -1 {
			v := value.String()
			if value.Kind() != reflect.String {
				return errors.New(fieldName + " type must be string if need function verification")
			}
			funcName := valid[strings.Index(valid, "(")+1 : strings.Index(valid, ")")]
			option := c.FuncChain[funcName]
			if err := option.Valid(v); err != nil {
				return err
			}
		}

		//字符串特殊处理
		if value.Type().Name() == "string" {
			if strings.Index(valid, StrLength) != -1 {
				v := value.String()
				split := strings.Split(valid, "=")
				lenStr := split[1]
				length, err := strconv.Atoi(lenStr)
				if err != nil {
					return errors.New(fieldName + " " + StrLength + " rule is error")
				}
				if len(v) != length {
					return errors.New(fieldName + " str length  must be " + lenStr)
				}
			}
			if strings.Index(valid, StrMaxLength) != -1 {
				v := value.String()
				split := strings.Split(valid, "=")
				lenStr := split[1]
				length, err := strconv.Atoi(lenStr)
				if err != nil {
					return errors.New(fieldName + " " + StrLength + " rule is error")
				}
				if len(v) > length {
					return errors.New(fieldName + " str length  <= " + lenStr)
				}
			}

			if strings.Index(valid, StrMinLength) != -1 {
				v := value.String()
				split := strings.Split(valid, "=")
				lenStr := split[1]
				length, err := strconv.Atoi(lenStr)
				if err != nil {
					return errors.New(fieldName + " " + StrLength + " rule is error")
				}
				if len(v) < length {
					return errors.New(fieldName + " str length  >= " + lenStr)
				}
			}
		}
	}
	return nil
}

func (c *Context) Register(option Option) {
	c.register(option)
}

func (c *Context) register(verifier Verifier) {
	if _, ok := c.FuncChain[verifier.CallName()]; !ok {
		c.FuncChain[verifier.CallName()] = verifier
	}
}
