package mybase

import (
	"encoding/json"
	"testing"
)

func TestEasyGet(t *testing.T) {
	h := H{}
	if err := json.Unmarshal([]byte(`{"a":"1111","b":222,"c":{"d":333}}`), &h); err != nil {
		t.Error(err)
		return
	}
	if a, ok := h.GetString("a"); !ok {
		t.Errorf("a not found")
		return
	} else {
		t.Log("a=", a)
	}

	if b, ok := h.GetInt("b"); !ok {
		t.Errorf("b not found")
		return
	} else {
		t.Log("b=", b)
	}

	c, ok := h.GetH("c")
	if !ok {
		t.Errorf("c not found")
		return
	}
	if d, ok := c.GetInt("d"); !ok {
		t.Errorf("d not found")
		return
	} else {
		t.Log("d=", d)
	}

	//测试buff数据存入和获取
	h2 := H{"buff": []byte("abc中文123"), "gids": []int32{1, 2, 3, 4}}
	buf := make([]byte, 0)
	t.Log(h2.Get("buff", &buf))
	t.Log(string(buf))

	gids := make([]int64, 0)
	t.Log(h2.Get("gids", &gids))
	t.Log(gids)
}
