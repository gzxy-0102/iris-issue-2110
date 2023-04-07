package jsonApi

import (
	"2110/app/model"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/jinzhu/inflection"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type QueryContext struct {
	RequestParam *orderedmap.OrderedMap
	Code         int
	NodeTree     map[string]*QueryNode
	NodePathMap  map[string]*QueryNode
	Err          error
	Explain      bool
	Source       model.Source
	ExplainMap   map[string]any
}

type QueryNode struct {
	ctx       *QueryContext
	start     int64
	depth     int8
	running   bool
	completed bool
	isList    bool
	page      any
	count     any

	sqlExecutor *DatabaseExecutor
	primaryKey  string
	relateKV    map[string]string

	Key         string
	Path        string
	RequestMap  orderedmap.OrderedMap
	CurrentData map[string]any
	ResultList  []map[string]any
	//ExplainList []map[string]any
	children map[string]*QueryNode
}

func (c *QueryContext) DoParse() {
	//startTime := time.Now().Nanosecond()
	for _, key := range c.RequestParam.Keys() {
		if c.Err != nil {
			return
		}
		if key == "@explain" {
			explain, _ := c.RequestParam.Get(key)
			c.Explain = explain.(bool)
		} else if c.NodeTree[key] == nil {
			c.parseByKey(key)
		}
	}
}

func (c *QueryContext) DoQuery() {
	for _, n := range c.NodeTree {
		if c.Err != nil {
			return
		}
		n.doQueryData()
	}
}

func (c *QueryContext) parseByKey(key string) {
	queryObject, _ := c.RequestParam.Get(key)
	if queryObject == nil {
		c.Err = fmt.Errorf("值不能为空, key: %s, value: %v", key, queryObject)
		return
	}
	r := reflect.TypeOf(queryObject)
	log.Infof("queryObject %+v", r)
	if queryMap, ok := queryObject.(orderedmap.OrderedMap); !ok {
		c.Err = fmt.Errorf("值类型不对， key: %s, value: %v", key, queryObject)
	} else {
		node := NewQueryNode(c, key, key, queryMap)
		log.Debugf("parse %s: %+v", key, node)
		c.NodeTree[key] = node
	}
}

func NewQueryNode(c *QueryContext, path, key string, queryMap orderedmap.OrderedMap) *QueryNode {
	n := &QueryNode{
		ctx:         c,
		Key:         strings.ToLower(key),
		Path:        path,
		RequestMap:  queryMap,
		start:       time.Now().UnixNano(),
		sqlExecutor: &DatabaseExecutor{Source: c.Source},
		isList:      strings.HasSuffix(key, "[]"),
	}
	c.NodePathMap[path] = n
	if n.isList {
		n.parseList()
	} else {
		n.parseOne()
	}
	return n
}

func (n *QueryNode) parseList() {
	root := n.ctx
	if root.Err != nil {
		return
	}
	if value, exists := n.RequestMap.Get(n.Key[0 : len(n.Key)-2]); exists {
		if kvs, ok := value.(orderedmap.OrderedMap); ok {
			root.Err = n.sqlExecutor.ParseTable(n.Key)
			n.parseKVs(kvs)
		} else {
			root.Err = fmt.Errorf("列表同名参数展开出错，listKey: %s, object: %v", n.Key, value)
			root.Code = http.StatusBadRequest
		}
		return
	}
	for _, field := range n.RequestMap.Keys() {
		value, _ := n.RequestMap.Get(field)
		log.Infof("当前字段：%v 值：%+v", field, value)
		if value == nil {
			root.Err = fmt.Errorf("field of [%s] value error, %s is nil", n.Key, field)
			return
		}
		switch field {
		case "page":
			n.page = value
		case "count":
			n.count = value
		default:
			if kvs, ok := value.(orderedmap.OrderedMap); ok {
				child := NewQueryNode(root, n.Path+"/"+field, field, kvs)
				if root.Err != nil {
					return
				}
				if n.children == nil {
					n.children = make(map[string]*QueryNode)
				}
				n.children[field] = child
				if nonDepend(n, child) && len(n.primaryKey) == 0 {
					n.primaryKey = field
				}
			}
		}
	}
}

func (n *QueryNode) parseOne() {
	root := n.ctx
	root.Err = n.sqlExecutor.ParseTable(n.Key)
	if root.Err != nil {
		root.Code = http.StatusBadRequest
		return
	}
	n.sqlExecutor.PageSize(0, 1)
	n.parseKVs(n.RequestMap)
}

func nonDepend(parent, child *QueryNode) bool {
	if len(child.relateKV) == 0 {
		return true
	}
	for _, v := range child.relateKV {
		if strings.HasPrefix(v, parent.Path) {
			return false
		}
	}
	return true
}

