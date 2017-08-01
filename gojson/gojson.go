package gojson

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"reflect"
)

type Node struct {
	Tag   string
	Value interface{}
}

type mapData struct {
	InQuotes     bool
	Key          []byte
	InKey        bool
	Value        interface{}
	InValue      bool
	Tag          []byte
	InTag        bool
	PrevChar     byte
	AfterCol     bool
	Type         string
	AfterClosing bool
}
// SerializeStruct serializes gojson string using any struct or []struct.
// Similar to "encoding/json" package it will take json struct tag as a
// key of json property if it exists. Also, it will ignore json tag value in
// gojson tag serialization. So `json: "..."` will never be used in gojson.
func SerializeStruct(s interface{}, trim bool) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	stucV := reflect.ValueOf(s)
	r, err := getNode(s, stucV)
	if err != nil {
		return "", err
	}

	return Serialize(r.Value, trim)
}

func getMapFromStruct(s interface{}) (map[string]Node, error) {
	result := make(map[string]Node)
	stucV := reflect.ValueOf(s)
	stucT := reflect.TypeOf(s)
	n := stucT.NumField()

	for i := 0; i < n; i ++ {
		var err error
		field := stucT.Field(i)
		jsonTag := field.Tag.Get("json")
		tag := string(field.Tag)
		name := field.Name
		if jsonTag != "" {
			name = jsonTag
			tag = strings.TrimSpace(
				strings.Replace(tag, `json:`+`"`+ name +`"`, "", -1),
			)
		}
		value := stucV.Field(i).Interface()
		stucVV := reflect.ValueOf(value)
		node, err := getNode(value, stucVV)
		if err != nil {
			return nil, err
		}
		if tag != "" {
			node.Tag = tag
		}
		result[name] = node
	}
	return result, nil
}

func getSlice(items []interface{}) ([]Node, error) {
	result := []Node{}
	for _, item := range items {
		stucV := reflect.ValueOf(item)
		node, err := getNode(item, stucV)
		if err != nil {
			return nil, err
		}
		result = append(result, node)
	}

	return result, nil
}

func getNode(item interface{}, stucV reflect.Value) (Node, error) {
	v := Node{}
	var err error

	switch stucV.Kind() {
	case reflect.Struct:
		v.Value, err = getMapFromStruct(item)
		if err != nil {
			return v, err
		}
	case reflect.Slice:
		val := interfaceSlice(item)
		v.Value, err = getSlice(val)
		if err != nil {
			return v, err
		}
	case reflect.Map:
		val := interfaceMap(item)
		v.Value, err = getMap(val)
		if err != nil {
			return v, errors.New("Only map[string]interface{} is acceptable")
		}
	default:
		v.Value = item
	}
	return v, nil
}

func interfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("Non-slice value in interfaceSlice argument")
	}

	ret := make([]interface{}, s.Len())

	for i:=0; i<s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

func interfaceMap(m interface{}) map[string]interface{} {
	s := reflect.ValueOf(m)
	if s.Kind() != reflect.Map {
		panic("Non-map value in interfaceSlice argument")
	}

	ret := make(map[string]interface{}, s.Len())

	keys := s.MapKeys()

	for _, key := range keys {
		value := s.MapIndex(key)
		ret[key.String()] = value.Interface()
	}

	return ret
}

func getMap(items map[string]interface{}) (map[string]Node, error) {
	result := make(map[string]Node)

	for key, item := range items {
		stucV := reflect.ValueOf(item)
		node, err := getNode(item, stucV)
		if err != nil {
			return nil, err
		}
		result[key] = node
	}

	return result, nil
}


// Parses gojson by string returns map[string]Data{}||nil, []Data||nil in success and nil
// or nil, nil, error if fails. Values in map or slice can be: Data (if value is primitive),
// map[string]Data{} (if Value if JSON object {}), []Data{} if value is
// JSON array and nil if value if JSON null.
func ParseAsArrayOrSlice(str string) (map[string]Node, []Node, error) {
	if strings.HasPrefix(str, "null") {
		return nil, nil, nil
	} else if str[0] == '{' {
		result, _, err := parseAsMap([]byte(str), 1)
		return result, nil, err
	} else if str[0] == '[' {
		result, _, err := parseAsSlice([]byte(str), 1)
		return nil, result, err
	}
	return nil, nil, syntaxError(0, str[0])
}

