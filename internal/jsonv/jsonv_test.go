package jsonv

import (
	"errors"
	"strings"
	"testing"
)

func TestTraversal(t *testing.T) {
	j, err := ParseString(`{"data":{"list":[{"id":42,"name":"x","ratio":1.5,"ok":true,"none":null}],"count":"7"}}`)
	if err != nil {
		t.Fatal(err)
	}
	item := j.Path("data", "list").At(0)
	if i, _ := item.Get("id").Int(); i != 42 {
		t.Errorf("id = %d", i)
	}
	if s := item.Get("id").String(); s != "42" {
		t.Errorf("id string = %q", s)
	}
	if f, _ := item.Get("ratio").Float(); f != 1.5 {
		t.Errorf("ratio = %v", f)
	}
	if b, ok := item.Get("ok").Bool(); !b || !ok {
		t.Error("ok")
	}
	if !item.Get("none").IsNull() || !item.Has("none") {
		t.Error("none")
	}
	if c, _ := j.Path("data", "count").Int64(); c != 7 {
		t.Errorf("count = %d", c)
	}
	if !j.Path("missing", "deep").IsNull() || j.At(3).Exists() {
		t.Error("missing values must be null")
	}
	var missing *MissingKeyError
	if _, err := j.Require("nope"); !errors.As(err, &missing) || missing.Key != "nope" {
		t.Errorf("Require: %v", err)
	}
	if !strings.Contains(j.Path("data", "list").String(), `"id":42`) {
		t.Errorf("serialized: %s", j.Path("data", "list").String())
	}
}

func TestBigIntegersKeepTheirPrecision(t *testing.T) {
	j, _ := ParseString(`{"aid":1145141919810,"ts":1700000000,"f":2.0,"url":"https://a/b?c=1&d=<2>"}`)
	if i, _ := j.Get("aid").Int64(); i != 1145141919810 {
		t.Errorf("aid = %d", i)
	}
	if s := j.Get("ts").String(); s != "1700000000" {
		t.Errorf("ts = %q", s)
	}
	if s := j.Get("f").String(); s != "2" {
		t.Errorf("f = %q", s)
	}
	if s := j.Serialized(); !strings.Contains(s, `"url":"https://a/b?c=1&d=<2>"`) {
		t.Errorf("serialized escapes: %s", s)
	}
}

func TestInvalidJSON(t *testing.T) {
	if _, err := ParseString("{"); err == nil {
		t.Error("expected an error")
	}
}
