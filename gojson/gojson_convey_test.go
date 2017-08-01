package gojson

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
	"strings"
)

func TestConveyParseAsArrayOrSlice(t *testing.T) {
	Convey("Parsing to Slice and Map", t, func() {
		testJSON := `{
			"testing": 1 `+"`tag:\"custom\"`"+`,
			"name": "John Doe",
			"isActive": true,
			"colors": [
				"red",
				"blue",
				"green"
			]`+ "`list: [\"red\", \"blue\", \"green\"]`" +`}`
		testArray := `[
			{
				"name": "nickname",
				"id": 0 `+"`unique:true`"+`
			},
			{
				"name": "nickname2",
				"id": 1 `+"`unique:true`"+`
			}
		]`
		wrongJSON := `{
			"testing": mistake 1 `+"`tag:\"custom\"`"+`,
			"name": "John Doe",
			"isActive": true
			"colors": [
				"red",
				"blue",
				"green"
			]`+ "`list: [\"red\", \"blue\", \"green\"]`" +`}`
		cmp, _, cerr := ParseAsArrayOrSlice(testJSON)
		_, asl, _ := ParseAsArrayOrSlice(testArray)
		_, _, werr := ParseAsArrayOrSlice(wrongJSON)

		Convey("When gojson string is wrong should return error", func() {
			So(werr, ShouldNotBeNil)
		})

		Convey("When gojson string is wrong shouldn't return error", func() {
			So(cerr, ShouldBeNil)
		})

		Convey("Should return correct data structure depending on input", func() {
			So(cmp, ShouldNotBeNil)
			So(asl, ShouldNotBeNil)
		})

		Convey("Should contain correct data on primitives", func() {
			testing, isOk := cmp["testing"]
			val := (testing.Value).(int)
			So(isOk, ShouldBeTrue)
			So(val, ShouldEqual, 1)
		})

		Convey("Should contain correct data on complex structures", func() {
			colors, isOk := cmp["colors"]
			So(isOk, ShouldBeTrue)
			val, ok := (colors.Value).([]Node)
			So(ok, ShouldBeTrue)
			col, ok := (val[0].Value).(string)
			So(ok, ShouldBeTrue)
			So(col, ShouldEqual, "red")
		})

		Convey("Should contain tags", func() {
			testing := cmp["testing"]
			So(testing.Tag, ShouldEqual, "tag:\"custom\"")
			colors := cmp["colors"]
			So(colors.Tag, ShouldEqual, "list: [\"red\", \"blue\", \"green\"]")
		})
	})
}

func TestConveySerialize(t *testing.T) {
	Convey("Serializing from Slice and Map", t, func() {
		jessey := Node{
			Value: "Jessey",
			Tag:   `"unique": true`,
		}
		jid := Node{
			Value: 324.54,
			Tag:   `"number": < 400`,
		}
		colors := Node{
			Value: []Node{
				Node{Value: "red"},
				Node{Value: "blue"},
			},
			Tag: `"list": ["red", "blue", "green"]`,
		}
		m := make(map[string]Node)
		m["name"] = Node{
			Tag:   `"editable": false`,
			Value: "John",
		}
		m["created"] = Node{
			Tag:   `"editable": false`,
			Value: time.Now().Format(time.UnixDate),
		}
		m["id"] = Node{
			Value: 5,
		}
		m["colors"] = colors
		m["sister"] = Node{
			Value: map[string]Node{
				"name":   jessey,
				"_id":    jid,
				"colors": colors,
			},
		}
		r, err1 := Serialize(m, true)

		_, err2 := Serialize([]Node{
			Node{Value: "red"},
			Node{Value: "blue"},
		}, false)


		Convey("When data is correct it shouldn't return error", func() {
			So(err1, ShouldBeNil)
			So(err2, ShouldBeNil)
		})

		Convey("Should return a valid gojson string", func() {
			_, _, err := ParseAsArrayOrSlice(r)
			So(err, ShouldBeNil)
		})

		Convey("Should have particular values", func() {
			tagIndex := strings.Index(r, "`\"editable\": false`")
			So(tagIndex, ShouldBeGreaterThan, -1)
			valueKeyIndex := strings.Index(r, `"name":"John"`)
			So(valueKeyIndex, ShouldBeGreaterThan, -1)
			valueKeyTagIndex := strings.Index(r, `"colors":["red","blue"]`+
				"`\"list\": [\"red\", \"blue\", \"green\"]`")
			So(valueKeyTagIndex, ShouldBeGreaterThan, -1)
		})
	})
}

func TestConveySerializeStruct(t *testing.T) {
	Convey("Serializing from Struct", t, func() {
		type Friend struct {
			Name string `json:"name"`
			Id int
		}

		type someStr struct {
			Name string `json:"name" limit:"10"`
			Colors []string `json:"colors"`
			Friends []Friend `json:"friends"`
		}

		f := Friend{
			Name: "Simone",
			Id: 0,
		}
		f2 := Friend{
			Name: "Victor",
			Id: 1,
		}
		te := someStr{
			Name: "Author",
			Colors: []string {"red", "blue", "white"},
			Friends: []Friend{f, f2},
		}

		re, err := SerializeStruct(te, true)

		Convey("When data is correct it shouldn't return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("Should return a valid gojson string", func() {
			_, _, err := ParseAsArrayOrSlice(re)
			So(err, ShouldBeNil)
		})

		Convey("Should have particular values", func() {
			tagIndex := strings.Index(re, "`limit:\"10\"`")
			So(tagIndex, ShouldBeGreaterThan, -1)
			valueKeyIndex := strings.Index(re, `"name":"Simone"`)
			So(valueKeyIndex, ShouldBeGreaterThan, -1)
			valueKeyTagIndex := strings.Index(re, `"name":"Author"`+
				"`limit:\"10\"`")
			So(valueKeyTagIndex, ShouldBeGreaterThan, -1)
		})
	})
}