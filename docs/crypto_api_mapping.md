# CryptoService SDK 与 Server API 映射文档

本文档说明 `openim-sdk-core` 中 Crypto 接口与 `open-im-server` HTTP API 的对应关系。

## 服务端 HTTP API 路由

位置: `open-im-server/internal/api/router.go:354-363`

```go
cryptoGroup := r.Group("/crypto")
cryptoGroup.POST("/register_device", cr.RegisterDevice)
cryptoGroup.POST("/get_devices", cr.GetDevices)
cryptoGroup.POST("/revoke_device", cr.RevokeDevice)
cryptoGroup.POST("/get_virgil_jwt", cr.GetVirgilJWT)
cryptoGroup.POST("/get_group_key_version", cr.GetGroupKeyVersion)
cryptoGroup.POST("/get_group_key_events", cr.GetGroupKeyEvents)
cryptoGroup.POST("/security_precheck", cr.SecurityPrecheck)
cryptoGroup.POST("/integrity_report", cr.IntegrityReport)
```

## SDK 接口映射表

| 序号 | 服务端 HTTP API | SDK API 定义 | SDK 对外接口 | 用途 |
|------|----------------|-------------|-------------|------|
| 1 | `POST /crypto/register_device` | `api.CryptoRegisterDevice` | `CryptoRegisterDevice()` | 注册设备以获取 Virgil JWT |
| 2 | `POST /crypto/get_devices` | `api.CryptoGetDevices` | `CryptoGetDevices()` | 获取用户的所有设备列表 |
| 3 | `POST /crypto/revoke_device` | `api.CryptoRevokeDevice` | `CryptoRevokeDevice()` | 吊销设备（无法再获取 JWT） |
| 4 | `POST /crypto/get_virgil_jwt` | `api.CryptoGetVirgilJWT` | `CryptoGetVirgilJWT()` | 获取短期 Virgil JWT token |
| 5 | `POST /crypto/get_group_key_version` | `api.CryptoGetGroupKeyVersion` | `CryptoGetGroupKeyVersion()` | 查询群密钥版本号 |
| 6 | `POST /crypto/get_group_key_events` | `api.CryptoGetGroupKeyEvents` | `CryptoGetGroupKeyEvents()` | 获取群密钥变更事件 |
| 7 | `POST /crypto/security_precheck` | `api.CryptoSecurityPrecheck` | `CryptoSecurityPrecheck()` | 安全检查（设备状态校验） |
| 8 | `POST /crypto/integrity_report` | `api.CryptoIntegrityReport` | `CryptoIntegrityReport()` | 上报设备完整性证明 |

## 重要说明

### 1. BumpGroupKeyVersion 仅限内部调用

`BumpGroupKeyVersion` 接口**不在 HTTP API 中暴露**，仅通过内部 gRPC 调用（Group Service → Crypto Service）。

- ❌ **已从 SDK 中移除**: `api.CryptoBumpGroupKeyVersion`
- ✅ **服务端内部使用**: `internal/rpc/group/group.go` 在群成员变更后调用 `cryptoClient.BumpGroupKeyVersion()`

**原因**: 防止客户端恶意触发群密钥轮转 DoS 攻击。

### 2. SDK 实现结构

```
openim-sdk-core/
├── pkg/api/crypto.go              # API 端点定义（使用 newApi 泛型）
├── internal/crypto/crypto.go       # 业务逻辑层（自动填充 loginUserID）
├── open_im_sdk/crypto.go          # 对外接口（支持回调方式调用）
└── protocol/crypto/                # Protobuf 定义（与 server 共享）
    ├── crypto.proto
    ├── crypto.pb.go
    └── crypto_grpc.pb.go
```

### 3. SDK 调用示例

#### Go 层直接调用

```go
import "github.com/openimsdk/openim-sdk-core/v3/internal/crypto"

cryptoClient := crypto.NewCrypto(loginUserID)

// 注册设备
resp, err := cryptoClient.RegisterDevice(ctx, &crypto.RegisterDeviceReq{
    DeviceID:    "device-123",
    Platform:    "iOS",
    DeviceModel: "iPhone 14",
    AppVersion:  "3.5.0",
})

// 获取 Virgil JWT
jwtResp, err := cryptoClient.GetVirgilJWT(ctx, &crypto.GetVirgilJWTReq{
    DeviceID: "device-123",
})
```

#### C 接口回调方式（供 Flutter/React Native 等调用）

```go
import "github.com/openimsdk/openim-sdk-core/v3/open_im_sdk"

// 注册设备
open_im_sdk.CryptoRegisterDevice(callback, operationID, `{
    "deviceID": "device-123",
    "platform": "iOS",
    "deviceModel": "iPhone 14",
    "appVersion": "3.5.0"
}`)

// 获取 Virgil JWT
open_im_sdk.CryptoGetVirgilJWT(callback, operationID, `{
    "deviceID": "device-123"
}`)
```

### 4. 自动填充字段

SDK 会自动填充以下字段（客户端无需传递）：

- `RegisterDevice/GetDevices/RevokeDevice/GetVirgilJWT/SecurityPrecheck/IntegrityReport`: 自动填充 `req.UserID = loginUserID`
- ~~`BumpGroupKeyVersion`: 自动填充 `req.OperatorUserID = loginUserID`~~ (已移除)

### 5. 权限校验

服务端在以下接口中进行权限校验（`authverify.CheckAccessV3`）：

- `RegisterDevice`: 只能注册自己的设备或管理员操作
- `GetDevices`: 只能查询自己的设备列表或管理员操作
- `RevokeDevice`: 只能吊销自己的设备或管理员操作
- `GetVirgilJWT`: 只能获取自己设备的 JWT 或管理员操作
- `SecurityPrecheck`: 只能检查自己的设备或管理员操作

群相关接口 (`GetGroupKeyVersion`, `GetGroupKeyEvents`) 需要通过 HTTP token 认证，但不验证群成员身份（在服务端这是有意设计，避免循环依赖）。

## 相关文档

- 服务端设计文档: `open-im-server/docs/virgil-e2ee-single-group-minimal-design.md`
- 服务端实现: `open-im-server/internal/rpc/crypto/crypto.go`
- Code Review: `open-im-server/.cursor/plans/cryptoservice_code_review_*.plan.md`

## 版本历史

- 2024-04-16: 初始版本，所有 8 个公开接口已实现
- 2024-04-16: 移除 `BumpGroupKeyVersion` 公开 API（仅限内部调用）
