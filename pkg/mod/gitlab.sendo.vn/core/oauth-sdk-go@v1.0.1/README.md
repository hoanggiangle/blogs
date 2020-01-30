# OAuth SDK for Go

## Getting Started

Bước #1: Thêm OAuth SDK for Go vào danh sách dependencies trong file
`Gopkg.toml` - lưu ý `branch = "v2"`.

```toml
[[constraint]]
  name = "gitlab.sendo.vn/core/oauth-sdk-go"
  branch = "v2"
```

Bước #2: Dùng `dep ensure -v` để mang source code mới nhất của
OAuth SDK for Go về sử dụng.

Bước #3: Đăng ký `soauth` vào `resprovider`

```
# ./appsrc/resprovider/resource.go

type ResourceProvider interface {
    ...
    GetUserUsingToken(string) (soauth.UserStruct, error)
    ...
}

type myResourceProvider struct {
    ...
    oauthService soauth.OAuthService
    ...
}

func newRP(app sdms.Application) *myResourceProvider {
    ...
    // OAuth Service
    rp.oauthService = soauth.NewOAuthService(&soauth.OAuthConfig{
        App: app,
    })
    app.RegService(rp.oauthService)
    ...
}

func (m *myResourceProvider) GetUserUsingToken(token string) (soauth.UserStruct, error) {
    return m.oauthService.GetUserUsingToken(token)
}
```

Bước #4: Cập nhật `.env` để kết nối đến OAuth Service, tùy môi trường và giao thức:

*Sử dụng giao thức gRPC*

```
# .env

## Application ID for OAuth Service, provided by Account team. (-oauth-app-id)
OAUTH_APP_ID=d22c3e34-33a0-40ee-8827-23b2ea930782

## Choose one of 'grpc' or 'http'; default is 'grpc'. (-oauth-protocol)
OAUTH_PROTOCOL="grpc"

## Connecting URI for OAuth Service, provided by SysAdmin team. (-oauth-uri)
OAUTH_URI=192.168.1.12:30000
```

*Sử dụng giao thức HTTP (không khuyến khích)*

```
# .env

## Application ID for OAuth Service, provided by Account team. (-oauth-app-id)
OAUTH_APP_ID=d22c3e34-33a0-40ee-8827-23b2ea930782

## Choose one of 'grpc' or 'http'; default is 'grpc'. (-oauth-protocol)
OAUTH_PROTOCOL="http"

## Connecting URI for OAuth Service, provided by SysAdmin team. (-oauth-uri)
OAUTH_URI=http://192.168.1.12:38080
```

Bước #5: Sử dụng hàm `GetUserUsingToken` để lấy thông tin người dùng, ví dụ:

```go
package scripts

import (
    "gitlab.sendo.vn/services/account-gateway-service/appsrc/resprovider"
)

type ExampleScript struct {
    // Nothing goes here...
}

func (s *ExampleScript) InitFlags() {
    // Nothing goes here...
}

func (s *ExampleScript) Configure() error {
    return nil
}

func (s *ExampleScript) Cleanup() {
    // Nothing goes here...
}

func (s *ExampleScript) Run() error {

    log := resprovider.GetInstance().Logger("ExampleScript")

    accessToken := "Y3/j56uaRJpNZpa4u9aMvrOVGHkClOK2biloPKNp7JKOOPPs68sedwqwP80RDf+3mxoPHr2Q5NpZY6f/0KDerY7vnVK8GvQ3la8/wMbMfAHI0n+2Qy21vpE9lkZIWhij89WwgjvOdLnLthjU7A+3ZsglVQDQuK17wcFzu73vCKc="

    userObject, err := resprovider.GetInstance().GetUserUsingToken(accessToken)

    log.Infof("userObject: %v", userObject)
    log.Infof("err: %v", err)

    return nil

}

func (s *ExampleScript) Stop() {}
```

*Tham khảo thêm những thông tin mà hàm `GetUserUsingToken` cung cấp:*

```
package soauth

type UserStruct struct {
    CustomerID uint64 `json:"customer_id" bson:"customer_id"`
    FPTID      uint64 `json:"fpt_id" bson:"fpt_id"`

    Email string `json:"email" bson:"email"`
    Phone string `json:"phone" bson:"telephone"`

    CheckoutVerified bool `json:"checkout_verified" bson:"checkout_verified"`
    EmailVerified    bool `json:"email_verified" bson:"email_verified"`
    PhoneVerified    bool `json:"phone_verified" bson:"phone_verified"`

    FirstName string `json:"first_name" bson:"first_name"`
    LastName  string `json:"last_name" bson:"last_name"`

    Avatar   string `json:"avatar" bson:"avatar"`
    Birthday uint64    `json:"birthday" bson:"dob"`
    Gender   uint64    `json:"gender" bson:"gender"`

    CreatedAt uint64 `json:"created_at" bson:"created_at"`
    UpdatedAt uint64 `json:"updated_at" bson:"updated_at"`
}
```
