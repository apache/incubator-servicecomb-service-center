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

package mongo

import (
	"context"

	"github.com/apache/servicecomb-service-center/datasource/cache"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/discovery"
)

func GetSchema(ctx context.Context, serviceID string, schemaID string) (*model.Schema, error) {
	schema, ok := cache.GetSchema(ctx, serviceID, schemaID)
	if ok && schema != nil {
		return schema, nil
	}
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceID(serviceID), mutil.SchemaID(schemaID))
	return dao.GetSchema(ctx, filter)
}

func GetSchemasByServiceID(ctx context.Context, serviceID string) ([]*discovery.Schema, error) {
	schemas, ok := cache.GetSchemas(ctx, serviceID)
	if ok {
		return schemas, nil
	}
	filter := mutil.NewDomainProjectFilter(util.ParseDomain(ctx), util.ParseProject(ctx), mutil.ServiceID(serviceID))
	return dao.GetSchemas(ctx, filter)

}
