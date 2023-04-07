package controller

import (
	"2110/app/model"
	"2110/pkg/database/dynamicSource"
	"2110/pkg/database/jsonApi"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/iancoleman/orderedmap"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type JSONApi struct {
	Base
	ORM *gorm.DB
}

func (api *JSONApi) Configure(r iris.Party) {
	r.Post("/{mark:uuid}/get", api.get)
	r.Post("/{mark:uuid}/post", api.create)
	r.Post("/{mark:uuid}/put", api.update)
	r.Post("/{mark:uuid}/delete", api.delete)
}

func (api *JSONApi) get(ctx iris.Context) {
	//if !ctx.Values().Exists("source") {
	//	api.FAIL(ctx, "数据源错误", nil)
	//	return
	//}
	//source := ctx.Values().Get("source")
	source, err := api.getSourceByCtx(ctx)
	if err != nil {
		api.FAIL(ctx, "数据源错误", nil)
		return
	}
	reqJson := orderedmap.New()
	if err := ctx.ReadJSON(&reqJson); err != nil {
		api.Errorf(ctx, "JSON API GET Error: %v", err)
		api.FAIL(ctx, err.Error(), nil)
		return
	}
	api.Infof(ctx, "请求key %+v", reqJson.Keys())
	queryContext := &jsonApi.QueryContext{
		RequestParam: reqJson,
		Code:         http.StatusOK,
		NodeTree:     make(map[string]*jsonApi.QueryNode),
		NodePathMap:  make(map[string]*jsonApi.QueryNode),
		Source:       *source,
	}
	_ = ctx.JSON(queryContext.Response())
	ctx.StopWithStatus(http.StatusOK)
}

func (api *JSONApi) create(ctx iris.Context) {
	//if !ctx.Values().Exists("source") {
	//	api.FAIL(ctx, "数据源错误", nil)
	//	return
	//}
	//source := ctx.Values().Get("source")
	source, err := api.getSourceByCtx(ctx)
	if err != nil {
		api.FAIL(ctx, "数据源错误", nil)
		return
	}
	reqJson := make(map[string]any)
	if err := ctx.ReadJSON(&reqJson); err != nil {
		api.Errorf(ctx, "JSON API Create Error: %v", err)
		api.FAIL(ctx, err.Error(), nil)
		return
	}
	var errs []string
	for k, v := range reqJson {
		ex, err := dynamicSource.TableExists(*source, k)
		if ex {
			if kvs, ok := v.(map[string]any); ok {
				sql, args := generateInsert(k, kvs)
				err, result := dynamicSource.QueryBySource(*source, sql, args)
				if err != nil {
					errs = append(errs, fmt.Sprintf("Table: %s 创建数据失败：%s,%v 错误：%v", k, sql, args, err))
					api.Errorf(ctx, "JsonApiCreate Table Not Exists : %s; err = %v ", k, err)
				} else if r, ok := result.(bool); ok {
					if !r {
						errs = append(errs, fmt.Sprintf("Table: %s 创建数据失败：%s,%v", k, sql, args))
					}
				}
			}
		} else {
			errs = append(errs, fmt.Sprintf("Table Not Exists: %s err=%v", k, err))
			api.Errorf(ctx, "JsonApiCreate Table Not Exists : %s; err = %v ", k, err)
		}
	}
	errLen := len(errs)
	insertLen := len(reqJson)
	if errLen > 0 && errLen < insertLen {
		api.SUCCESS(ctx, "数据创建完成，但出现部分错误", iris.Map{
			"errors": errs,
		})
		return
	} else if errLen == insertLen {
		api.FAIL(ctx, "数据创建失败", iris.Map{
			"errors": errs,
		})
		return
	}
	api.SUCCESS(ctx, "数据创建完成", iris.Map{
		"errors": errs,
	})
	return
}