func (n *QueryNode) parseKVs(kvs orderedmap.OrderedMap) {
	root := n.ctx
	for _, field := range kvs.Keys() {
		value, _ := kvs.Get(field)
		log.Debugf("%s -> parse %s %v", n.Key, field, value)
		if value == nil {
			root.Err = fmt.Errorf("field value error, %s is nil", field)
			root.Code = http.StatusBadRequest
			return
		}
		if queryPath, ok := value.(string); ok && strings.HasSuffix(field, "@") { // @ 结尾表示有关联查询
			if n.relateKV == nil {
				n.relateKV = make(map[string]string)
			}
			fullPath := queryPath
			if strings.HasPrefix(queryPath, "/") {
				fullPath = n.Path + queryPath
			}
			n.relateKV[field[0:len(field)-1]] = fullPath
		} else {
			n.sqlExecutor.ParseCondition(field, value)
		}
	}
}

func (n *QueryNode) doQueryData() {
	if n.completed {
		return
	}
	n.running = true
	defer func() { n.running, n.completed = false, true }()
	root := n.ctx
	if len(n.relateKV) > 0 {
		for field, queryPath := range n.relateKV {
			value := root.findResult(queryPath)
			if root.Err != nil {
				return
			}
			n.sqlExecutor.ParseCondition(field, value)
		}
	}
	if !n.isList {
		n.ResultList, root.Err = n.sqlExecutor.Exec()
		if root.Explain {
			if root.ExplainMap == nil {
				root.ExplainMap = make(map[string]any)
			}
			sql := n.sqlExecutor.ToSQL()
			if _, exists := root.ExplainMap[sql]; !exists {
				explain, err := n.sqlExecutor.ExexExplain()
				if err == nil {
					root.ExplainMap[sql] = explain
				}
			}
		}
		if len(n.ResultList) > 0 {
			n.CurrentData = n.ResultList[0]
			return
		}
		return
	}
	primary := n.children[n.primaryKey]
	log.Infof("分页：page:%v limit:%v", n.page, n.count)
	primary.sqlExecutor.PageSize(n.page, n.count)
	primary.doQueryData()
	if root.Err != nil {
		return
	}
	listData := primary.ResultList
	n.ResultList = make([]map[string]any, len(listData))
	for i, x := range listData {
		n.ResultList[i] = make(map[string]any)
		n.ResultList[i][n.primaryKey] = x
		primary.CurrentData = x
		if len(n.children) > 0 {
			for _, child := range n.children {
				if child != primary {
					child.doQueryData()
					n.ResultList[i][child.Key] = child.Result()
				}
			}
		}
	}
}

func (c *QueryContext) findResult(value string) any {
	i := strings.LastIndex(value, "/")
	path := value[0:i]
	node := c.NodePathMap[path]
	if node == nil {
		c.Err = fmt.Errorf("关联查询参数有误: %s", value)
		return nil
	}
	if node.running {
		c.Err = fmt.Errorf("有循环依赖")
		return nil
	}
	node.doQueryData()
	if c.Err != nil {
		return nil
	}
	if node.CurrentData == nil {
		log.Infof("查询结果为空，queryPath: %v", value)
		return nil
	}
	key := value[i+1:]
	return node.CurrentData[key]
}

func (n *QueryNode) Result() any {
	if n.isList {
		//	doQueryData中处理Key名的话 数据不稳定 会有问题 在最后获取结果时 对数据进行循环处理
		for index, value := range n.ResultList {
			for k, v := range value {
				if strings.HasSuffix(k, "[]") {
					key := k
					if key == "[]" {
						key = "list"
					} else {
						key = inflection.Plural(strings.Replace(key, "[]", "", -1))
					}
					value[key] = v
					delete(value, k)
					n.ResultList[index] = value
				}
			}
		}
		return n.ResultList
	}
	if len(n.ResultList) > 0 {
		return n.ResultList[0]
	}
	return nil
}

func (c *QueryContext) End(code int, msg string) {
	c.Code = code
	log.Errorf("发生错误，终止处理, code: %d, msg: %s", code, msg)
}

func (c *QueryContext) Response() map[string]any {
	c.DoParse()
	if c.Err == nil {
		c.DoQuery()
	}
	resultMap := make(map[string]any)
	resultMap["code"] = c.Code
	if c.Err != nil {
		resultMap["code"] = http.StatusInternalServerError
		resultMap["message"] = c.Err.Error()
	} else {
		data := make(map[string]any)
		for k, v := range c.NodeTree {
			//log.Debugf("response.nodeMap K: %s, V: %v", k, v)
			key := k
			if strings.HasSuffix(key, "[]") {
				if key == "[]" {
					key = "list"
				} else {
					key = inflection.Plural(strings.Replace(key, "[]", "", -1))
				}
			}
			data[key] = v.Result()
		}
		if c.Explain {
			data["explain"] = c.ExplainMap
		}
		resultMap["data"] = data
	}
	log.Infof("执行结果：%v 响应结果: %v", c.NodeTree, resultMap)
	return resultMap
}
