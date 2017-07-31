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


##### Golang version is also [available](/)

