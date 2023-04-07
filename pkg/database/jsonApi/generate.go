package jsonApi

import (
	"2110/app/model"
	"2110/pkg/database/dynamicSource"
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type DatabaseExecutor struct {
	//	表名
	table string
	//	字段信息
	columns []string
	//	条件
	where  []string
	params []any
	order  string
	group  string
	limit  int
	page   int
	//	使用的数据源
	Source model.Source
}

func (e *DatabaseExecutor) Table() string {
	return e.table
}

func (e *DatabaseExecutor) ParseTable(t string) error {
	if strings.HasSuffix(t, "[]") {
		t = t[0 : len(t)-2]
	}
	exists, err := dynamicSource.TableExists(e.Source, t)
	if err != nil {
		return err
	}
	e.table = t
	if !exists {
		return fmt.Errorf("table: %s not exists", e.table)
	}
	return nil
}

func (e *DatabaseExecutor) ToSQL() string {
	var buf bytes.Buffer
	buf.WriteString("SELECT ")
	if e.columns == nil {
		buf.WriteString("*")
	} else {
		buf.WriteString(strings.Join(e.columns, ","))
	}
	buf.WriteString(" FROM ")
	buf.WriteString(e.table)
	if len(e.where) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(e.where, " and "))
	}
	if e.order != "" {
		buf.WriteString(" ORDER BY ")
		buf.WriteString(e.order)
	}
	buf.WriteString(" LIMIT ")
	buf.WriteString(strconv.Itoa(e.limit))
	if e.page > 0 {
		//	0为第一页 第二页开始指定offset
		buf.WriteString(" OFFSET ")
		buf.WriteString(strconv.Itoa(e.limit * e.page))
	}
	return buf.String()
}

func (e *DatabaseExecutor) ParseCondition(field string, value any) {
	if values, ok := value.([]any); ok {
		// 数组使用 IN 条件
		condition := field + " in ("
		for i, v := range values {
			if i == 0 {
				condition += "?"
			} else {
				condition += ",?"
			}
			e.params = append(e.params, v)
		}
		e.where = append(e.where, condition+")")
	} else if valueStr, ok := value.(string); ok {
		if strings.HasPrefix(field, "@") {
			switch field[1:] {
			case "order":
				e.order = valueStr
			case "column":
				e.columns = strings.Split(valueStr, ",")
			}
		} else {
			e.where = append(e.where, field+"=?")
			e.params = append(e.params, valueStr)
		}
	} else {
		e.where = append(e.where, field+"=?")
		e.params = append(e.params, value)
	}
}

func (e *DatabaseExecutor) Exec() ([]map[string]any, error) {
	sql := e.ToSQL()
	log.Debugf("exec %s, params: %v", sql, e.params)
	err, result := dynamicSource.QueryBySource(e.Source, sql, e.params)
	if err != nil {
		return nil, err
	}
	return result.([]map[string]any), err
}

func (e *DatabaseExecutor) ExexExplain() (map[string]any, error) {
	sql := fmt.Sprintf("EXPLAIN %s", e.ToSQL())
	err, result := dynamicSource.QueryBySource(e.Source, sql, e.params)
	if err != nil {
		return nil, err
	}
	return result.([]map[string]any)[0], err
}

func (e *DatabaseExecutor) PageSize(page any, count any) {
	e.page = parseNum(page, 0)
	e.limit = parseNum(count, 10)
	log.Infof("Page: %v Limit: %v SQL: %s", e.page, e.limit, e.ToSQL())
}

func parseNum(value any, defaultVal int) int {
	if n, ok := value.(float64); ok {
		return int(n)
	}
	if n, ok := value.(int); ok {
		return n
	}
	return defaultVal
}