func (api *JSONApi) update(ctx iris.Context) {
	//if !ctx.Values().Exists("source") {
	//	api.FAIL(ctx, "数据源错误", nil)
	//	return
	//}
	//source := ctx.Values().Get("source")
	source, err := api.getSourceByCtx(ctx)
	if err != nil {
		api.FAIL(ctx, "数据源错误", nil)
		return
	}
	reqJson := make(map[string]any)
	if err := ctx.ReadJSON(&reqJson); err != nil {
		api.Errorf(ctx, "JSON API Update Error: %v", err)
		api.FAIL(ctx, err.Error(), nil)
		return
	}
	var errs []string
	for k, v := range reqJson {
		ex, err := dynamicSource.TableExists(*source, k)
		if ex {
			if kvs, ok := v.(map[string]any); ok {
				pk, pkExists := kvs["@pk"]
				if !pkExists {
					pk = "id"
				} else {
					delete(kvs, "@pk")
				}
				_, idExists := kvs[convertor.ToString(pk)]
				_, idInexists := kvs[convertor.ToString(pk)+"{}"]
				if idExists || idInexists {
					if idInexists {
						if _, ok := kvs["id{}"].([]any); !ok {
							errs = append(errs, fmt.Sprintf("Table：%s 更新数据失败：当主键为数组格式时 请使用 {\"id{}\":[1,2,3]}", k))
							continue
						}
					}
					sql, args := generateUpdate(k, kvs, convertor.ToString(pk), idInexists)
					err, result := dynamicSource.QueryBySource(*source, sql, args)
					if err != nil {
						errs = append(errs, fmt.Sprintf("Table: %s 更新数据失败：%s,%v 错误：%v", k, sql, args, err))
					} else if r, ok := result.(bool); ok {
						if !r {
							errs = append(errs, fmt.Sprintf("Table: %s 更新数据失败：%s,%v", k, sql, args))
						}
					}
				} else {
					errs = append(errs, fmt.Sprintf("Tbale: %s 更新数据失败 缺少主键ID或使用@pk自定义主键名称; 更新数据：%v", k, kvs))
				}
			}
		} else {
			errs = append(errs, fmt.Sprintf("Table Not Exists: %s err=%v", k, err))
			api.Errorf(ctx, "JsonApiUpdate Table Not Exists : %s; err = %v ", k, err)
		}
	}
	errLen := len(errs)
	insertLen := len(reqJson)
	if errLen > 0 && errLen < insertLen {
		api.SUCCESS(ctx, "数据更新完成，但出现部分错误", iris.Map{
			"errors": errs,
		})
		return
	} else if errLen == insertLen {
		api.FAIL(ctx, "数据更新失败", iris.Map{
			"errors": errs,
		})
		return
	}
	api.SUCCESS(ctx, "数据更新完成", iris.Map{
		"errors": errs,
	})
	return
}