func parseAsMap(str []byte, c int) (map[string]Node, int, error) {
	node := map[string]Node{}
	m := mapData{}
	var err error

	//we can do byte iteration because special JSON chars has always 1 byte length
	for c < len(str) {
		if str[c] == '"' && !m.InTag {
			if !m.InKey && len(m.Key) > 0 && !m.AfterCol {
				return nil, c, syntaxError(c, str[c])
			}
			if m.PrevChar == '\\' {
				if m.InKey {
					m.Key = append(m.Key, str[c])
				} else if m.InValue {
					if m.Value == nil {
						m.Value = []byte{str[c]}
					} else {
						v := (m.Value).([]byte)
						m.Value = append(v, str[c])
					}
				}
			}
			if m.Value != nil {
				m.InValue = false
			} else if len(m.Key) != 0 {
				if m.InKey {
					m.InKey = false
				} else {
					m.InValue = true
					m.Type = "string"
				}
			} else {
				m.InKey = true
			}
			m.InQuotes = !m.InQuotes
		} else if !m.InQuotes && !m.InTag {
			if str[c] == '`' {
				if m.Key != nil {
					m.InTag = true
				} else {
					return nil, c, syntaxError(c, str[c])
				}
			} else if str[c] == '}' {
				if m.Key != nil {
					m.AfterClosing = true
					if m.Type == "" {
						m.Type, err = detectType(&m)
						if err != nil {
							return nil, c, err
						}
					}
					err := createPair(node, &m)
					if err != nil {
						return nil, c, err
					}
				} else {
					return nil, c, syntaxError(c, str[c])
				}
				return node, c, nil
			} else if str[c] == '{' {
				c++
				m.Value, c, err = parseAsMap(str, c)
				if err != nil {
					return nil, c, err
				}
				m.Type = "map"
			} else if str[c] == '[' {
				c++
				m.Value, c, err = parseAsSlice(str, c)
				m.Type = "slice"
			} else if str[c] == ':' {
				m.AfterCol = true
			} else if str[c] == ',' {
				if m.Type == "" {
					m.Type, err = detectType(&m)
					if err != nil {
						return nil, c, err
					}
				}
				err := createPair(node, &m)
				if err != nil {
					return nil, c, err
				}
				reset(&m)
			} else if str[c] == '`' {
				m.InTag = !m.InTag
			} else {
				r, _ := utf8.DecodeRune([]byte{str[c]})
				if !unicode.IsSpace(r) {
					if m.InValue {
						if m.Value == nil {
							m.Value = []byte{str[c]}
						} else {
							v := (m.Value).([]byte)
							m.Value = append(v, str[c])
						}
					} else if m.Key != nil && !m.InValue && m.AfterCol && m.Value == nil {
						m.InValue = true
						if m.Value == nil {
							m.Value = []byte{str[c]}
						} else {
							v := (m.Value).([]byte)
							m.Value = append(v, str[c])
						}
					} else {

						return nil, c, syntaxError(c, str[c])
					}
				} else {
					if m.InValue {
						m.InValue = false
					}
				}
			}
		} else if m.InTag {
			if str[c] == '`' {
				m.InTag = false
			} else {
				m.Tag = append(m.Tag, str[c])
			}
		} else {
			if str[c] == '\\' {
				m.PrevChar = '\\'
			} else {
				m.PrevChar = 0
			}
			if m.InKey {
				m.Key = append(m.Key, str[c])
			} else if m.InValue {
				if m.Value == nil {
					m.Value = []byte{str[c]}
				} else {
					v := (m.Value).([]byte)
					m.Value = append(v, str[c])
				}
			} else {
				return nil, c, syntaxError(c, str[c])
			}
		}
		c++
	}

	if m.AfterClosing {
		if m.Key != nil {
			createPair(node, &m)
			reset(&m)
		}
		return node, c, nil
	} else {
		return nil, 0, errors.New("Invalid JSON")
	}
	return node, c, nil
}

