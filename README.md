## GOJSON

GOJSON is an advanced version of JSON data tranfer format. It allows to use Golang-like tags inside JSON properties.

### Basic exmaple of gojson

```
{
    "name": "Joe" `"max-length": 4`,
    "sename": "Doe",
    "sister": {
        "name": "Jessy",
        "sename": "Doe"
    } `"editable": false`,
    "colors": [
        "red",
        "blue",
        "dark"
    ]
}
```

### Golang version

This is Golang lib for parsing/serializing gojson. Serializer may take fallowing structure as input:

```go
var input map[string]Node
```

or

```go
var input []Node
```

where

```go
type Node struct {
	Tag   string
	Value interface{}
}
```

or any custom struct in the same way as `encoding/json` does. In this case,
any tags of the struct except "json" will be used as gojson property tags.


##### Allowed methods

`SerializeStruct(s interface{}, bool) (string, error)`:

SerializeStruct serializes gojson string using any struct or []struct.
Similar to "encoding/json" package it will take json struct tag as a
key of json property if it exists. Also, it will ignore json tag value in
gojson tag serialization. So `json: "..."` will never be used in gojson.


`ParseAsArrayOrSlice(string) (map[string]Node, []Node, error)`:

Parses gojson by string returns map[string]Data{}||nil, []Data||nil in success and nil
or nil, nil, error if fails. Values in map or slice can be: Data (if value is primitive),
map[string]Data{} (if Value if JSON object {}), []Data{} if value is
JSON array and nil if value if JSON null.


`Serialize(interface{}, bool) (string, error)`

SerializeMap transforms map[string]Node into gojson string, trim parameter
responsible for turning on/off whitespacing inside json string.



##### Golang version is also [available](/)

