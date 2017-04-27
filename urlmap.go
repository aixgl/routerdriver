package routerdriver

import (
	"errors"
	"fmt"
	"strings"
)

var Print = fmt.Println

const (
	METHOD_STATIC = "static"
)

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
func (um *UrlMap) getValue(reqPath string, margs ...string) *RouterRet {
	reqPath = SlashPath(reqPath)

	rpLen := len(reqPath)
	var nods []*node
	var key, method string = "", ""
	var values []string = make([]string, 0)
	var jk = rpLen - 1

	if len(margs) >= 1 {
		method = margs[0]
	}

	//Reversal traversal url
	for i := rpLen - 1; i >= 0; i-- {
		if reqPath[i] == '/' {
			if jk != i {
				vals := make([]string, 1)
				vals[0] = reqPath[i+1 : jk]
				values = append(vals, values...)
				jk = i
			}

			if um.Store[reqPath[:(i+1)]] != nil {
				key = reqPath[:(i + 1)]
				nods = um.Store[key]
				jk = i
				break
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
		if len(nod.ParamSlice) == len(values) && strings.Contains(nod.Type, method) {
			param.Handle = nod.Handle
			param.Type = nod.Type
			param.ParamSlice = values
			param.ParamMap = nod.ParamMap
		}
		//static server deal
		if nod.Type == METHOD_STATIC {
			param.Handle = nod.Handle
			param.Type = nod.Type
			if reqPath[rpLen-1] == '/' {
				rpLen--
			}
			param.ParamSlice = []string{reqPath[jk:rpLen]}
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
