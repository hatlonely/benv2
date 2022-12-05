package recorder

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileAnalyst(t *testing.T) {
	testBenJson := "test.ben.json"
	ioutil.WriteFile(testBenJson, []byte(`{"Time":"2022-12-05T16:36:53.653687+08:00","Name":"test-name","Step":[{"Time":"2022-12-05T16:36:53.653687+08:00","Req":{"Key1":"val1","Key2":"val2"},"Res":{"Key4":"val4","Key3":"val3"},"Err":null,"ErrCode":"","ResTime":2000000}],"ErrCode":"","ResTime":2000000}
{"Time":"2022-12-05T16:36:54.653687+08:00","Name":"test-name","Step":[{"Time":"2022-12-05T16:36:54.653687+08:00","Req":{"Key1":"val1","Key2":"val2"},"Res":{"Key3":"val3","Key4":"val4"},"Err":null,"ErrCode":"","ResTime":2000000}],"ErrCode":"","ResTime":2000000}
{"Time":"2022-12-05T16:36:55.653687+08:00","Name":"test-name","Step":[{"Time":"2022-12-05T16:36:55.653687+08:00","Req":{"Key1":"val1","Key2":"val2"},"Res":{"Key3":"val3","Key4":"val4"},"Err":null,"ErrCode":"","ResTime":2000000}],"ErrCode":"","ResTime":2000000}
`), 0644)
	defer os.RemoveAll(testBenJson)

	Convey("TestFileAnalyst", t, func() {
		analyst, err := NewFileAnalystWithOptions(&FileAnalystOptions{
			FilePath: testBenJson,
		})
		So(err, ShouldBeNil)

		st, et, err := analyst.TimeRange()
		So(err, ShouldBeNil)
		So(st.UnixNano(), ShouldEqual, 1670229413653687000)
		So(et.UnixNano(), ShouldEqual, 1670229415653687000)

		stream, err := analyst.Stat()
		So(err, ShouldBeNil)
		var stats []*UnitStat
		for {
			stat, err := stream.Next()
			So(err, ShouldBeNil)
			if stat == nil {
				break
			}
			stats = append(stats, stat)
		}

		So(len(stats), ShouldEqual, 3)
		So(stats[0].Time, ShouldEqual, "2022-12-05T16:36:53.653687+08:00")
		So(stats[1].Time, ShouldEqual, "2022-12-05T16:36:54.653687+08:00")
		So(stats[2].Time, ShouldEqual, "2022-12-05T16:36:55.653687+08:00")
	})
}