func (api *JSONApi) delete(ctx iris.Context) {
	//if !ctx.Values().Exists("source") {
	//	api.FAIL(ctx, "数据源错误", nil)
	//	return
	//}
	//source := ctx.Values().Get("source")
	source, err := api.getSourceByCtx(ctx)
	if err != nil {
		api.FAIL(ctx, "数据源错误", nil)
		return
	}
	reqJson := make(map[string]any)
	if err := ctx.ReadJSON(&reqJson); err != nil {
		api.Errorf(ctx, "JSON API Delete Error: %v", err)
		api.FAIL(ctx, err.Error(), nil)
		return
	}
	var errs []string
	for k, v := range reqJson {
		ex, err := dynamicSource.TableExists(*source, k)
		if ex {
			if kvs, ok := v.(map[string]any); ok {
				pk, pkExists := kvs["@pk"]
				if !pkExists {
					pk = "id"
				} else {
					delete(kvs, "@pk")
				}
				_, idExists := kvs[convertor.ToString(pk)]
				_, idInexists := kvs[convertor.ToString(pk)+"{}"]
				if idExists || idInexists {
					if idInexists {
						if _, ok := kvs["id{}"].([]any); !ok {
							errs = append(errs, fmt.Sprintf("Table：%s 删除数据失败：当主键为数组格式时 请使用 {\"id{}\":[1,2,3]}", k))
							continue
						}
					}
					sql, args := generateDelete(k, kvs, convertor.ToString(pk), idInexists)
					err, result := dynamicSource.QueryBySource(*source, sql, args)
					if err != nil {
						errs = append(errs, fmt.Sprintf("Table: %s 删除数据失败：%s,%v 错误：%v", k, sql, args, err))
					} else if r, ok := result.(bool); ok {
						if !r {
							errs = append(errs, fmt.Sprintf("Table: %s 删除数据失败：%s,%v", k, sql, args))
						}
					}
				} else {
					errs = append(errs, fmt.Sprintf("Tbale: %s 删除数据失败 缺少主键ID或使用@pk自定义主键名称; 更新数据：%v", k, kvs))
				}
			}
		} else {
			errs = append(errs, fmt.Sprintf("删除数据失败: Table Not Exists: %s err=%v", k, err))
			api.Errorf(ctx, "JsonApiDelete Table Not Exists : %s; err = %v ", k, err)
		}
	}
	errLen := len(errs)
	insertLen := len(reqJson)
	if errLen > 0 && errLen < insertLen {
		api.SUCCESS(ctx, "删除完成，但出现部分错误", iris.Map{
			"errors": errs,
		})
		return
	} else if errLen == insertLen {
		api.FAIL(ctx, "删除失败", iris.Map{
			"errors": errs,
		})
		return
	}
	api.SUCCESS(ctx, "删除完成", iris.Map{
		"errors": errs,
	})
	return
}

func (api *JSONApi) getSourceByCtx(ctx iris.Context) (*model.Source, error) {
	mark := ctx.Params().Get("mark")
	if mark == "" {
		return nil, errors.New("数据源未找到")
	}
	var source model.Source
	result := api.ORM.Where("source_mark = ?", mark).First(&source)
	if result.Error != nil {
		return nil, errors.New("数据源未找到")
	}
	return &source, nil
}

func generateInsert(table string, kvs map[string]any) (string, map[string]any) {
	size := len(kvs)
	keys := make([]string, size)
	values := make([]string, size)
	args := make(map[string]any, size)
	i := 0
	for field, value := range kvs {
		keys[i] = field
		values[i] = "@" + field
		args[field] = value
		i++
	}
	sql := fmt.Sprintf("insert into %s (%s) values (%s)", table, strings.Join(keys, ","), strings.Join(values, ","))
	log.Info("sql: %s, args: %v", sql, args)
	return sql, args
}

func generateUpdate(table string, kvs map[string]any, pk string, isIdArray bool) (string, map[string]any) {
	size := len(kvs) - 1
	fields := make([]string, size)
	args := make(map[string]any, size)
	i := 0
	for field, value := range kvs {
		if field != pk && field != pk+"{}" {
			fields[i] = "`" + field + "`=@" + field
			args[field] = value
			i++
		}
	}
	var sql string
	if isIdArray {
		sql = fmt.Sprintf("update %s set %s where %s in(%v)", table, strings.Join(fields, ","), pk, kvs[pk+"{}"])
	} else {
		sql = fmt.Sprintf("update %s set %s where %s=%v", table, strings.Join(fields, ","), pk, kvs[pk])
	}
	log.Info("sql: %s, args: %v", sql, args)
	return sql, args
}

func generateDelete(table string, kvs map[string]any, pk string, isIdArray bool) (string, map[string]any) {
	var sql string
	args := make(map[string]any)
	if isIdArray {
		sql = fmt.Sprintf("delete from %s where %s in(@id)", table, pk)
		args["id"] = kvs[pk+"{}"]
	} else {
		sql = fmt.Sprintf("delete from %s where %s=@id", table, pk)
		args["id"] = kvs[pk]
	}
	return sql, args
}
