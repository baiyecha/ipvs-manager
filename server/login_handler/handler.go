package login_handler

import (
	"baiyecha/ipvs-manager/server/jwt"
	"net/http"

	"baiyecha/ipvs-manager/constant"

	"github.com/labstack/echo/v4"
)

func Login(c echo.Context) error {
	name := c.FormValue("name")
	password := c.FormValue("password")
	if name != "admin" || password != "123456" {
		return c.String(http.StatusOK, "密码错误！")
	}
	// 初始化cookie对象
	var maxAge = 3600 // 有效期，秒
	cookie := new(http.Cookie)
	cookie.Name = constant.CookieName
	var err error
	cookie.Value, err = jwt.GenerateToken(maxAge)
	if err != nil {
		return err
	}
	cookie.Path = "/"
	// cookie有效期为3600秒
	cookie.MaxAge = maxAge

	// 设置cookie
	c.SetCookie(cookie)
	return c.Redirect(http.StatusMovedPermanently, "/table")
}

func Logout(c echo.Context) error {
	// 初始化cookie对象
	cookie := new(http.Cookie)
	// 删除cookie只需要设置cookie名字就可以
	cookie.Name = constant.CookieName
	// cookie有效期为-1秒，注意这里不能设置为0，否则不会删除cookie
	cookie.MaxAge = -1

	// 设置cookie
	c.SetCookie(cookie)
	return c.Redirect(http.StatusMovedPermanently, "/")
}