func detectType(m *mapData) (string, error) {
	bytes := (m.Value).([]byte)
	val := string(bytes)
	if val == "" {
		return "", errors.New("Syntax error")
	}
	if v, isOk := strconv.ParseFloat(val, 64); isOk == nil {
		if isIntegral(v) {
			return "int", nil
		}
		return "float64", nil
	}
	if isBoolean(val) {
		return "bool", nil
	}
	return "string", nil
}

func createPair(node map[string]Node, m *mapData) error {
	pair := Node{
		Tag: string(m.Tag),
	}
	if m.Type == "map" || m.Type == "slice" {
		pair.Value = m.Value
		node[string(m.Key)] = pair
		return nil
	}
	bytes := (m.Value).([]byte)
	val := string(bytes)
	switch m.Type {
	case "int":
		v, _ := strconv.ParseFloat(val, 64)
		pair.Value = int(v)
		node[string(m.Key)] = pair
	case "float64":
		v, _ := strconv.ParseFloat(val, 64)
		pair.Value = v
		node[string(m.Key)] = pair
	case "bool":
		if val == "true" {
			pair.Value = true
			node[string(m.Key)] = pair
		} else {
			pair.Value = false
			node[string(m.Key)] = pair
		}
	case "string":
		pair.Value = val
		node[string(m.Key)] = pair
	default:
		return errors.New("Wrong JSON syntax")
	}
	return nil
}

func parseAsSlice(str []byte, c int) ([]Node, int, error) {
	node := []Node{}
	m := mapData{}
	var err error

	//we can do byte iteration because special JSON chars has always 1 byte length
	for c < len(str) {
		if str[c] == '"' && !m.InTag {
			if m.PrevChar == '\\' {
				if m.InValue {
					if m.Value == nil {
						m.Value = []byte{str[c]}
					} else {
						v := (m.Value).([]byte)
						m.Value = append(v, str[c])
					}
				}
			}
			if m.Value != nil {
				if !m.InQuotes {
					return nil, c, syntaxError(c, str[c])
				}
				m.InValue = false
			} else {
				m.InValue = true
				m.Type = "string"
			}
			m.InQuotes = !m.InQuotes
		} else if !m.InQuotes && !m.InTag {
			if str[c] == '`' {
				if m.Value != nil || m.Type == "string" {
					m.InTag = true
				} else {
					return nil, c, syntaxError(c, str[c])
				}
			} else if str[c] == ']' {
				if m.Value != nil || m.Type == "string" {
					m.AfterClosing = true
					if m.Type == "" {
						m.Type, err = detectType(&m)
						if err != nil {
							return nil, c, err
						}
					}
					pair, err := createValue(&m)
					if err != nil {
						return nil, c, err
					}
					node = append(node, pair)
				} else {
					return nil, c, syntaxError(c, str[c])
				}
				return node, c, nil
			} else if str[c] == '{' {
				c++
				m.Value, c, err = parseAsMap(str, c)
				if err != nil {
					return nil, c, err
				}
				m.Type = "map"
			} else if str[c] == '[' {
				c++
				m.Value, c, err = parseAsSlice(str, c)
				if err != nil {
					return nil, c, err
				}
				m.Type = "slice"
			} else if str[c] == ':' {
				m.AfterCol = true
			} else if str[c] == ',' {
				if m.Type == "" {
					m.Type, err = detectType(&m)
					if err != nil {
						return nil, c, err
					}
				}
				pair, err := createValue(&m)
				if err != nil {
					return nil, c, err
				}
				node = append(node, pair)
				reset(&m)
			} else if str[c] == '`' {
				m.InTag = !m.InTag
			} else {
				r, _ := utf8.DecodeRune([]byte{str[c]})
				if !unicode.IsSpace(r) {
					if m.InValue && m.Value == nil {
						m.InValue = true
					} else {
						return nil, c, syntaxError(c, str[c])
					}
					if m.Value == nil {
						m.Value = []byte{str[c]}
					} else {
						v := (m.Value).([]byte)
						m.Value = append(v, str[c])
					}
				} else {
					if m.InValue {
						m.InValue = false
					}
				}
			}
		} else if m.InTag {
			if str[c] == '`' {
				m.InTag = false
			} else {
				m.Tag = append(m.Tag, str[c])
			}
		} else {
			if str[c] == '\\' {
				m.PrevChar = '\\'
			} else {
				m.PrevChar = 0
			}
			if m.InValue {
				if m.Value == nil {
					m.Value = []byte{str[c]}
				} else {
					v := (m.Value).([]byte)
					m.Value = append(v, str[c])
				}
			}
		}
		c++
	}

	if m.AfterClosing {
		if m.Key != nil {
			pair, err := createValue(&m)
			if err != nil {
				return nil, c, err
			}
			node = append(node, pair)
			reset(&m)
		}
		return node, c, nil
	} else {
		return nil, 0, errors.New("Invalid JSON")
	}
	return node, c, nil
}

