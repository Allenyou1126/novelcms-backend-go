package models

import (
	"backend-go/utils"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log/slog"
)

type User struct {
	ID             int64    `gorm:"column:uid;primaryKey;" json:"uid"`
	Name           string   `gorm:"column:username" json:"username"`
	SaltedPassword string   `json:"-"`
	Salt           string   `json:"-"`
	Mail           string   `json:"-"`
	Sign           string   `gorm:"column:sign;not null;default:这个人很懒，还没写签名" json:"sign"`
	AvatarUrl      string   `gorm:"column:avatar_url;not null;default:'https://localhost/'" json:"avatar_url"`
	Permissions    []string `gorm:"type:string;serializer:json"`
}

var defaultPermissions = [...]string{"basic", "article", "collection", "comment", "user"}

var uidGen *utils.IdGenerator = nil

var DefaultUser = User{
	ID:          0,
	Name:        "",
	Permissions: make([]string, 0),
}

func CreateUser(name string, mail string, hashedPassword string) User {
	if uidGen == nil {
		g := utils.GetIdGenerator(0, 4, 0)
		uidGen = &g
	}
	saltBuf := make([]byte, 2048)
	_, err := rand.Reader.Read(saltBuf)
	if err != nil {
		slog.Warn("Generate salt failed. Use empty salt instead.")
		saltBuf = make([]byte, 2048)
	}
	salt := fmt.Sprintf("%X", sha256.Sum256(saltBuf))
	var per = make([]string, len(defaultPermissions))
	copy(per, defaultPermissions[0:])
	u := User{
		ID:          uidGen.Next(),
		Name:        name,
		Mail:        mail,
		Salt:        salt,
		Permissions: per,
	}
	u.SetPassword(hashedPassword)
	return u
}

func (u *User) SetPassword(hashedPassword string) string {
	v := fmt.Sprintf("%X", sha256.Sum256([]byte(u.Salt+hashedPassword)))
	u.SaltedPassword = v
	return v
}
