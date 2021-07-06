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
	"github.com/stretchr/testify/assert"
)

var schemaCache *MongoCacher

var schema1 = model.Schema{
	Domain:    "default",
	Project:   "default",
	ServiceID: "svc1",
	SchemaID:  "123456789",
	Schema:    "schema1",
}

var schema2 = model.Schema{
	Domain:    "default",
	Project:   "default",
	ServiceID: "svc1",
	SchemaID:  "987654321",
	Schema:    "schema2",
}

func init() {
	schemaCache = newSchemaStore()
}

func TestSchemaCacheBasicFunc(t *testing.T) {
	t.Run("init schemaCache,should pass", func(t *testing.T) {
		schemaCache := newSchemaStore()
		assert.NotNil(t, schemaCache)
		assert.Equal(t, schema, schemaCache.cache.Name())
	})
	event1 := MongoEvent{
		DocumentID: "id1",
		Value:      schema1,
	}
	event2 := MongoEvent{
		DocumentID: "id2",
		Value:      schema2,
	}

	t.Run("add schemaCache, should pass", func(t *testing.T) {
		schemaCache.cache.ProcessUpdate(event1)
		assert.Equal(t, schemaCache.cache.Size(), 1)
		assert.Nil(t, schemaCache.cache.Get("id_not_exist"))
		assert.Equal(t, schema1.SchemaID, schemaCache.cache.Get("id1").(model.Schema).SchemaID)
		assert.Len(t, schemaCache.cache.GetValue("default/default/svc1/123456789"), 1)
		schemaCache.cache.ProcessUpdate(event2)
		assert.Equal(t, schemaCache.cache.Size(), 2)
		assert.Len(t, schemaCache.cache.GetValue("default/default/svc1/987654321"), 1)
		assert.Len(t, schemaCache.cache.GetValue("default/default/svc1"), 2)

	})

	t.Run("update schemaCache, should pass", func(t *testing.T) {
		assert.Equal(t, schema1, schemaCache.cache.Get("id1").(model.Schema))
		schemaUpdate := model.Schema{
			Domain:    "default",
			Project:   "default",
			ServiceID: "svc1",
			SchemaID:  "123456789",
			Schema:    "schema1",
		}
		eventUpdate := MongoEvent{
			DocumentID: "id1",
			Value:      schemaUpdate,
		}
		schemaCache.cache.ProcessUpdate(eventUpdate)
		assert.Equal(t, schemaUpdate, schemaCache.cache.Get("id1").(model.Schema))
	})

	t.Run("delete schemaCache, should pass", func(t *testing.T) {
		schemaCache.cache.ProcessDelete(event1)
		assert.Nil(t, schemaCache.cache.Get("id1"))
		assert.Len(t, schemaCache.cache.GetValue("default/default/svc1"), 1)
		assert.Len(t, schemaCache.cache.GetValue("default/default/svc1/123456789"), 0)
		schemaCache.cache.ProcessDelete(event2)
		assert.Len(t, schemaCache.cache.GetValue("default/default/svc1"), 0)
		assert.Nil(t, schemaCache.cache.Get("id2"))
		assert.Len(t, schemaCache.cache.GetValue("default/default/svc1/987654321"), 0)
		assert.Len(t, schemaCache.cache.GetValue("default/default/svc1/123456789"), 0)
	})
}
