package routerdriver

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var Print = fmt.Println

const (
	METHOD_STATIC = "static"
)

// Store define drouter
type UrlMap struct {
	Store map[string][]*node
	Num   uint
	lock  sync.RWMutex
	pool  sync.Pool
}

//defined router node that stored detail
type node struct {
	path       string
	ParamSlice []string
	ParamMap   map[string]uint
	Handle     interface{}
	Type       string
}

//
type RouterRet struct {
	Path       string
	RealPath   string
	ParamSlice []string
	ParamMap   map[string]uint
	Handle     interface{}
	Type       string
	code       uint32
}

func NewMap() *UrlMap {
	m := &UrlMap{Store: make(map[string][]*node), Num: 0}
	m.pool.New = func() interface{} {
		return &RouterRet{}
	}

	return m
}

func (um *UrlMap) readMap(key string) (nodes []*node, err error) {
	return
}

//add defined router
func (um *UrlMap) addRouter(path string, handle interface{}, method string) {
	path = SlashPath(path)

	entityPath, paramArr := ParseDefinedUrl(path)

	if um.Store[entityPath] == nil {
		um.Store[entityPath] = make([]*node, 0)
	}

	nod := &node{
		path:       entityPath,
		ParamSlice: make([]string, 0),
		ParamMap:   make(map[string]uint),
		Handle:     handle,
		Type:       method,
	}

	for inx, val := range paramArr {
		nod.ParamMap[val] = uint(inx)
	}

	if can, err := um.exists(entityPath, nod); can {
		panic(err.Error() + " exists asert")
	}

	for inx, param := range paramArr {
		nod.ParamMap[param] = uint(inx)
	}
	nod.ParamSlice = paramArr

	um.Store[entityPath] = append(um.Store[entityPath], nod)
	um.Num++
	return
}

//Determine whether or not to repeat the route
func (um *UrlMap) exists(real string, node *node) (bool, error) {
	if um.Store[real] == nil {
		return false, nil
	}

	for _, nod := range um.Store[real] {
		if nod == nil {
			continue
		}

		if len(nod.ParamSlice) == len(node.ParamSlice) && strings.Contains(nod.Type, node.Type) {
			return true, errors.New("url has been defined" + node.path)
		}
	}
	return false, nil
}

//get return data based on the requested route
//return
func (um *UrlMap) getValue(reqPath string, margs ...string) (*RouterRet, error) {
	//reqPath = SlashPath(reqPath)

	rpLen := len(reqPath)
	var nods []*node
	var key = ""
	method := ""
	var values []string = make([]string, strings.Count(reqPath, "/"))
	var jk = rpLen - 1
	var valuesRealLen int

	if len(margs) >= 1 {
		method = margs[0]
	}

	//Reversal traversal url
	for i := rpLen - 1; i >= 0; i-- {
		if reqPath[i] != '/' && i != (rpLen-1) {
			continue
		}

		// if not find loop backward;
		if jk != i {
			values[valuesRealLen] = reqPath[i+1 : jk]
			jk = i
			valuesRealLen++
		}

		// if find the url map break the loop
		key = reqPath[:i+1]
		nods = um.Store[key]
		if len(nods) != 0 {
			jk = i
			break
		}
	}
	//not found urlmap || urlMap[key] is empty
	if key == "" || len(nods) == 0 {
		if jk > 0 || valuesRealLen > 0 {
		}
		return nil, errors.New("path is not valid" + method + values[0])
	}

	rt := um.pool.Get().(*RouterRet)
	rt.Path = reqPath
	rt.RealPath = key
	rt.code = 0
	rt.Type = method

	values = values[:valuesRealLen+1]
	vl := valuesRealLen
	for i := 0; i <= vl; i++ {
		if i >= (vl - i) {
			break
		}
		values[i], values[vl-i] = values[vl-i], values[i]
	}

	for _, nod := range nods {
		if len(nod.ParamSlice) == len(values) && strings.Contains(nod.Type, method) {
			rt.Handle = nod.Handle
			rt.Type = nod.Type
			rt.ParamSlice = values
			rt.ParamMap = nod.ParamMap
			break
		}
		//static server deal
		if nod.Type == METHOD_STATIC {
			rt.Handle = nod.Handle
			rt.Type = nod.Type
			if reqPath[rpLen-1] == '/' {
				rpLen--
			}
			rt.ParamSlice = []string{reqPath[jk:rpLen]}
			rt.ParamMap = nod.ParamMap
			break
		}
	}
	um.pool.Put(rt)

	return rt, nil
}

func (ret *RouterRet) ByName(name string) (string, bool) {
	var index, ok = ret.ParamMap[name]
	if !ok {
		return "", false
	}
	return ret.By(index)
}

func (ret *RouterRet) By(id uint) (string, bool) {
	if uint(len(ret.ParamSlice)) <= id {
		return "", false
	}
	return ret.ParamSlice[id], true
}

func (ret *RouterRet) SetParam(name string, value string) {
	ret.ParamSlice = append(ret.ParamSlice, value)
	ret.ParamMap[name] = uint(len(ret.ParamSlice) - 1)
}
