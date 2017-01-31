// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zapcore_test

import (
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap/internal/buffers"
	. "go.uber.org/zap/zapcore"
)

func testJSONConfig() JSONConfig {
	msgF := func(msg string) Field {
		return Field{Type: StringType, String: msg, Key: "msg"}
	}
	timeF := func(t time.Time) Field {
		millis := t.UnixNano() / int64(time.Millisecond)
		return Field{Type: Int64Type, Integer: millis, Key: "ts"}
	}
	levelF := func(l Level) Field {
		return Field{Type: StringType, String: l.String(), Key: "level"}
	}
	nameF := func(n string) Field {
		if n == "" {
			return Field{Type: SkipType}
		}
		return Field{Type: StringType, String: n, Key: "name"}
	}
	return JSONConfig{
		MessageFormatter: msgF,
		TimeFormatter:    timeF,
		LevelFormatter:   levelF,
		NameFormatter:    nameF,
	}
}

func BenchmarkJSONLogMarshalerFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		enc := NewJSONEncoder(testJSONConfig())
		enc.AddObject("nested", ObjectMarshalerFunc(func(enc ObjectEncoder) error {
			enc.AddInt64("i", int64(i))
			return nil
		}))
	}
}

func BenchmarkZapJSON(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			enc := NewJSONEncoder(testJSONConfig())
			enc.AddString("str", "foo")
			enc.AddInt64("int64-1", 1)
			enc.AddInt64("int64-2", 2)
			enc.AddFloat64("float64", 1.0)
			enc.AddString("string1", "\n")
			enc.AddString("string2", "💩")
			enc.AddString("string3", "🤔")
			enc.AddString("string4", "🙊")
			enc.AddBool("bool", true)
			buf, _ := enc.EncodeEntry(Entry{
				Message: "fake",
				Level:   DebugLevel,
			}, nil)
			buffers.Put(buf)
		}
	})
}

func BenchmarkStandardJSON(b *testing.B) {
	record := struct {
		Level   string                 `json:"level"`
		Message string                 `json:"msg"`
		Time    time.Time              `json:"ts"`
		Fields  map[string]interface{} `json:"fields"`
	}{
		Level:   "debug",
		Message: "fake",
		Time:    time.Unix(0, 0),
		Fields: map[string]interface{}{
			"str":     "foo",
			"int64-1": int64(1),
			"int64-2": int64(1),
			"float64": float64(1.0),
			"string1": "\n",
			"string2": "💩",
			"string3": "🤔",
			"string4": "🙊",
			"bool":    true,
		},
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			json.Marshal(record)
		}
	})
}