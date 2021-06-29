/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package sd

import (
	"testing"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

var ruleCache *MongoCacher

var rule1 = model.Rule{
	Domain:    "default",
	Project:   "default",
	ServiceID: "svc1",
	Rule: &discovery.ServiceRule{
		RuleId:   "123456789",
		RuleType: "black",
	},
}

var rule2 = model.Rule{
	Domain:    "default",
	Project:   "default",
	ServiceID: "svc1",
	Rule: &discovery.ServiceRule{
		RuleId:   "987654321",
		RuleType: "white",
	},
}

func init() {
	ruleCache = newRuleStore()
}

func TestRuleCacheBasicFunc(t *testing.T) {
	t.Run("init ruleCache,should pass", func(t *testing.T) {
		ruleCache := newRuleStore()
		assert.NotNil(t, ruleCache)
		assert.Equal(t, rule, ruleCache.cache.Name())
	})
	event1 := MongoEvent{
		DocumentID: "id1",
		Value:      rule1,
	}
	event2 := MongoEvent{
		DocumentID: "id2",
		Value:      rule2,
	}

	t.Run("add ruleCache, should pass", func(t *testing.T) {
		ruleCache.cache.ProcessUpdate(event1)
		assert.Equal(t, ruleCache.cache.Size(), 1)
		assert.Nil(t, ruleCache.cache.Get("id_not_exist"))
		assert.Equal(t, rule1.Rule.RuleId, ruleCache.cache.Get("id1").(model.Rule).Rule.RuleId)
		assert.Len(t, ruleCache.cache.GetValue("default/default/svc1/123456789"), 1)
		ruleCache.cache.ProcessUpdate(event2)
		assert.Equal(t, ruleCache.cache.Size(), 2)
		assert.Len(t, ruleCache.cache.GetValue("default/default/svc1/987654321"), 1)
		assert.Len(t, ruleCache.cache.GetValue("default/default/svc1"), 2)

	})

	t.Run("update ruleCache, should pass", func(t *testing.T) {
		assert.Equal(t, rule1, ruleCache.cache.Get("id1").(model.Rule))
		ruleUpdate := model.Rule{
			Domain:    "default",
			Project:   "default",
			ServiceID: "svc1",
			Rule: &discovery.ServiceRule{
				RuleId:      "123456789",
				RuleType:    "black",
				Description: "update",
			},
		}
		eventUpdate := MongoEvent{
			DocumentID: "id1",
			Value:      ruleUpdate,
		}
		ruleCache.cache.ProcessUpdate(eventUpdate)
		assert.Equal(t, ruleUpdate, ruleCache.cache.Get("id1").(model.Rule))
	})

	t.Run("delete ruleCache, should pass", func(t *testing.T) {
		ruleCache.cache.ProcessDelete(event1)
		assert.Nil(t, ruleCache.cache.Get("id1"))
		assert.Len(t, ruleCache.cache.GetValue("default/default/svc1"), 1)
		assert.Len(t, ruleCache.cache.GetValue("default/default/svc1/123456789"), 0)
		ruleCache.cache.ProcessDelete(event2)
		assert.Len(t, ruleCache.cache.GetValue("default/default/svc1"), 0)
		assert.Nil(t, ruleCache.cache.Get("id2"))
		assert.Len(t, ruleCache.cache.GetValue("default/default/svc1/987654321"), 0)
		assert.Len(t, ruleCache.cache.GetValue("default/default/svc1/123456789"), 0)
	})
}
