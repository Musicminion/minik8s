package entity

type Endpoints map[string][]string

func (endpoints Endpoints) Add(key, value string) {
	if len(endpoints[key]) == 0 {
		endpoints[key] = []string{value}
	} else {
		endpoints[key] = append(endpoints[key], value)
	}
}

func (endpoints Endpoints) Get(key string) []string {
	if endpoints == nil {
		return nil
	}
	return endpoints[key]
}

func (endpoints Endpoints) Del(key string) {
	delete(endpoints, key)
}
