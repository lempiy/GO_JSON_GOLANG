package main

import (
	"fmt"
	"github.com/lempiy/GoJSON/gojson"
	"time"
)

type Friend struct {
	Name string `json:"name"`
	Id   int
}

type someStr struct {
	Name    string   `json:"name" limit:"10"`
	Colors  []string `json:"colors"`
	Friends []Friend `json:"friends"`
}

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

	f := Friend{
		Name: "Simone",
		Id:   0,
	}
	f2 := Friend{
		Name: "Victor",
		Id:   1,
	}
	te := someStr{
		Name:    "Author",
		Colors:  []string{"red", "blue", "white"},
		Friends: []Friend{f, f2},
	}
	te2 := someStr{}
	re, err := gojson.SerializeStruct(te, false)
	gojson.ParseToStruct(&te2, re)
	fmt.Println("TE2", te2, err)
}
