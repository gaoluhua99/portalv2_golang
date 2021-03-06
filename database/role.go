package database

import (
	"fmt"
	"time"
	"portal/util"
	"portal/model"
)

var createRole = `INSERT INTO portal_role(name, remark, created_at, updated_at) VALUES(?, ?, ?, ?)`
var selectRole = `SELECT id, name, remark, created_at, updated_at FROM portal_role WHERE status = 1`
var selectUserByRole = `SELECT`           +              
												` u.id,`          +
												` u.name`         +
											`	FROM`               +
												`	portal_user AS u` +
												` WHERE`            +
													` u.id IN ( SELECT ur.user_id FROM portal_user_role AS ur WHERE ur.role_id = ? )` +
													` AND status != 3`
// Create role
func CreateRole(name, remark string) error {
	tx, err := ConnDB().Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(createRole, name, remark, time.Now().Format(util.TimeFormat), time.Now().Format(util.TimeFormat))
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
// Find Row by role name
func FindRoleByName(where string, query...interface{}) (bool, error) {
	type user struct {
		id int
		name string
	}
	var list = make([]interface{}, 0)
	Sql := `SELECT id, name FROM portal_role WHERE status = 1 AND `
	res, err := ConnDB().Query(Sql+where, query...)
	if err != nil {
		return false, err
	}
	for res.Next() {
		var ele = &user{}
		if err := res.Scan(
			&ele.id,
			&ele.name,
		); err != nil {
			return false, err
		} else {
			list = append(list, ele)
		}
	}
	
	return len(list) > 0, nil
}
// Update Role Info
func UpdateRole(where string, query...interface{}) error {
	var Sql string = `UPDATE portal_role SET `
	_, err := ConnDB().Exec(Sql+where, query...)
	if err != nil {
		return err
	}
	return nil
}
// Delete Role, set status = 2 
func DeleteRole(id int) (int, error) {
	Sql := `UPDATE portal_role SET status = ?, deleted_at = ? WHERE id = ?`
	stmt, err := ConnDB().Prepare(Sql)
	if err != nil {
		return 1, err
	}
	// exec sql
	_, err = stmt.Exec(2, time.Now().Format(util.TimeFormat), id)
	if err != nil {
		return 1, err
	}
	return 0, nil
}
// Find All User, Return Role List
func FindAllRole(where string, query ...interface{}) ([]interface{}, error) {
	var result = make([]interface{}, 0)
	rows, err := ConnDB().Query(selectRole + where, query...)
	fmt.Println(selectRole+where)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// 遍历行, 追加到result slice
	for rows.Next() {
		var	data = &model.Role{}
		if err = rows.Scan(
			&data.Id,
			&data.Name,
			&data.Remark,
			&data.CreatedAt,
			&data.UpdatedAt,
		); err != nil {
			return result, err
		} else {
			result = append(result, data)
		}
	}
	return result, nil
}
// Get user list by role id
func GetUserByRoleId(roleId int) ([]interface{}, error) {
	type List struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
	var userList = make([]interface{}, 0)
	rows, err := ConnDB().Query(selectUserByRole, roleId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Next
	for rows.Next() {
		var ele = &List{}
		if err := rows.Scan(&ele.Id, &ele.Name); err != nil {
			return nil, err
		}
		userList = append(userList, ele)
	}
	return userList, nil
}
// Migrate user to role group
func MigrateUser(roleId int, userId []int) error {
	Sql := 	`UPDATE portal_user_role SET role_id = ? WHERE user_id = ?`
	tx, err := ConnDB().Begin()
	if err != nil {
		return err
	}
	// update role_id
	for _, ele := range userId {
		_, err = ConnDB().Exec(Sql, roleId, ele)
		if err != nil {
			return err
		}
	}
  return tx.Commit()
}