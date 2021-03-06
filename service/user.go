// User module
package service

import (
	// "fmt"
	"strconv"
	"time"

	"portal/database"
	"portal/config"
	"portal/model"
	"portal/util"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)
// 用户登录
// 验证用户信息,生成token
func Signin(email, passwd string) (int, interface{}) {
	user, err := database.Signin(email)
	// 账户不存在
	if err != nil {
		return 10001, "账户不存在"
	}
	// 审核状态
	switch user.CheckStatus {
		case 1:
			return 10003, "账户未审核"
		case 3:
			return 10004, "账户未通过审核"
	}
	// 用户状态
	switch user.Status {
	case 2:
		return 10005, "账户已禁用"
	case 3:
		return 10006, "账户已注销"
	}
	// 比较密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwd)); err != nil {
		return 10002, "用户名或密码不正确"
	}
	// token
	token, err := generateToken(strconv.Itoa(user.Id), strconv.Itoa(user.RoleId))
	// set cookie
	user.Token = token
	user.Email = email
	return 0, user
}
// claims
// type MyClaims struct {
// 	UserId string `json:"userId,omitempty"`
// 	RoleId string `json:"roleId,omitempty"`
// 	jwt.StandardClaims
// }
// Generate Token
func generateToken(userId, roleId string) (string, error) {
	AppConf := config.AppConfig
	maxAge, _ := strconv.ParseInt(strconv.Itoa(AppConf.TokenMaxAge), 10, 64)
	// 失效时间
	expireTime := time.Now().Add(time.Duration(maxAge)*time.Second)
	// 加密userid, roleid
	userId, _ = util.Encrypt([]byte(AppConf.AesKey), userId)
	roleId, _ = util.Encrypt([]byte(AppConf.AesKey), roleId)
	// Create the Claims
	claims := model.MyClaims{
		UserId: userId,
		RoleId: roleId,
		StandardClaims: jwt.StandardClaims {
			ExpiresAt: expireTime.Unix(),
			Issuer:    "yourlin127@gmail.com",
		},
	}
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(AppConf.TokenSecrect))
	// IF error
	if err != nil {
		return "",err
	}
	return ss, nil
}
// 用户注册
// @params 
//   User struct (post body)
// @return
//   code int
//   msg  error
func Signup(User model.SignupForm) (int, interface{})  {
	// 验证邮箱,手机号是否已注册
	where := "(`email` = ? OR `mobile` = ?) AND `deleted_at` = ?"
	if _, exsited := database.FindOneUser(where, User.Email, User.Mobile, model.DeletedAt); exsited {
		return 100010, "该邮箱或是手机号已注册"
	}
	// 密码加密
	hash, _ := bcrypt.GenerateFromPassword([]byte(User.Password), bcrypt.DefaultCost)
	User.Password = string(hash)
	// insert user data to user table
	err := database.AddUser(User)
	if (err != nil) {
		return 1, err
	}
	return 0, ""
}
// 查询用户列表
// @params
//   query: struct (query参数)
// @reutrn 
//   resutl [] *model.User
//   msg    error
func GetUserList(query model.UserQueryBody) ([]interface{}, error) {
	var (
		where string = "`u`.`status` "
		values []string
	)
	// 用户状态, status default value != 3
	// 1启用, 2禁用, 3注销
	if query.Where.Status != "" {
		where += " = ?"
		values = append(values, query.Where.Status)
	} else {
		where += " != ?"
		values = append(values, "3")
	}
	// 邮箱
	if query.Where.Email != "" {
		where += " AND `u`.`email` LIKE ?"
		values = append(values, query.Where.Email)
	}
	// 审核状态 1 未审核,2 审核通过,3 审核不通过
	if query.Where.CheckStatus != "" {
		where += " AND `u`.`check_status` = ?"
		values = append(values, query.Where.CheckStatus)
	}
	if query.Limit == 0 {
		query.Limit = 10
	}
	// select offset and limit
	where += " LIMIT ?, ?"
	// slice不能直接传递给interface slice
	params := make([]interface{}, len(values)+2)
	for i, v := range values {
		params[i] = v
	}
	// 加入分页
	params[len(values)] = query.Offset
	params[len(values) + 1] = query.Limit
	// Select user table
	res, err := database.FindAllUser(where, params...)
	if err != nil {
		return nil, err
	}
	return res, nil
}
// 用户状态变更, status 1:启用,2:禁用,3:注销
// @params 
//   id:     用户id
//   status: 状态值
//   remark: 操作描述
func UpdateUserStatus(id int, status int, remark string) (int, interface{}) {
	if code, _ := database.FindById(id, `portal_user`); code != 0 {
		return 1, "未找到该用户"
	}
	err := database.UpdateUserStatus(id, status, remark)
	// return
	if err != nil {
		return 1, err
	}
	return 0, nil
}
// 审核账户, check_status 1：未审核，2：通过，3：不通过
// @params 
//   id:           用户id
//   check_status: 状态值
//   check_remark: 描述
func ReviewUser(id int, check_status int, check_remark string) (int, interface{}) {
	if code, _ := database.FindById(id, `portal_user`); code != 0 {
		return 1, "未找到该用户"
	}
	err := database.ReviewUser(id, check_status, check_remark)
	// return
	if err != nil {
		return 1, err
	}
	return 0, nil
}
// 编辑用户
// @params
//   id string
//   form struct
func EditUser(id int, form model.EditUserForm) (int, interface{}) {
	if code, _ := database.FindById(id, `portal_user`); code != 0 {
		return 1, "未找到该用户"
	}
	// update data is not nil
	if form.Name != "" || form.Mobile != "" || form.Password != "" {
		sql := "UPDATE portal_user SET"
		// 用户名
		if form.Name != "" {
			sql += " `name` = " + `"` + form.Name + `"`
		}
		// 手机号
		if form.Mobile != "" {
			// 手机号是否已占用
			_, exsited := database.FindOneUser(`mobile = ? AND id != ? AND deleted_at = ?`, form.Mobile, id, model.DeletedAt)
			if exsited {
				return 1, "该手机号已占用"
			}
			if form.Name != "" {
				sql += ", `mobile` = " + `"` + form.Mobile + `"`
			} else {
				sql += " `mobile` = " + `"` + form.Mobile + `"`
			}
		}
		// password
		if form.Password != "" {
			hash, _ := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
			if form.Name != "" || form.Mobile != "" {
				sql += ", `password` = " + `"` + string(hash) + `"`
			} else {
				sql += " `password` = " + `"` + string(hash) + `"`
			}
		}
		sql += " WHERE `id` = ?"
		// update table
		err := database.EditUser(id, sql)
		if err != nil {
			return 1, err
		}
	}
	return 0, nil
}
// change password
// find user => verify oldpasswd => generate hash => update
func ChangePasswd(id int, oldPasswd, passwd string) (int, interface{}) {
	if code, _ := database.FindById(id, `portal_user`); code != 0 {
		return 1, "未找到该用户"
	}
	// verify oldPasswd
	hash, _ := database.GetPasswd(id)
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(oldPasswd)); err != nil {
		return 1, "原密码不正确"
	}
	// hash password
	hashPasswd, _ := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	// update password
	if err := database.ChangePasswd(id, string(hashPasswd)); err != nil {
		return 1, err
	}
	return 0, nil
}