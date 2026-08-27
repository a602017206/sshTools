package service

import (
	"context"
	"fmt"
)

func unsupportedNativeResourceDetails(database string) (NativeResourceDetails, error) {
	return NativeResourceDetails{}, fmt.Errorf("%s 暂不支持资源详情", database)
}

func (*mongoNativeSession) DescribeResource(context.Context, string, string) (NativeResourceDetails, error) {
	return unsupportedNativeResourceDetails("MongoDB")
}
func (*memcachedNativeSession) DescribeResource(context.Context, string, string) (NativeResourceDetails, error) {
	return unsupportedNativeResourceDetails("Memcached")
}
func (*cassandraNativeSession) DescribeResource(context.Context, string, string) (NativeResourceDetails, error) {
	return unsupportedNativeResourceDetails("Cassandra")
}
func (*couchbaseNativeSession) DescribeResource(context.Context, string, string) (NativeResourceDetails, error) {
	return unsupportedNativeResourceDetails("Couchbase")
}
func (*influxDBNativeSession) DescribeResource(context.Context, string, string) (NativeResourceDetails, error) {
	return unsupportedNativeResourceDetails("InfluxDB")
}
func (*neo4jNativeSession) DescribeResource(context.Context, string, string) (NativeResourceDetails, error) {
	return unsupportedNativeResourceDetails("Neo4j")
}
