// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discovery

import (
	"encoding/json"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
)

// new
type CreateValueFunc func() interface{}

var (
	newBytes  CreateValueFunc = func() interface{} { return []byte(nil) }
	newString CreateValueFunc = func() interface{} { return "" }
	newMap    CreateValueFunc = func() interface{} { return make(map[string]string) }
)

// parse
type ParseValueFunc func(src []byte, dist interface{}) error

var (
	UnParse ParseValueFunc = func(src []byte, dist interface{}) error {
		d := dist.(*interface{})
		*d = src
		return nil
	}
	TextUnmarshal ParseValueFunc = func(src []byte, dist interface{}) error {
		d := dist.(*interface{})
		*d = util.BytesToStringWithNoCopy(src)
		return nil
	}
	MapUnmarshal ParseValueFunc = func(src []byte, dist interface{}) error {
		d := dist.(*interface{})
		m := (*d).(map[string]string)
		return json.Unmarshal(src, &m)
	}
	JsonUnmarshal ParseValueFunc = func(src []byte, dist interface{}) error {
		d := dist.(*interface{})
		return json.Unmarshal(src, *d)
	}
)

type Parser interface {
	Unmarshal(src []byte) (interface{}, error)
}

type CommonParser struct {
	NewFunc  CreateValueFunc
	FromFunc ParseValueFunc
}

func (p *CommonParser) Unmarshal(src []byte) (interface{}, error) {
	v := p.NewFunc()
	if err := p.FromFunc(src, &v); err != nil {
		return nil, err
	}
	return v, nil
}

var (
	BytesParser  = &CommonParser{newBytes, UnParse}
	StringParser = &CommonParser{newString, TextUnmarshal}
	MapParser    = &CommonParser{newMap, MapUnmarshal}
)