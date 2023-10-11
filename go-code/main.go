package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string
	Username  string
	Email     string
	pswdHash  string
	CreatedAt string
	Active    string
	verHash   string
	timeout   string
}

var db *sql.DB

var store = sessions.NewCookieStore([]byte("super-secret"))

func init() {
	store.Options.HttpOnly = true
	store.Options.Secure = true
	gob.Register(&User{})
}

func main() {

	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/bgn")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	authRouter := router.Group("/user", auth)

	router.GET("/", indexhandler)
	router.GET("/login", loginGethandler)
	router.POST("/login", loginPosthandler)

	authRouter.GET("/profile", profileHandler)

	err = router.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}

}
func auth(c *gin.Context) {
	fmt.Println("auth middleware running")
	session, _ := store.Get(c.Request, "session")
	fmt.Println("Session:", session)
	_, ok := session.Values["user"]
	if !ok {
		c.HTML(http.StatusForbidden, "login.html", nil)
		c.Abort()
		return
	}
	fmt.Println("middleware done")
	c.Next()

}

func indexhandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func loginGethandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func loginPosthandler(c *gin.Context) {
	var user User
	user.Username = c.PostForm("username")
	password := c.PostForm("password")
	err := user.getUserByUsername()
	if err != nil {
		fmt.Println("Error selecting password in db bt username, err:", err)
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"message": "check username and password"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.pswdHash), []byte(password))
	fmt.Println("Error from bycrypt:", err)
	if err == nil {
		session, _ := store.Get(c.Request, "session")
		session.Values["user"] = user
		session.Save(c.Request, c.Writer)
		c.HTML(http.StatusOK, "loggedin.html", gin.H{"username": user.Username})
		return
	}
	c.HTML(http.StatusOK, "profile.html", gin.H{"user": user})
}

func profileHandler(c *gin.Context) {
	session, _ := store.Get(c.Request, "session")
	var user = &User{}
	val := session.Values["user"]
	var ok bool
	if user, ok = val.(*User); !ok {
		fmt.Println("was not of type *User")
		c.HTML(http.StatusForbidden, "login.html", nil)
		return
	}
	c.HTML(http.StatusOK, "profile.html", gin.H{"user": user})
}

func (u *User) getUserByUsername() error {
	stmt := "select * from users where username = ?"
	row := db.QueryRow(stmt, u.Username)
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.pswdHash, &u.CreatedAt, &u.Active, &u.verHash, &u.timeout)
	if err != nil {
		fmt.Println("getUser() error selecting User , err:", err)
		return err
	}
	return nil
}
