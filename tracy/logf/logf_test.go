// Copyright 2023 appkit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logf

import (
	"errors"
	"fmt"
	"testing"

	"github.com/decentplatforms/appkit/tracy"
)

type TestWriter struct {
	Fail bool
	Last string
}

func (writer *TestWriter) Write(msg []byte) (n int, err error) {
	if writer.Fail {
		return 0, errors.New("writer configured to fail")
	}
	writer.Last = string(msg)
	return len(msg), nil
}

var formats = map[string]tracy.Formatter{
	"syslog_rfc3164": Syslog3164Format(SyslogConfig{}),
	"syslog_rfc5424": Syslog5424Format(SyslogConfig{}),
	"json":           JSONFormat(JSONConfig{}),
	"json_pretty":    JSONPrettyFormat(JSONConfig{Indent: "  "}),
}

var syslog_formats = map[string]tracy.Formatter{
	"syslog_rfc3164": Syslog3164Format(SyslogConfig{}),
	"syslog_rfc5424": Syslog5424Format(SyslogConfig{}),
}

func TestLogger(t *testing.T) {
	t.Run("default logger", func(t *testing.T) {
		for name, format := range formats {
			tw := &TestWriter{}
			conf := tracy.Config{
				MaxLevel:     tracy.Warning,
				DefaultLevel: tracy.Informational,
				Format:       format,
				Output:       tw,
			}
			t.Run(name+" format", func(t *testing.T) {
				conf.Format = format
				log, err := tracy.NewLogger(conf)
				if err != nil {
					t.Fatal(err)
				}
				for i := tracy.MOST_SEVERE; i < tracy.LEAST_SEVERE; i++ {
					lvl := tracy.LogLevel(i)
					msg := fmt.Sprintf("test log at level %s", lvl)
					log.Log(tracy.LogLevel(i), msg)
					if i <= conf.MaxLevel {
						if expected := format.FormatAndNormalize(lvl, msg, tracy.NewProps()); tw.Last != expected {
							t.Error("wrong log at", lvl, tw.Last, expected)
						}
					} else {
						if tw.Last != "" {
							t.Error("logger shouldn't have logged at", lvl)
						}
					}
					tw.Last = ""
				}
			})
		}
	})
	t.Run("with props", func(t *testing.T) {
		for name, format := range formats {
			tw := &TestWriter{}
			conf := tracy.Config{
				MaxLevel:     tracy.Warning,
				DefaultLevel: tracy.Informational,
				Format:       format,
				Output:       tw,
			}
			props := []tracy.Prop{{Name: "prop1", Value: "hello world"}, {Name: "prop2", Value: 100}, {Name: "prop3", Value: []string{"hello", "world"}}}
			t.Run(name+" format", func(t *testing.T) {
				conf.Format = format
				log, err := tracy.NewLogger(conf)
				if err != nil {
					t.Fatal(err)
				}
				for i := tracy.MOST_SEVERE; i < tracy.LEAST_SEVERE; i++ {
					lvl := tracy.LogLevel(i)
					msg := fmt.Sprintf("test log at level %s", lvl)
					log.Log(tracy.LogLevel(i), msg, props...)
					if i <= conf.MaxLevel {
						if expected := format.FormatAndNormalize(lvl, msg, tracy.NewProps(props...)); tw.Last != expected {
							t.Error("wrong log at", lvl, tw.Last, expected)
						}
					} else {
						if tw.Last != "" {
							t.Error("logger shouldn't have logged at", lvl)
						}
					}
					tw.Last = ""
				}
			})
		}
	})
	t.Run("syslog logger", func(t *testing.T) {
		for name, format := range syslog_formats {
			hostname := "jnichols@debbie"
			appname := "some-other-app"
			msgid := "testing"
			tw := &TestWriter{}
			conf := tracy.Config{
				MaxLevel:     tracy.Warning,
				DefaultLevel: tracy.Informational,
				Format:       format,
				Output:       tw,
			}
			props := []tracy.Prop{{Name: SYSLOG_HOSTNAME, Value: hostname}, {Name: SYSLOG_APPNAME, Value: appname}, {Name: SYSLOG_TAG, Value: msgid}}
			t.Run(name+" format", func(t *testing.T) {
				conf.Format = format
				log, err := tracy.NewLogger(conf)
				if err != nil {
					t.Fatal(err)
				}
				for i := tracy.MOST_SEVERE; i < tracy.LEAST_SEVERE; i++ {
					lvl := tracy.LogLevel(i)
					msg := fmt.Sprintf("test log at level %s", lvl)
					log.Log(tracy.LogLevel(i), msg, props...)
					if i <= conf.MaxLevel {
						if expected := format.FormatAndNormalize(lvl, msg, tracy.NewProps(props...)); tw.Last != expected {
							t.Error("wrong log at", lvl, tw.Last, expected)
						}
					} else {
						if tw.Last != "" {
							t.Error("logger shouldn't have logged at", lvl)
						}
					}
					tw.Last = ""
					t.Fail()
				}
			})
		}
	})
}