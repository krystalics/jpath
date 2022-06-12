package jpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/**
    整体设计:
    jPath:=jpath.New("source") 构建最初的外围map
	jPath.find("path.p1.p2)    寻找嵌套的数据结构 value

*/

const (
	ARRAY_TYPE_PREFIX = "["
	ARRAY_TYPE_SUFFIX = "]"
	DEFAULT_SEPARATOR = "."
)

var (
	ARRAY_REGEX, _ = regexp.Compile(`\[-?\d+\]`)
)

type JPath struct {
	data      map[string]interface{}
	Separator string //路径分隔符默认是.
}

func New(source string) (*JPath, error) {
	return NewWithSep(source, DEFAULT_SEPARATOR)
}

func NewWithSep(source, sep string) (*JPath, error) {
	m, err := strToMap(source, false)
	if err != nil {
		return nil, err
	}

	return newJPath(m, sep)
}

func NewWithMap(m map[string]interface{}) (*JPath, error) {
	return newJPath(m, DEFAULT_SEPARATOR)
}

func NewWithMapAndSep(m map[string]interface{}, sep string) (*JPath, error) {
	return newJPath(m, sep)
}

func newJPath(data map[string]interface{}, sep string) (*JPath, error) {
	return &JPath{
		data:      data,
		Separator: sep,
	}, nil
}

//NewConcurrencySafe 提前将map递归构造好、这样在find的过程中就完全是读的过程了、线程安全！
func NewConcurrencySafe(source string) (*JPath, error) {
	jPath, err := New(source)
	if err != nil {
		return nil, err
	}
	deepRecursion(jPath.data)
	return jPath, nil
}

func deepRecursion(m map[string]interface{}) {
	for k, v := range m {
		switch v.(type) {
		case map[string]interface{}:
			deepRecursion(v.(map[string]interface{}))
		case string:
			m_sub, err := strToMap(v.(string), true)
			if err == nil {
				m[k] = m_sub
				deepRecursion(m_sub)
			} else {
				m[k] = v
			}
		case []interface{}:
			arr := make([]map[string]interface{}, 0)
			ori := make([]interface{}, 0)
			for _, val := range v.([]interface{}) {
				switch val.(type) {
				case map[string]interface{}:
					arr = append(arr, val.(map[string]interface{}))
				case string:
					m_sub, err := strToMap(val.(string), true)
					if err != nil {
						ori = append(ori, val.(string))
					} else {
						arr = append(arr, m_sub)
					}
				default:
					ori = append(ori, val)
				}
			}
			if len(arr) > 0 {
				m[k] = arr
				for _, v := range arr {
					deepRecursion(v)
				}
			} else {
				m[k] = ori
			}
		default:

		}
	}
}

//Find 除了在 NewConcurrencySafe的构建中、并发安全
// 其他的New方法构造出来的由于需要写map、所以并不安全
func (jp *JPath) Find(path string) (interface{}, error) {
	keys := strings.Split(path, jp.Separator)
	index := len(keys) - 1 //路径的尽头

	m := jp.data
	for i, v := range keys {
		if m == nil {
			return nil, errors.New(fmt.Sprintf("path %+v not found ", path))
		}
		//数组判断
		if !strings.HasSuffix(v, ARRAY_TYPE_SUFFIX) {
			data := m[v]
			if i == index {
				if data == nil {
					return nil, errors.New(fmt.Sprintf("path %+v not found ", path))
				}
				return data, nil
			}
			switch data.(type) {
			case map[string]interface{}:
				m = data.(map[string]interface{})
			case string:
				m_sub, _ := strToMap(m[v].(string), false)
				m[v] = m_sub
				m = m_sub
			case []interface{}:
				return nil, errors.New(fmt.Sprintf("key %+v error,array should be like this [index]", path))

			default:
				return nil, errors.New(fmt.Sprintf("key %+v error,it is not string or map", v))
			}
		} else {
			//对于数组需要特殊处理
			flag := ARRAY_REGEX.FindString(v)
			value, err := strconv.Atoi(flag[1 : len(flag)-1])
			if err != nil {
				return nil, errors.New(fmt.Sprintf("illegal index args %+v", flag))
			}
			key := strings.Split(v, flag)[0]
			data := m[key]
			switch data.(type) {
			case []interface{}:
				arr := data.([]interface{})
				if value > len(arr)-1 {
					return nil, errors.New(fmt.Sprintf("key %+v error,it is larger than array last index %+v", v, len(arr)-1))
				}
				tmp := arr[value]
				if i == index {
					if tmp == nil {
						return nil, errors.New(fmt.Sprintf("path %+v not found ", path))
					}
					return tmp, nil
				}
				switch tmp.(type) {
				case map[string]interface{}:
					m = tmp.(map[string]interface{})
				case string:
					m_sub, _ := strToMap(tmp.(string), false)
					m[v] = m_sub
					m = m_sub
				default:
					return nil, errors.New(fmt.Sprintf("key %+v error,it is not string or map", v))
				}
				//在完全递归场景下可能出现的类型
			case []map[string]interface{}:

				arr := data.([]map[string]interface{})
				if value > len(arr)-1 {
					return nil, errors.New(fmt.Sprintf("key %+v error,it is larger than array last index %+v", v, len(arr)-1))
				}
				tmp := arr[value]
				if i == index {
					if tmp == nil {
						return nil, errors.New(fmt.Sprintf("path %+v not found ", path))
					}
					return tmp, nil
				}

				m = tmp

			default:
				return nil, errors.New(fmt.Sprintf("key %+v error,it is not a array", v))

			}
		}
	}

	return nil, errors.New(fmt.Sprintf("path %+v not found ", path))
}

//isDeep 表示并发安全需要的完全递归、所以当json转换失败后、不再打印日志
func strToMap(str string, isDeep bool) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(str), &m)
	if err != nil {
		if isDeep {
			return nil, err
		}
		return nil, err
	}
	return m, nil
}

//FindString find the value and not process the error
func (jp *JPath) FindString(path string) string {
	value, _ := jp.Find(path)
	if value != nil {
		return value.(string)
	}
	return ""
}

func (jp *JPath) FindInt64(path string) int64 {
	value, _ := jp.Find(path)
	if value != nil {
		return value.(int64)
	}
	return 0
}

func (jp *JPath) FindFloat64(path string) float64 {
	value, _ := jp.Find(path)
	if value != nil {
		return value.(float64)
	}

	return 0.0
}
