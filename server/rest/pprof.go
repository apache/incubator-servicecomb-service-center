/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rest

import (
	"net/http/pprof"

	"github.com/apache/servicecomb-service-center/server/config"
)

func init() {
	if config.GetServer().EnablePProf {
		RegisterServerHandleFunc("/debug/pprof/", pprof.Index)
		RegisterServerHandleFunc("/debug/pprof/profile", pprof.Profile)
		RegisterServerHandleFunc("/debug/pprof/symbol", pprof.Symbol)
		RegisterServerHandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		RegisterServerHandleFunc("/debug/pprof/trace", pprof.Trace)
		RegisterServerHandler("/debug/pprof/heap", pprof.Handler("heap"))
		RegisterServerHandler("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		RegisterServerHandler("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
		RegisterServerHandler("/debug/pprof/block", pprof.Handler("block"))
	}
}
