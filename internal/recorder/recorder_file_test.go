package recorder

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileRecorder(t *testing.T) {
	Convey("TestFileRecorder", t, func() {
		fr, err := NewFileRecorderWithOptions(&FileRecorderOptions{
			FilePath: "ben.json",
		})
		So(err, ShouldBeNil)

		Convey("Record", func() {
			So(fr.Record(&UnitStat{
				Name: "test-name",
				Step: []*StepStat{
					{
						Time: "2022-12-05T16:36:53.653687+08:00",
						Req: map[string]interface{}{
							"Key1": "val1",
							"Key2": "val2",
						},
						Res: map[string]interface{}{
							"Key3": "val3",
							"Key4": "val4",
						},
						Err:     nil,
						ErrCode: "",
						ResTime: 2 * time.Millisecond,
					},
				},
				ErrCode: "",
				ResTime: 2 * time.Millisecond,
			}), ShouldBeNil)

			So(fr.Record(&UnitStat{
				Name: "test-name",
				Step: []*StepStat{
					{
						Time: "2022-12-05T16:36:54.653687+08:00",
						Req: map[string]interface{}{
							"Key1": "val1",
							"Key2": "val2",
						},
						Res: map[string]interface{}{
							"Key3": "val3",
							"Key4": "val4",
						},
						Err:     nil,
						ErrCode: "",
						ResTime: 2 * time.Millisecond,
					},
				},
				ErrCode: "",
				ResTime: 2 * time.Millisecond,
			}), ShouldBeNil)

			So(fr.Record(&UnitStat{
				Name: "test-name",
				Step: []*StepStat{
					{
						Time: "2022-12-05T16:36:55.653687+08:00",
						Req: map[string]interface{}{
							"Key1": "val1",
							"Key2": "val2",
						},
						Res: map[string]interface{}{
							"Key3": "val3",
							"Key4": "val4",
						},
						Err:     nil,
						ErrCode: "",
						ResTime: 2 * time.Millisecond,
					},
				},
				ErrCode: "",
				ResTime: 2 * time.Millisecond,
			}), ShouldBeNil)

			So(fr.Close(), ShouldBeNil)
		})
	})
}
