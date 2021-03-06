package model

import (
	"time"
)
/*************************/
/********菜单路由结构体*********/
/*************************/
type Route struct {
	Id        int       `json:"id"`                              //ID
	AppId     string    `json:"appid" binding:"required,max=32"` //所属应用
	Name      string    `json:"name" binding:"required,min=1"`   //名称
	Route     string     `json:"item" binding:"required"`        //路由地址
	Type      int       `json:"action" binding:"required,min=1"` //路由类型
	Parent    int       `json:"parent" binding:"required,min=-1"` //父级id
	Priority  int       `json:"priority" binding:"required"`     //权重
	Schema    interface{}    `json:"schema"`                          //参数配置
	SchemaTo  string    `json:"_schema"`                          //接口返回改字段
	Remark    string    `json:"remark"`                          //描述
	CreatedAt time.Time `json:"created_at"`                      //创建时间
	UpdatedAt time.Time `json:"updated_at"`                      //更新时间
}
// Menu Router query condition
// type RouteWhere struct {
// 	Name      string   `json:"name,omitempty"`
// 	CreatedAt DateRang `json:"created_at,omitempty"`
// 	UpdatedAt DateRang `json:"updated_at,omitempty"`
// }
// // 菜单列表查询体参数
// // Search menu list by condtion
// type RouteQueryBody struct {
// 	QueryParams
// 	Where RouteWhere `json:"where"`
// }
// 更新menu router
type RouteUpdate struct {
	Name     string      `json:"name"`
	Route    string      `json:"item"`
	Parent   int         `json:"parent" binding:"required"`
	Priority int         `json:"priority"`
	Schema   interface{} `json:"schema"`
}