# [gorm-plugin-crypto](https://github.com/kangarooxin/gorm-plugin-crypto)

通过Gorm插件，加解密数据，代码入侵率低。
- 默认内置了AES加解密策略
- 自定义多种加解密策略，通过tag指定策略

## 使用步骤:
1. 引入包
```go
go get -u github.com/kangarooxin/gorm-plugin-crypto
```

2.  注册插件，注册默认的加解密策略
```go
db, _ = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
db.Use(crypto.NewCryptoPlugin())
// 注册默认的AES加解密策略
RegisterCryptoStrategy(crypto.NewAesCryptoStrategy("1234567890123456"))
db.AutoMigrate(&User{})

```
3. 使用tag标记struct字段
```go
type User struct {
	ID     uint   `gorm:"primarykey"`
	Name   string `gorm:"column:name"`
	Age    int    `gorm:"column:age"`
	Email  string `gorm:"column:email" crypto:"aes"`
	Mobile string `gorm:"column:mobile" crypto:"aes"`
}
```
4. 自定义加解密策略，实现`CryptoStrategy`接口。
```go
// 注册加解密策略
crypto.RegisterCryptoStrategy(MyAesCryptoStrategy("1234567890123456"))

// 使用自定义的策略tag标记
Email  string `gorm:"column:email" crypto:"myAes"`
```

3. 开始使用:

#### 插入数据
```go
user1 := &User{ID: 1, Name: "User1", Age: 18, Email: "user1@example.com", Mobile: "13812345671"}
user2 := &User{ID: 2, Name: "User2", Age: 12, Email: "user2@example.com", Mobile: "13812345672"}
user3 := &User{ID: 3, Name: "User3", Age: 16, Email: "user3@example.com", Mobile: "13812345673"}

// 单个插入
db.Create(user1)
assert.Equal(t, "{AES}g+2fA4EbDGDPpZVCF0quCwjz4w22BRHHb0xqEG86zL0=", user1.Email)
assert.Equal(t, "{AES}Q/FDK7HDVHpArPRm3kCwEw==", user1.Mobile)

// 批量插入
users := []*User{user2, user3}
db.Create(users)
assert.Equal(t, "{AES}VNrSbyrCwXfcBIoxYbO8hgjz4w22BRHHb0xqEG86zL0=", user2.Email)
assert.Equal(t, "{AES}yRPUirBKNe9UFlIStzft/gjz4w22BRHHb0xqEG86zL0=", user3.Email)
```

#### 查询
```go
// 通过主键查询
var queryUser User
db.First(&queryUser, 1)
assert.Equal(t, "user1@example.com", queryUser.Email)
assert.Equal(t, "13812345671", queryUser.Mobile)

// 通过主键批量查询
var queryUsers []User
db.Find(&queryUsers, []int{1, 2, 3})
assert.Equal(t, "user1@example.com", queryUsers[0].Email)
assert.Equal(t, "13812345672", queryUsers[1].Mobile)

// 查询全部
var queryAllUsers []User
db.Find(&queryAllUsers)
assert.Equal(t, "user1@example.com", queryUsers[0].Email)
assert.Equal(t, "13812345672", queryUsers[1].Mobile)

// 通过自定义条件查询
var queryUser1 User
db.Where("name = ?", "User1").First(&queryUser1)
assert.Equal(t, "user1@example.com", queryUser1.Email)
assert.Equal(t, "13812345671", queryUser1.Mobile)

// 通过加密字段查询时，由于无法识别模型，所以需要手动加密数据，可以通过`NewCryptoValue` 包装实现加密。
var queryUser2 User
db.Where("email = ?", NewCryptoValue("email", "user1@example.com")).First(&queryUser2)
assert.Equal(t, "user1@example.com", queryUser2.Email)
assert.Equal(t, "13812345671", queryUser2.Mobile)

// 通过Struct条件查询
var queryUser4 User
db.Where(&User{Email: "user1@example.com"}).First(&queryUser4)
assert.Equal(t, "13812345671", queryUser4.Mobile)

// 通过Map查询
var queryUser5 User
db.Where(map[string]interface{}{"email": "user1@example.com"}).First(&queryUser5)
assert.Equal(t, "13812345671", queryUser5.Mobile)

// 通过Map In 查询
var queryUser6 []User
db.Where(map[string]interface{}{
"email": []string{"user1@example.com", "user2@example.com"},
}).Find(&queryUser6)
assert.Equal(t, "13812345671", queryUser6[0].Mobile)
assert.Equal(t, "13812345672", queryUser6[1].Mobile)

// 通过原生Sql查询
var queryUser3 User
db.Raw("select * from test_user").Find(&queryUser3)
assert.Equal(t, "user1@example.com", queryUser3.Email)
```

#### 保存和更新
```go
// 有主键的保存，执行更新操作
var saveUser User
db.First(&saveUser)
saveUser.Email = "User11@example.com"
db.Save(&saveUser)
assert.Equal(t, "{AES}siKVK6qMulucOlmRoZWLiWcZIqVzlNkqP58lypIfHtg=", saveUser.Email)

// 无主键的保存，执行插入操作
user4 := &User{Name: "User4", Age: 18, Email: "user4@example.com", Mobile: "13812345674"}
db.Save(user4)
assert.Equal(t, "{AES}g1WxCfYDcw/2k5g9kyFDpAjz4w22BRHHb0xqEG86zL0=", user4.Email)

//使用 `struct`更新
db.Model(&User{}).Where("id = ?", 1).Update("email", "user111@example.com")
var queryUser7 User
db.First(&queryUser7, 1)
assert.Equal(t, "user111@example.com", queryUser7.Email)

// 使用 `map`更新
db.Model(&User{}).Where("id = ?", 2).Updates(map[string]interface{}{"email": "user222@example.com"})
var queryUser8 User
db.First(&queryUser8, 2)
assert.Equal(t, "user222@example.com", queryUser8.Email)
```