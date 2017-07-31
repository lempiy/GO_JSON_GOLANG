package main

import (
	"fmt"
	"github.com/lempiy/GoJSON/gojson"
	"time"
)

func main() {
	jessey := gojson.Node{
		Value: "Jessey",
		Tag:   `"unique": true`,
	}
	jid := gojson.Node{
		Value: 324.54,
		Tag:   `"number": < 400`,
	}
	colors := gojson.Node{
		Value: []gojson.Node{
			gojson.Node{Value: "red"},
			gojson.Node{Value: "blue"},
		},
		Tag: `"list": ["red", "blue", "green"]`,
	}
	m := make(map[string]gojson.Node)
	m["name"] = gojson.Node{
		Tag:   `"editable": false`,
		Value: "John",
	}
	m["created"] = gojson.Node{
		Tag:   `"editable": false`,
		Value: time.Now(),
	}
	m["id"] = gojson.Node{
		Value: 5,
	}
	m["colors"] = colors
	m["sister"] = gojson.Node{
		Value: map[string]gojson.Node{
			"name":   jessey,
			"_id":    jid,
			"colors": colors,
		},
	}
	r, err := gojson.Serialize(m, false)
	fmt.Println(r, err)
	s, err := gojson.Serialize([]gojson.Node{
		gojson.Node{Value: "red"},
		gojson.Node{Value: "blue"},
	}, false)
	fmt.Println(s, err)
}
