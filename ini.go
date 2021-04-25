package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

// MysqlConfig
type MysqlConfig struct {
	Address  string `ini:"address"`
	Port     int    `ini:"port"`
	Username string `ini:"username"`
	Passworl string `ini:"password"`
}

// RedisConfig
type RedisConfig struct {
	Host     string `ini:"host"`
	Port     int    `ini:"port"`
	Password int    `ini:"password"`
	Database int    `ini:"database"`
	Test     bool   `ini:"test"`
}

// Config 结构体嵌套
type Config struct {
	MysqlConfig `ini:"mysql"`
	RedisConfig `ini:"redis"`
}

func loadIni(fileName string, data interface{}) (err error) {
	//0 参数校验
	// 0.1 data参数必须是指针
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Ptr {
		err = errors.New("data 类型错误") //创建一个错误
		return
	}
	//0.2 data参数 必须是结构体类型指针
	if t.Elem().Kind() != reflect.Struct {
		err = errors.New("data 值类型错误")
		return
	}
	//1 读文件的字节类型
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	//string(b) //转字符串
	lineSlice := strings.Split(string(b), "\r\n")
	//fmt.Printf("%#v\n", lineSlice) //[]string{"; mysql config", "[mysql]", "address=10.20.30.40", "port=3306", "username=root", "password=root", "", "# redis config", "[redis]", "host=127.0.0.1", "port=6379", "datebase=0"}
	var structName string
	for index, line := range lineSlice {
		line = strings.TrimSpace(line) //去掉空格
		//如果是空行 直接跳过
		if len(line) == 0 {
			continue
		}
		//2 逐行读取数据
		//2.1 跳过注释
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		//2.2 []表示开头
		if strings.HasPrefix(line, "[") {

			if line[0] != '[' || line[len(line)-1] != ']' {
				err = fmt.Errorf("line:%d 错误", index+1)
				return
			}
			//把这行首尾[]去掉 取中间内容 把空格去掉
			sectionName := strings.TrimSpace(line[1 : len(line)-1])
			if len(sectionName) == 0 {
				err = fmt.Errorf("line:%d 错误", index+1)
				return
			}
			//根据字符串 找到结构体
			for i := 0; i < t.Elem().NumField(); i++ {
				field := t.Elem().Field(i)
				if sectionName == field.Tag.Get("ini") {
					//说明找到嵌套结构体 加下字段名\
					structName = field.Name
					//fmt.Printf("%v - %v\n", structName, sectionName)
					break
				}
			}

		} else {
			//2.3 非[]开头 分隔键值对
			if strings.Index(line, "=") == -1 || strings.HasPrefix(line, "=") {
				err = fmt.Errorf("line:%d 错误", index+1)
				return
			}
			indexs := strings.Index(line, "=")
			key := strings.TrimSpace(line[:indexs])
			value := strings.TrimSpace(line[indexs+1:])
			//2.4 key=value
			v := reflect.ValueOf(data)
			structObj := v.Elem().FieldByName(structName) //拿到嵌套体中值信息
			sTyep := structObj.Type()                     //拿到嵌套体中的结构体类型
			if sTyep.Kind() != reflect.Struct {
				err = fmt.Errorf("data 中的%s字段应该是结构体", structName)
				return
			}
			var fieldName string
			var fileType reflect.StructField
			//变量结构体 获取每个字段
			for i := 0; i < structObj.NumField(); i++ {
				field := sTyep.Field(i) //tag信息是存在类型信息中
				fileType = field
				if field.Tag.Get("ini") == key {
					//找到了对于字段 并储存
					fieldName = field.Name
					break
				}
			}
			//根据fieldName 这个字段进行赋值
			if len(fieldName) == 0 {
				//在结构体中找不到对于字符
				continue
			}

			fileObj := structObj.FieldByName(fieldName)
			//fmt.Println(fileObj)
			//对其赋值
			switch fileType.Type.Kind() {
			case reflect.String:
				fileObj.SetString(value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				var valueInt int64
				valueInt, err = strconv.ParseInt(value, 10, 64)
				if err != nil {
					err = fmt.Errorf("line%d 错误", index+1)
					return
				}

				fileObj.SetInt(valueInt)
				//return
			case reflect.Bool:
				var valueBool bool
				valueBool, err = strconv.ParseBool(value)
				if err != nil {
					err = fmt.Errorf("line%d 错误", index+1)
					return
				}
				fileObj.SetBool(valueBool)

			case reflect.Float32, reflect.Float64:
				var valueFloat float64
				valueFloat, err = strconv.ParseFloat(value, 64)
				if err != nil {
					err = fmt.Errorf("line%d 错误", index+1)
					return
				}
				fileObj.SetFloat(valueFloat)

			}

		}

	}

	return

}

func main() {
	var cfg Config
	err := loadIni("./conf.ini", &cfg)
	if err != nil {
		fmt.Printf("loadini 错误 err:%v\n", err)
		return
	}

	//fmt.Println(cfg.Address, cfg.Port, cfg.Username, cfg.Passworl)
	fmt.Println(cfg)         //{{10.20.30.40 3306 root root} {127.0.0.1 6379 0 0 false}}
	fmt.Printf("%#v\n", cfg) //main.Config{MysqlConfig:main.MysqlConfig{Address:"10.20.30.40", Port:3306, Username:"root", Passworl:"root"}, RedisConfig:main.RedisConfig{Host:"127.0.0.1", Port:6379, Password:0, Database:0, Test:false}}
}
