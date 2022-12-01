package eval

import (
	"testing"

	"github.com/hatlonely/go-kit/strx"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewEvaluate(t *testing.T) {
	Convey("TestNewEvaluate", t, func() {
		var v interface{}
		v = map[string]interface{}{
			"Key1":  "val1",
			"Key2":  2,
			"#Key3": "Field1",
			"#Key4": "Field2 + Field3",
			"Key5": map[string]interface{}{
				"Key6":  "val6",
				"#Key7": "Field4",
			},
		}

		ev, err := NewEvaluable(v)
		So(err, ShouldBeNil)
		So(ev.consts, ShouldResemble, map[string]interface{}{
			"Key1": "val1",
			"Key2": 2,
			"Key5": map[string]interface{}{
				"Key6": "val6",
			},
		})

		So(ev.variables, ShouldNotBeNil)
		So(len(ev.variables), ShouldEqual, 3)
	})

}

func TestEvaluable_Evaluate(t *testing.T) {
	Convey("TestNewEvaluate", t, func() {
		var v interface{}
		v = map[string]interface{}{
			"Key1":  "val1",
			"Key2":  2,
			"#Key3": "Field1",
			"#Key4": "Field2 + Field3",
			"Key5": map[string]interface{}{
				"Key6":  "val6",
				"#Key7": "Field4",
			},
		}

		ev, err := NewEvaluable(v)
		So(err, ShouldBeNil)

		Convey("normal", func() {
			val, err := ev.Evaluate(map[string]string{
				"Field1": "val1",
				"Field2": "val2",
				"Field3": "val3",
				"Field4": "val4",
			})
			So(err, ShouldBeNil)
			So(val, ShouldResemble, map[string]interface{}{
				"Key1": "val1",
				"Key2": 2,
				"Key3": "val1",
				"Key4": "val2val3",
				"Key5": map[string]interface{}{
					"Key6": "val6",
					"Key7": "val4",
				},
			})
		})
	})
}

// cpu: Intel(R) Core(TM) i5-6600 CPU @ 3.30GHz
// BenchmarkDeepCopy
// BenchmarkDeepCopy/deepCopyByRefxSet
// BenchmarkDeepCopy/deepCopyByRefxSet-4         	  267333	      4251 ns/op
// BenchmarkDeepCopy/deepCopyByJsonMarshal
// BenchmarkDeepCopy/deepCopyByJsonMarshal-4     	  618873	      1952 ns/op
// BenchmarkDeepCopy/deepCopyByGoDeepcopy
// BenchmarkDeepCopy/deepCopyByGoDeepcopy-4      	  637680	      1871 ns/op
func TestDeepCopy(t *testing.T) {
	Convey("TestDeepCopy", t, func() {
		src := map[string]interface{}{
			"Key1": "val1",
			"Key2": 2,
			"Key5": map[string]interface{}{
				"Key6": "val6",
			},
		}

		Convey("deepCopyByRefxSet", func() {
			dst := deepCopyByRefxSet(src)
			So(dst, ShouldResemble, src)
		})

		Convey("deepCopyByGoDeepcopy", func() {
			dst := deepCopyByGoDeepcopy(src)
			So(dst, ShouldResemble, src)
		})

		Convey("deepCopyByJsonMarshal", func() {
			dst := deepCopyByJsonMarshal(src)
			So(strx.JsonMarshalSortKeys(dst), ShouldResemble, strx.JsonMarshalSortKeys(src))
		})
	})
}

func BenchmarkDeepCopy(b *testing.B) {
	src := map[string]interface{}{
		"Key1": "val1",
		"Key2": 2,
		"Key5": map[string]interface{}{
			"Key6": "val6",
		},
	}
	b.Run("deepCopyByRefxSet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = deepCopyByRefxSet(src)
		}
	})
	b.Run("deepCopyByJsonMarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = deepCopyByJsonMarshal(src)
		}
	})
	b.Run("deepCopyByGoDeepcopy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = deepCopyByGoDeepcopy(src)
		}
	})
}