func createValue(m *mapData) (Node, error) {
	pair := Node{
		Tag: string(m.Tag),
	}
	if m.Type == "map" || m.Type == "slice" {
		pair.Value = m.Value
		return pair, nil
	}
	bytes := (m.Value).([]byte)
	val := string(bytes)
	switch m.Type {
	case "int":
		v, _ := strconv.ParseFloat(val, 64)
		pair.Value = int(v)
	case "float64":
		v, _ := strconv.ParseFloat(val, 64)
		pair.Value = v
	case "bool":
		if val == "true" {
			pair.Value = true
		} else {
			pair.Value = false
		}
	case "string":
		pair.Value = val
	default:
		return pair, errors.New("Wrong JSON syntax")
	}
	return pair, nil
}

func isIntegral(val float64) bool {
	return val == float64(int(val))
}

func isBoolean(val string) bool {
	return val == "true" || val == "false"
}

func syntaxError(c int, crc byte) error {
	return errors.New(fmt.Sprintf("Syntax error at %d - char %s", c, string(crc)))
}

func reset(data *mapData) {
	*data = mapData{}
}

//SerializeMap transforms map[string]Node into gojson string, trim parameter
//responsible for turning on/off whitespacing inside json string.
func Serialize(m interface{}, trim bool) (string, error) {
	config := serializeConfig{
		Trim:       trim,
		BasicSpace: 4,
	}
	switch v := m.(type) {
	case map[string]Node:
		return serializeMap(v, config, 0)
	case []Node:
		return serializeSlice(v, config, 0)
	default:
		return "", errors.New(`Error upon serialization - wrong input type`)
	}
}

type serializeConfig struct {
	Trim       bool
	BasicSpace int
}

func serializeMap(m map[string]Node, c serializeConfig, ns int) (string, error) {
	i := 0
	result := "{"
	row := ""
	if !c.Trim {
		result += "\n"
		ns += c.BasicSpace
	}
	for key, node := range m {
		switch v := node.Value.(type) {
		case map[string]Node:
			value, err := serializeMap(v, c, ns)
			if c.Trim {
				row = fmt.Sprintf(`"%s":%s`, key, value)
			} else {
				row = strings.Repeat(" ", ns) + row
				row += fmt.Sprintf(`"%s": %s `, key, value)
			}
			if node.Tag != "" {
				if !c.Trim {
					row += fmt.Sprintf(" `%s`", node.Tag)
				} else {
					row += fmt.Sprintf("`%s`", node.Tag)
				}
			}
			if err != nil {
				return "", err
			}
			if i != len(m)-1 {
				row += ","
			}
		case []Node:
			value, err := serializeSlice(v, c, ns)
			if c.Trim {
				row = fmt.Sprintf(`"%s":%s`, key, value)
			} else {
				row = strings.Repeat(" ", ns) + row
				row += fmt.Sprintf(`"%s": %s`, key, value)
			}
			if node.Tag != "" {
				if !c.Trim {
					row += fmt.Sprintf(" `%s`", node.Tag)
				} else {
					row += fmt.Sprintf("`%s`", node.Tag)
				}
			}
			if err != nil {
				return "", err
			}
			if i != len(m)-1 {
				row += ","
			}
		default:
			row = createRow(key, node, c.Trim)
			if !c.Trim {
				row = strings.Repeat(" ", ns) + row
			}
			if i != len(m)-1 {
				row += ","
			}
		}
		if !c.Trim {
			row += "\n"
		}
		result += row
		row = ""
		i++
	}
	if !c.Trim {
		result += strings.Repeat(" ", ns-c.BasicSpace) + "}"
	} else {
		result += "}"
	}
	return result, nil
}

