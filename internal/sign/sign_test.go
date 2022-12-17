package sign

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestTransformJSONStream(t *testing.T) {
	t.Run("Simple stream", func(t *testing.T) {
		f, err := os.Open("testdata/appjson.json")
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()

		buf := &bytes.Buffer{}
		if err := TransformJSONStream(buf, f, map[string]int{"some tag": 5}, map[int]int{4: 5}); err != nil {
			t.Error(err)
			return
		}

		res := []OldSign{}
		if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
			t.Error(err)
			return
		}

		if len(res) != 1 {
			t.Errorf("Expected 1 result, got %d", len(res))
			return
		}

		if res[0].ID != "5" {
			t.Errorf("Expected ID 5, got %s", res[0].ID)
			return
		}

		if res[0].Tags[0].ID != "5" {
			t.Errorf("Expected ID 5, got %s", res[0].ID)
			return
		}
	})
}
