package resource

import (
	"encoding/base64"

	"gopkg.in/mgo.v2/bson"
)

type Pack func(failover interface{}) (string, error)
type UnPack func(failover interface{}, data string) error

func BsonPack(failover interface{}) (string, error) {
	data, err := bson.Marshal(failover)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func BsonUnPack(failover interface{}, data string) error {
	res, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return bson.Unmarshal(res, failover)
}

func ExpandInt(values []interface{}) []int {
	results := make([]int, 0, len(values))
	for _, value := range values {
		results = append(results, value.(int))
	}
	return results
}

func ExpandString(values []interface{}) []string {
	results := make([]string, 0, len(values))
	for _, value := range values {
		results = append(results, value.(string))
	}
	return results
}