func serializeSlice(m []Node, c serializeConfig, ns int) (string, error) {
	i := 0
	result := "["
	row := ""
	if !c.Trim {
		result += "\n"
		ns += c.BasicSpace
	}
	for _, node := range m {
		switch v := node.Value.(type) {
		case map[string]Node:
			value, err := serializeMap(v, c, ns)
			if c.Trim {
				row = fmt.Sprintf(`%s`, value)
			} else {
				row = strings.Repeat(" ", ns) + row
				row += fmt.Sprintf(`%s`, value)
			}
			if err != nil {
				return "", err
			}
			if i != len(m)-1 {
				row += ","
			}
		case []Node:
			value, err := serializeSlice(v, c, ns)
			if c.Trim {
				row = fmt.Sprintf(`%s`, value)
			} else {
				row = strings.Repeat(" ", ns) + row
				row += fmt.Sprintf(`%s`, value)
			}
			if err != nil {
				return "", err
			}
			if i != len(m)-1 {
				row += ","
			}
		default:
			row = createElement(node, c.Trim)
			if !c.Trim {
				row = strings.Repeat(" ", ns) + row
			}
			if i != len(m)-1 {
				row += ","
			}
		}
		if !c.Trim {
			row += "\n"
		}
		result += row
		row = ""
		i++
	}
	if !c.Trim {
		result += strings.Repeat(" ", ns-c.BasicSpace) + "]"
	} else {
		result += "]"
	}
	return result, nil
}

func createRow(key string, node Node, trim bool) string {
	value := getValue(node.Value)
	if node.Tag != "" {
		if trim {
			value = fmt.Sprintf(`%s`+"`%s`", value, node.Tag)
		} else {
			value = fmt.Sprintf(`%s `+"`%s`", value, node.Tag)
		}
	}

	if trim {
		return fmt.Sprintf(`"%s":%s`, key, value)
	}

	return fmt.Sprintf(`"%s": %s`, key, value)
}

func createElement(node Node, trim bool) string {
	value := getValue(node.Value)
	if node.Tag != "" {
		if trim {
			value = fmt.Sprintf(`%s`+"`%s`", value, node.Tag)
		} else {
			value = fmt.Sprintf(`%s `+"`%s`", value, node.Tag)
		}
	}

	return fmt.Sprintf(`%s`, value)
}

func getValue(val interface{}) string {
	needQuotes := false
	value := ""
	switch v := val.(type) {
	case time.Time:
		if v.String()[:4] == "0000" {
			value = v.Format("15:04:05")
		} else {
			value = v.Format("2006-01-02 15:04:05")
		}
		needQuotes = true
	case int:
		value = fmt.Sprintf("%d", v)
	case int32:
		value = fmt.Sprintf("%d", v)
	case int64:
		value = fmt.Sprintf("%d", v)
	case float32:
		value = fmt.Sprintf("%v", v)
	case float64:
		value = fmt.Sprintf("%v", v)
	case nil:
		value = "null"
	case string:
		//sqlite interprets BIT type as a string so we need an explicit conversation
		if _, err := strconv.ParseBool(fmt.Sprint(value)); err == nil {
			value = fmt.Sprintf("%s", v)
			break
		}
		value = strings.Replace(v, `"`, `\"`, -1)
		needQuotes = true
	default:
		value = fmt.Sprintf("%s", v)
	}
	if needQuotes {
		value = fmt.Sprintf(`"%s"`, value)
	}
	return value
}
