package routerdriver

import (
	"errors"
	"fmt"
	"strings"
)

var Print = fmt.Println

// Store define drouter
type UrlMap struct {
	Store map[string][]*node
	Num   uint
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
}

//add defined router
func (um *UrlMap) addRouter(path string, handle interface{}, method string) {
	path = CleanPath(path)
	pathArr := strings.Split(path, "*")

	al := len(pathArr)
	if al == 0 {
		return
	}

	entityPath := pathArr[0]
	pathArr1 := strings.Split(entityPath, ":")
	entityPath = pathArr1[0]
	pathArr1 = append(pathArr1, pathArr[1:]...)
	count := 1
	if um.Store[entityPath] == nil {
		um.Store[entityPath] = make([]*node, count)
	}

	nod := &node{
		path:       entityPath,
		ParamSlice: pathArr1[1:],
		ParamMap:   make(map[string]uint),
		Handle:     handle,
		Type:       method,
	}

	for inx, val := range pathArr1[1:] {
		nod.ParamMap[val] = uint(inx)
	}

	if can, err := um.exists(entityPath, nod); can {
		panic(err.Error())
	}

	switch al {
	case 1:
	case 2:
	}

	um.Store[entityPath] = append(um.Store[entityPath], nod)
	um.Num++
	return
}

//Determine whether or not to repeat the route
func (um *UrlMap) exists(real string, node *node) (bool, error) {
	if um.Store[real] == nil {
		return true, nil
	}

	for _, nod := range um.Store[real] {
		if nod == nil {
			continue
		}

		if len(nod.ParamSlice) == len(node.ParamSlice) {
			return false, errors.New("url has been defined" + node.path)
		}
	}
	return true, nil
}

//get return data based on the requested route
//return
func (um *UrlMap) getValue(reqPath string, margs ...string) *RouterRet {

	rpLen := len(reqPath)
	var nods []*node
	var key, method string
	var values []string = make([]string, 0)
	var jk = rpLen - 1

	if len(margs) >= 1 {
		method = margs[0]
	}

	//Reversal traversal url
	for i := rpLen - 1; i >= 0; i-- {
		if reqPath[i] == '/' {
			if um.Store[reqPath[:(i+1)]] != nil {
				key = reqPath[:(i + 1)]
				nods = um.Store[key]
				break
			} else {
				if jk != i {
					vals := make([]string, 1)
					vals = append(vals, reqPath[i:jk])
					values = append(vals, values...)
					jk = i
				}
			}
		}
	}

	//not found urlmap || urlMap[key] is empty
	if key == "" || len(nods) == 0 {
		return nil
	}

	param := &RouterRet{}
	param.Path = reqPath
	param.RealPath = key
	for _, nod := range nods {
		if len(nod.ParamSlice) == len(values) || strings.Contains(nod.Type, method) {
			param.Handle = nod.Handle
			param.Type = nod.Type
			param.ParamSlice = values
			param.ParamMap = nod.ParamMap
		}
	}

	return param
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
