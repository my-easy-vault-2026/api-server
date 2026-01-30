# 开发规范

## 准备工作

1、准备好Golang环境

2、swagger：https://github.com/swaggo/swag?tab=readme-ov-file#descriptions-over-multiple-lines

3、本地调试: mysql/redis等

4、https://dashboard.ngrok.com/ localhost -> public domain

## 開發規範細節

### coding style
可以參考uber coding guideline
[https://github.com/ianchen0119/uber_go_guide_tw](https://github.com/ianchen0119/uber_go_guide_tw)

### 錯誤碼
每個service應有自己的錯誤碼，即使錯誤碼定義重複也要在自己的service建立錯誤碼

### 共用變數
1. 如果變數是public的，一律放在common資料夾
2. 有意義的常數或字串都必須設為const，不要直接用原生型別定義，例如status=1，要改為status=common.STATUS_ACTIVE
3. 不要使用iota
4. 數值從1開始定義，字串也不要定義空字串

### 三層架構
1. 分為handler, service, dao三層
2. handler只處理網路與顯示(加解密,驗證,前端資料轉換)，service層只處理業務邏輯，dao層只處理sql與redis增刪查改
3. 同一層之間不可互相引用(ex: service A不能引用service B)，因為這個情況，同樣的業務邏輯可能需複製到不同的service使用

### 資料層
1. 每個表有獨立的dao，我們不允許跨表操作，如果需跨表請拆分成多個查詢
2. 每個查詢必須分別撰寫func(ex: GetByID(id))，不可使用通用的查詢
3. 只有update操作可以通用，但是必須BY ID來更新(如update中的reflect不懂可以問josh)
4. 建議將所有資料欄位都加上`gorm:"default:null"`，我們遵循golang主流規範，會將0和空字串當成不存在。如果需要特定辨別0和nil的差異，請改用指標(ex: *int)

### 帳變
1. 系統中每個bill type只能出現一次，不可重用
2. 由於系統帳戶會是所有請求搶佔的對象，所以帳變必須放在transaction的最後，系統帳戶的帳變要放在帳變的最後

### 時間處理
由於伺服器的時區為+8, db時區為+0, gorm會誤將+8時區的時間當成+0來寫入資料庫，並且從資料庫取出資料時，也會誤將+0的時間當成+8，因此時間相關開發須遵守以下規範
1. `created_at`和`updated_at`需加上以下tag，由資料庫自行生成時間
```golang
	CreatedAt         time.Time `gorm:"default:null"`
	UpdatedAt         time.Time `gorm:"default:null;autoUpdateTime:false"`
```
2. 寫入與查詢時間時，需將時間透過`DBQueryTime()`進行轉換
```golang
// 更新與寫入
ClearedAt:    utils.DBQueryTime(form.Data.ClearedDate),

// 查詢
records, current, size, total, err = as.billDao.PageByUserIDInAssetIDCurrency(
  ctx,
  userIDIn,
  form.AssetID,
  form.Currency,
  utils.DBQueryTime(time.UnixMilli(form.CreatedAtFrom)),
  utils.DBQueryTime(time.UnixMilli(form.CreatedAtTo)),
  form.Current,
  form.PageSize,
)
```
3. 從資料庫中讀取物件出來時，須將讀出的物件做時區轉換
```golang
// 轉換整個物件
user, err := userDao.GetByID(ctx, id)
user = utils.ObjectFromDB(user)

// 轉換單筆時間
record, err := transactionRecordDao.GetVyTransactionNO(ctx, orderNO)
refundedAt := utils.TimeFromDB(record.RefundedAt)
```

### git workflow
1. 創建分支: 從master拉出feature branch
2. 測試: 本地測完沒有問題後，合併進入dev，在線上環境測試
3. 合併請求: 測試完畢沒有問題後，發mr給主管來合併到master，再進行線上測試
資料庫(sql-migrate)也同理，請參考sql-migrate的README

## 文件结构

```
.
├── common
│   ├── const.go // 常量
│   ├── error.go // 异常处理
│   ├── func.go // 常量轉換函數
│   └── struct.go  // 共用結構
├── conf // 配置文件
│   ├── config-dev.yaml
│   ├── config-local.yaml
│   ├── config-prod.yaml
│   ├── config-test.yaml
│   └── config.yaml
├── dao // 数据层（数据操作+数据模型）
│   ├── account.go
|   ...
├── docs // 文档 自动生成的
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── entities // 请求响应层
│   ├── account_form.go
│   ├── account_vo.go
│   ...
├── go.mod
├── go.sum
├── handlers // MVC层
│   ...
│   └── webhook
├── logger // 日誌
│   ├── logger.go
│   └── logger_test.go
├── main.go
├── middleware.go // 中介軟體
├── readme.md
├── router.go // 路徑
├── services // 逻辑层
│   ├── account.go
│   
├── tasks // 调度任务
│   ├── c2c_order.go
│   ├── deposit_order.go
│   ├── order.go
│   ├── trader.go
│   ├── wallet.go
│   └── withdraw_order.go
└── utils // 常用方法
    ├── aes.go
    ...
```

## 配置

### coinface
coinface api key:

aB3dE5fG7hI9jK1lM3nO5pQ7rS9tU0vX

coinface的api key

### mailgun

Verification public key
pubkey-b424c393e942c18bbf24331590f3c1f2

HTTP webhook signing key
91a2094a20b059a29c70f9cf61ee9751

API key：6243ecb569e1911b44dc2deea70cd091-51356527-0af0615a


```
using System;
using System.IO;
using RestSharp;
using RestSharp.Authenticators;
public class SendSimpleMessageChunk
{
 public static void Main (string[] args)
 {
  Console.WriteLine (SendSimpleMessage ().Content.ToString ())
 }
 public static IRestResponse SendSimpleMessage ()
 {
  RestClient client = new RestClient ();
  client.BaseUrl = new Uri ("https://api.mailgun.net/v3");"
  client.Authenticator ='
   new HttpBasicAuthenticator ("api",
    "YOUR_API_KEY");
  RestRequest request = new RestRequest ();
  request.AddParameter ("domain", "YOUR_DOMAIN_NAME", ParameterType.UrlSegment);
  request.Resource = "{domain}/messages";
  request.AddParameter ("from", "Excited User <mailgun@YOUR_DOMAIN_NAME>");
  request.AddParameter ("to", "bar@example.com");
  request.AddParameter ("to", "YOU@YOUR_DOMAIN_NAME");
  request.AddParameter ("subject", "Hello");
  request.AddParameter ("text", "Testing some Mailgun awesomeness!");
  request.Method = Method.POST;
  return client.Execute (request);
 }
}
```

## coinsdo

api key: 280306711d234cfa
api secret: 5c79569f08604f1e9ffa0156bce54ca6

文档地址：
https://coinsdo.gitbook.io/docs/v/cn/coinsdo-api-integration/coinsdo-api-jie-ru-wen-dang

回调服务器ip: 54.169.141.185

coinsend公私鑰(測試環境)
```
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnU8tN7EPHwI8zRY75SbO
tYuqWMi7JPF5HSvzuHtuNHJw8cEpJnsffsMaTfcG/C/XdATGXWfu4BmnKMM8R5SP
jZEzMVMCLpfdxindwbCH+gR2kPJqwQN0gyNapXChZUvnRAaikGMWIYB1BWI/eOWJ
vK62TtB3inzQqAaBjFneL5buj54yroAFcQvUpORUbdtIhnt5fmyVsJByaO7HgL73
vbjoL36cjSLgZ49aZ+wEao6z5kBGGHXTZL610jS+RtMtn7MGkjfiE21GoRtVf0UD
FD6Ct+nvM95RKJDnP0b5nE2ZbkxUyVHJNksv0Tk58QAj1UXRMOWwDiMHkqRx/oDY
aQIDAQAB
-----END PUBLIC KEY-----

-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCdTy03sQ8fAjzN
FjvlJs61i6pYyLsk8XkdK/O4e240cnDxwSkmex9+wxpN9wb8L9d0BMZdZ+7gGaco
wzxHlI+NkTMxUwIul93GKd3BsIf6BHaQ8mrBA3SDI1qlcKFlS+dEBqKQYxYhgHUF
Yj945Ym8rrZO0HeKfNCoBoGMWd4vlu6PnjKugAVxC9Sk5FRt20iGe3l+bJWwkHJo
7seAvve9uOgvfpyNIuBnj1pn7ARqjrPmQEYYddNkvrXSNL5G0y2fswaSN+ITbUah
G1V/RQMUPoK36e8z3lEokOc/RvmcTZluTFTJUck2Sy/ROTnxACPVRdEw5bAOIweS
pHH+gNhpAgMBAAECggEAIemuTVnJ8Tzpv6rh6bJWkNPVuwM3OS0nl47goW5WoU7k
o3GpfebAMPM9qf4tztM/hv76hqt/12cgXszMI4BW/EWHvEPxbfsGcBCZgoacod0x
dGkWc5rblOPbyFvCJ5TX/BXUGP2LiVhooer+1QDjEz61BcOyabQjxX11kzzShf6U
lYTFGiZCwtPc82YDy0h6sTDxLVewcxL0WkSaXV2CJsx8rQoI0pQXVSSWSFpUkOpF
/UKTiLbJ1LlUA4IGEJslYYWPvAAkyfW9ZJrNelJ2hyiC8ici4VhEuCLnM+IKWhDL
yxbxv/tI/kuzDIft1LhZTQ+eEPFA0VEuFxBOAAWXKQKBgQDYUxMMej/7ccvZzobr
Ac0qIsFd5+X6GilopgFaDTR4FhTyn4sQ9MBnjv9DbUxNRJD1qZ+Nspa+FOhlbUE2
USjPnIJaHW6wUC1dbQ5hxiUbdP1g74YJccYrMVs+c6LQlezMxtlGkQAlCdaAnulv
Q7qgFLQgXA6ndcsPKEXmcRUXFQKBgQC6KTQn4bEK1oiLAjuLC9qyP/dFruq9Rpjp
wi9zxoxdWWhNMkLbPwDRswMIE65Uh3MFztBgyQ0bEIZr2SQr2zm4GV6MHvmBKPix
/w6N2oUPHsTeARD0j+ROuhvXlWoemrW+6aBSGw4xOQJRR4xLfnmUk+WuUh7tOXYf
bBMO/BMRBQKBgQChsI5nYCTcs3Tz6tuLYoBQQ0QXBZMu+kkDMDmIbqBONesYYknW
tanufcKsSlCi3GIhTNS2W8sybnw5+4ynpcgETe5cnu0yGeuejjoWuLzZpfsRblbY
TlMZy71wk4wZrkYd1W9nwE/EX3MWFjFS+ePPbUopecV2Q6QwQyDkGpfx9QKBgHxB
CkXgV0oTnXmjGNkbJXK6TTJeqOGC4IeODBwrlv6rsXltJcCvEb3lzQ00DbTv328t
9lnTeALrib0sZv86yRC/JiNCfWifTzeHNVCrXQqVj/NaJNYHwOxnPjQrz3Pz8YEm
8NI8qsFh+tEDf3nYRhBMkw5CU9Ak/VnFygbDa3p9AoGBAJun5WsvY2UDT8Yq6dbc
3Usdj6yr3fLa2uErJhv8M8eWW2qutIOdVEZ3LxaJue8Ah642EfLnjqdVYXaWjRPN
zWzHNPCBH4M6w6wwL1VgPlxljE2lphEclUSLyRjRAAyPEjNTl32P6ylBBspNICKe
7q0BJ+PTq4fr/k7aWv7kbCF8
-----END PRIVATE KEY-----
```

## 備註

```
go build .
go run . -c conf/config-local.yaml
./api-server -p conf/ -c config-local.yaml -logPath ./log -logToStderr true
./api-server -p conf/ -c config-ngrok.yaml -logPath ./log -logToStderr true
```

1. stringer
  ```
  go get golang.org/x/tools/cmd/stringer
  ```

2. swaggo
  ```
  go install github.com/swaggo/swag/cmd/swag@latest
  swag init --parseDependency
  ```

3. ngrok
  ```
  brew install ngrok/ngrok/ngrok
  ngrok config add-authtoken 2i26gVbSZel7HkdRgiWfjZnzu3g_3aSgzXxPepPXhuAxztA2G
  ngrok http http://localhost:9487

  docker run -it -e NGROK_AUTHTOKEN=2i26gVbSZel7HkdRgiWfjZnzu3g_3aSgzXxPepPXhuAxztA2G ngrok/ngrok http 9487
  ```

4. gocron
  ```shell
  docker run --name gocron -p 5920:5920 -d ouqg/gocron
  ```
  https://127.0.0.1:5920
  ```shell
  # 取得本機ip
  ipconfig getifaddr en0 
  ```

# 分支管理方法

- develop开发中的分支
- feature/xx功能分支 紧急发布
- master发布分支，最新可发布代码，发布后回流到develop分支
- hotfix/xx紧急发布