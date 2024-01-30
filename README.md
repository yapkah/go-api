# SEC API

Modified from repository [eddycjy/go-gin-example](https://github.com/eddycjy/go-gin-example) ([Docs](https://github.com/EDDYCJY/go-gin-example/blob/master/README_ZH.md))

# Installation

```go
go get github.com/yapkah/go-api
```

# Getting Started

## 1. app.ini

copy `conf/app.sample.ini` to another file name `conf/app.ini` and change the setting inside

## 2. Generate public and private key

generate public and private key for access token using command below:

```sh
cd ./storage/
- login vers.
openssl genrsa -out private.pem 4096
openssl rsa -in private.pem -out public.pem -pubout -outform PEM
- password vers.
openssl genrsa -out encrypt_private.pem 2048
openssl rsa -in encrypt_private.pem -out encrypt_public.pem -pubout -outform PEM
- bc admin vers.
openssl genrsa -out bc_admin_private.pem 4096
openssl rsa -in bc_admin_private.pem -out bc_admin_public.pem -pubout -outform PEM
```

## 3. DB

import the sample table from `docs/sql/template.sql` to your DB

## 4. Install go modules

run `go install` at the root of the repo

# Modified files

- `middleware/jwt.go` (access token validation)
- `pkg/util/jwt.go` (token generation function)
- `pkg/setting/setting.go` (change `var Cfg` to be access by other package)
- `conf/app.ini` added new config in section `[custom]`
- `pkg/app/form` change `BindAndValid()` function
- `pkg/app/response` (for return response)

# Adding first admin / member

comment line

- `auth.Use(jwt.JWT())` (for token authentication)
- `memberGroup.Use(jwt.CheckScope("MEM"))` (for checking token scope)
- `adminGroup.Use(jwt.CheckScope("ADM"))` (for checking token scope)

```go
// authentication api group
auth := apiv1.Group("/")
auth.Use(jwt.JWT())
{
  // member api
  memberGroup := auth.Group("/member")
  memberGroup.Use(jwt.CheckScope("MEM")) // check token scope
  {
    memberGroup.GET("/", member.GetMember)  // get member
    memberGroup.POST("/", member.AddMember) // add member
  }

  // admin api
  adminGroup := auth.Group("/admin")
  adminGroup.Use(jwt.CheckScope("ADM")) // check token scope
  {
    adminGroup.GET("/", admin.GetAdmin)
    adminGroup.POST("/", admin.AddAdmin)
  }
}
```

## 5. Update app.ini file in /conf folder.

value to update under [custom] section

1. AppName
2. AdminServerDomain
3. MemberServerDomain
4. ApiServerDomain
5. CryptoSalt1
6. CryptoSalt2
7. CryptoSalt3
8. EncryptSalt
9. PKSalt
