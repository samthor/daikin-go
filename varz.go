package main

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

type varzValue struct {
	value interface{}
	when  time.Time
}

func (vz *varzValue) Render() string {
	if vz == nil {
		return "0"
	}
	switch v := vz.value.(type) {
	case float64:
		return fmt.Sprintf("%.1f", v) // always render as float
	case int64:
		return fmt.Sprintf("%d", v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	case nil:
		return "0"
	default:
		return "-"
	}
}

type VarzHandler struct {
	lock   sync.RWMutex
	values map[string]*varzValue
}

func (vh *VarzHandler) Update(key string, value interface{}) {
	vh.lock.Lock()
	defer vh.lock.Unlock()

	if vh.values == nil {
		vh.values = make(map[string]*varzValue)
	}

	if value == nil {
		delete(vh.values, key)
	} else {
		vh.values[key] = &varzValue{value, time.Now()}
	}
}

func (vh *VarzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vh.lock.RLock()
	defer vh.lock.RUnlock()

	keys := make([]string, 0, len(vh.values))
	for k := range vh.values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := vh.values[k]
		fmt.Fprintf(w, "%s %s\n", k, v.Render())
	}
}
