package rediscache

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/go-redis/redis/v8"
)

type RedisCache interface {
	// 初始化删除所有的key
	InitCache() error

	// 增删改查的接口操作
	Put(key string, value interface{}) error
	Delete(key string) error
	Get(key string) (string, error)
	GetObject(key string, valueType interface{}) (interface{}, error)
	GetAllObject(valueType interface{}) (map[string]interface{}, error)
	Update(key string, value interface{}) error

	// 判断是否存在Redis缓存里面
	ifExists(key string) (bool, error)
}

// 对于redis缓存的实现
type rediscache struct {
	lock        sync.RWMutex
	redisClient *redis.Client
}

// 清空redis里面的所有的数据
func (r *rediscache) InitCache() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	ctx := context.Background()
	result := r.redisClient.FlushAll(ctx)
	return result.Err()
}

// 增删改查的接口操作
func (r *rediscache) Put(key string, value interface{}) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	ctx := context.Background()
	// result := r.redisClient.Set(ctx, key, value, 0)

	// 由于Redis的value需要是string类型，所以需要将value转换成string类型
	jsonStrByte, err := json.Marshal(value)
	if err != nil {
		return err
	}
	// jsonStrByte转化为string类型
	jsonStr := string(jsonStrByte)

	// 将key和value存入Redis
	result := r.redisClient.Set(ctx, key, jsonStr, 0)
	return result.Err()
}

func (r *rediscache) Delete(key string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	ctx := context.Background()
	result := r.redisClient.Del(ctx, key)
	return result.Err()
}

func (r *rediscache) Get(key string) (string, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	ctx := context.Background()
	value := r.redisClient.Get(ctx, key)
	return value.Result()
}

/*
// - 特别注意：这个函数的用法，可以参考测试代码里面的
// - 假设你自定义了一个类型type MyType struct {Name string Age int}

//   - 那么你可以这样调用GetObject函数：
//  ```
//     var myType MyType
//     obj, err := testRedisCache.GetObject("test-2", &myType)
//
//  ```
// 值得强调的是，第二个参数一定要是一个指针类型，否则会报错！！！函数会通过两个方式返回解析后的对象，你可以自己选择一种解析
// 因为你传递的第二个参数是一个指针类型，所以函数会直接修改这个指针指向的对象，你直接解析传入的第二个参数就可以了
// 函数会返回一个interface{}类型的对象(这是一个指针!),你可以解析这个指针指向的对象，然后获取返回的结果
// 如果你仍然觉得困难，可以参考测试代码里面的用法
// 【强调】：第二个参数是一个【指针】,需要用【&XXX】的方法传递
*/
func (r *rediscache) GetObject(key string, valueType interface{}) (interface{}, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	ctx := context.Background()
	result := r.redisClient.Get(ctx, key)
	if result.Err() != nil {
		// // 如果没有查找到对象，那么不返回错误，直接返回nil
		if result.Err() == redis.Nil {
			return nil, nil
		}
		return nil, result.Err()
	}
	jsonStr, err := result.Result()

	if err != nil {
		return nil, err
	} else {
		err := json.Unmarshal([]byte(jsonStr), valueType)
		return valueType, err
	}
}

// 获取所有的对象
// - 特别注意：这个函数的用法,需要再valeType里面传递一组集合的类型，比如[]People, 就需要在valueType传递一个People的类型
// 但是返回的时候，返回的是一个[]People类型的对象，所以需要慎重使用
// 此外，如果json解析失败，会直接跳过这个对象，不会报错
func (r *rediscache) GetAllObject(valueType interface{}) (map[string]interface{}, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	ctx := context.Background()
	result := r.redisClient.Keys(ctx, "*")
	if result.Err() != nil {
		return nil, result.Err()
	}
	keys, err := result.Result()

	if err != nil {
		return nil, err
	} else {
		// 遍历所有的key，然后获取所有的对象
		allObject := make(map[string]interface{})
		for _, key := range keys {
			// // 获取所有的对象
			// value, err := r.GetObject(key, valueType)

			innerCtx := context.Background()
			getKeyResult := r.redisClient.Get(innerCtx, key)
			if getKeyResult.Err() != nil {
				return nil, getKeyResult.Err()
			}
			jsonStr, err := getKeyResult.Result()

			if err != nil {
				return nil, err
			} else {
				// 创建一个valueType类型的对象，valueType是一个指针类型
				tmpValueType := reflect.New(reflect.TypeOf(valueType).Elem()).Interface()
				err := json.Unmarshal([]byte(jsonStr), tmpValueType)
				if err != nil {
					continue
				}
				allObject[key] = tmpValueType
			}
		}
		return allObject, nil
	}
}

func (r *rediscache) Update(key string, value interface{}) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	ctx := context.Background()

	// 由于Redis的value需要是string类型，所以需要将value转换成string类型
	jsonStrByte, err := json.Marshal(value)
	if err != nil {
		return err
	}
	// jsonStrByte转化为string类型
	jsonStr := string(jsonStrByte)

	// 将key和value存入Redis
	result := r.redisClient.Set(ctx, key, jsonStr, 0)
	return result.Err()
}

func (r *rediscache) ifExists(key string) (bool, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	result := r.redisClient.Exists(context.Background(), key)
	exists, err := result.Result()
	if err != nil {
		return false, err
	} else {
		return exists == 1, nil
	}
}

// 创建一个redis缓存的实例
func NewRedisCache(cacheID int) RedisCache {
	if cacheID < 0 || cacheID > 15 {
		panic("cacheID must be in [0,15]")
	}

	return &rediscache{
		lock: sync.RWMutex{},
		redisClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
			DB:   cacheID,
		}),
	}
}
