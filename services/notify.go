package services

import (
	authDao "api-server/dao/auth"
	notifyDao "api-server/dao/notify"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"
	"context"
	"encoding/json"
	"math/rand"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"sync"
	"time"
	"unicode"

	"fmt"
	"strings"

	"google.golang.org/api/option"
	"gorm.io/gorm"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
)

type NotifyService struct {
	callbackTaskDao      *notifyDao.CallbackTaskDao
	callbackLogDao       *notifyDao.CallbackLogDao
	notifyTemplateDao    *systemDao.NotifyTemplateDao
	notifySendSettingDao *systemDao.NotifySendSettingDao
	parameterDao         *systemDao.ParameterDao
	userDao              *userDao.UserDao
	notifyRecordDao      *notifyDao.NotifyRecordDao
	tokenDao             *authDao.TokenDao
}

func NewNotifyService() *NotifyService {
	return &NotifyService{
		callbackTaskDao:      notifyDao.NewCallbackTaskDao(),
		callbackLogDao:       notifyDao.NewCallbackLogDao(),
		notifyTemplateDao:    systemDao.NewNotifyTemplateDao(),
		notifySendSettingDao: systemDao.NewNotifySendSettingDao(),
		parameterDao:         systemDao.NewParameterDao(),
		userDao:              userDao.NewUserDao(),
		notifyRecordDao:      notifyDao.NewNotifyRecordDao(),
		tokenDao:             authDao.NewTokenDao(),
	}
}

var client *messaging.Client
var once sync.Once

// 初始設定
func InitFirebase() error {
	var initErr error
	once.Do(func() {
		jsonData := []byte(utils.Config.Firebase.JsonFile)
		app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsJSON(jsonData))
		if err != nil {
			fmt.Printf("error initializing app: %v\n", err)
			initErr = err
			return
		}

		ctx := context.Background()
		client, err = app.Messaging(ctx)
		if err != nil {
			fmt.Printf("error getting Messaging client: %v\n", err)
			initErr = err
			return
		}
	})
	return initErr
}

func (ns *NotifyService) SendOTP(ctx context.Context, form *entities.SendOTPForm) error {

	vcode := utils.RandString(6, "0123456789")

	if ret1 := utils.RDB.Set(
		ctx,
		utils.GetOTPRedisKey(form.Purpose, common.NOTIFY_TYPE_EMAIL, vcode),
		form.Email,
		time.Minute*utils.Config.Notify.OTPExpireTime,
	); ret1.Err() != nil {
		logger.Warn("get failed,", ret1.Err())
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	notifyVO := &entities.NotifyVO{
		Code:   vcode,
		Expire: fmt.Sprintf("%d", utils.Config.Notify.OTPExpireTime),
	}

	sendVO := &entities.SendVO{
		Email: form.Email,
		Code:  common.MSG_OPCODE_OTP,
	}

	err := ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}

	return nil

}

func (ns *NotifyService) SendNewDeviceLogin(ctx context.Context, notifyVO *entities.NotifyVO, sendVO *entities.SendVO) error {

	err := ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}

	return nil
}

func (ns *NotifyService) Send3ds(ctx context.Context, email string, userID uint64, threedNotify *entities.Check3dsVO) error {

	var threedsString = ""
	if threedNotify != nil {
		// 將結構體編碼為 JSON
		jsonBytes, err := json.Marshal(threedNotify)
		if err != nil {
			logger.Warnf("send3ds convert json string fail: %v", err)
			return err
		}
		threedsString = strings.ReplaceAll(string(jsonBytes), `"`, `\"`)
	}

	notifyVO := &entities.NotifyVO{
		Threeds: threedsString,
	}

	sendVO := &entities.SendVO{
		Email:  email,
		Code:   common.MSG_OPCODE_REQUEST_3DS,
		UserID: userID,
	}

	err := ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}

	return nil
}

func (ns *NotifyService) SendNotification(ctx context.Context, notifyVO *entities.NotifyVO, sendVO *entities.SendVO) error {
	err := ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}
	return nil
}

func (ns *NotifyService) SendUserUnblock(ctx context.Context, notifyVO *entities.NotifyVO, sendVO *entities.SendVO) error {
	err := ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}
	return nil
}

func (ns *NotifyService) SendReapFailureBlock(ctx context.Context, email string, cardNO string, userID uint64) error {

	notifyVO := &entities.NotifyVO{
		CardNo: cardNO,
	}

	sendVO := &entities.SendVO{
		Email:  email,
		Code:   common.MSG_OPCODE_CARD_BLOCKED_CONSECUTIVE_FAILURE,
		UserID: userID,
	}

	err := ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}

	return nil
}

func (ns *NotifyService) SendReapFailureUnblock(ctx context.Context, email string, cardNO string, userID uint64) error {

	notifyVO := &entities.NotifyVO{
		CardNo: cardNO,
	}

	sendVO := &entities.SendVO{
		Email:  email,
		Code:   common.MSG_OPCODE_CARD_REINSTATE_AFTER_BLOCKED,
		UserID: userID,
	}

	err := ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}

	return nil
}

func (ns *NotifyService) SendCardTransactionOTP(ctx context.Context, notifyVO *entities.NotifyVO, sendVO *entities.SendVO) error {

	err := ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}

	return nil
}

func (ns *NotifyService) SendApplyCardSuccess(ctx context.Context, notifyVO *entities.NotifyVO, sendVO *entities.SendVO) error {

	err := ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}

	return nil
}

func (ns *NotifyService) SendApplyCardFail(ctx context.Context, notifyVO *entities.NotifyVO, sendVO *entities.SendVO) error {

	err := ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}

	return nil
}

func (ns *NotifyService) SendCardAmountLow(ctx context.Context, notifyVO *entities.NotifyVO, sendVO *entities.SendVO, cardID uint64) error {
	amountLowAlertKey := utils.GetCardAmountLowAlertKey(cardID)
	err := utils.RDB.Set(ctx, amountLowAlertKey, "sent", time.Hour*1).Err()
	if err != nil {
		logger.Warn("set card amount low alert key failed, cardID=%d, error=%v", cardID, err)
	}

	err = ns.sendNotify(ctx, sendVO, notifyVO)
	if err != nil {
		return err
	}

	return nil
}

func (ns *NotifyService) sendNotify(ctx context.Context, sendVO *entities.SendVO, notifyVO *entities.NotifyVO) error {
	// 取得語言
	var language common.Language

	// 先從用戶設定取得
	user, err := ns.userDao.GetByUserID(ctx, sendVO.UserID)
	if err != nil {
		logger.Warn("Get user failed,", err)
		return err
	}
	language = user.Language
	logger.Infof("send %v notify language from user = %v", sendVO.Code, language)
	if !language.IsValid() {
		language = common.LANGUAGE_ENGLISH
	}

	// 如果用戶沒有設定language，則從物件取得
	if language == "" {
		language = sendVO.Language
	}

	// 如果沒有取得 language，則預設語言為英文
	if language == "" {
		logger.Infof("send %v language is empty", sendVO.Code)
		language = common.LANGUAGE_ENGLISH
	}

	// 取得設定
	notifySetting, err := ns.notifySendSettingDao.GetNotifySetting(ctx, sendVO.Code)

	if err != nil {
		logger.Warn("Get notify setting failed,", err)
		return err
	}

	// TODO 發送簡訊
	// if notifySetting.Sms {

	// }

	// 發送 app notification - firebase
	if notifySetting.App {
		ns.sendApp(ctx, sendVO, notifyVO, language)
	}

	return nil
}

func (ns *NotifyService) convertSubject(ctx context.Context, notifyTemplate *systemDao.NotifyTemplate, notifyVO *entities.NotifyVO, language common.Language) (string, error) {
	subject := notifyTemplate.Subject

	// 替換模板中的佔位符
	placeholders := map[string]string{
		"{$CARD_NO}":    notifyVO.CardNo,
		"{$CARD_TITLE}": notifyVO.CardTitle,
		"{$CURRENCY}":   notifyVO.Currency,
	}

	for placeholder, value := range placeholders {
		if value != "" {
			subject = strings.Replace(subject, placeholder, value, -1)
		}
	}

	return subject, nil
}

func (ns *NotifyService) convertTemplate(ctx context.Context, notifyTemplate *systemDao.NotifyTemplate, notifyVO *entities.NotifyVO, language common.Language) (string, error) {
	template := notifyTemplate.Template
	// 取得 style 並更新 template
	style, err := ns.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_MAIL_TEMPLATE_STYLE)
	if err != nil {
		logger.Warn("Get style failed,", err)
		return "", err
	}
	template = strings.Replace(template, "{$STYLE}", style.Value, -1)
	// 取得 footer 並更新 template
	footer, err := ns.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_MAIL_TEMPLATE_FOOTER)
	if err != nil {
		logger.Warn("Get footer failed,", err)
		return "", err
	}
	template = strings.Replace(template, "{$FOOTER}", footer.Value, -1)

	deviceName := notifyVO.DeviceName
	if deviceName == "" {
		deviceName = "Device"
	}

	if notifyVO.Code == string(common.MSG_OPCODE_REAP_TRANSACTION_FAIL) {
		// 轉換 失敗原因，先確認 fail reason 是否為數字
		var checkFlag = true
		for _, r := range notifyVO.TxFailReason {
			if !unicode.IsDigit(r) {
				checkFlag = false
			}
		}
		if checkFlag {
			// 如果是數字，則轉換為對應的錯誤碼
			txFailReasonInt, err := strconv.Atoi(notifyVO.TxFailReason)
			if err != nil {
				logger.Warnf("failed to convert TxFailReason to int: %v", err)
			}
			utils.TranslateByCode(ctx, language, txFailReasonInt)
			notifyVO.TxFailReason = strconv.Itoa(int(common.CODE_SYSTEM_ERROR))
		}
	}

	// 替換模板中的佔位符
	placeholders := map[string]string{
		"{$CODE}":           notifyVO.Code,
		"{$EXPIRE}":         notifyVO.Expire,
		"{$CLIENT_IP}":      notifyVO.ClientIp,
		"{$DEVICE_NAME}":    notifyVO.DeviceName,
		"{$CARD_NO}":        notifyVO.CardNo,
		"{$CARD_TITLE}":     notifyVO.CardTitle,
		"{$AMOUNT}":         notifyVO.Amount,
		"{$CHANNEL}":        notifyVO.Channel,
		"{$CURRENCY}":       notifyVO.Currency,
		"{$MAINNET}":        notifyVO.Mainnet,
		"{$WALLET}":         notifyVO.Wallet,
		"{$OTP}":            notifyVO.OTPCode,
		"{$TX_FAIL_REASON}": notifyVO.TxFailReason,
		"{$DATE}":           notifyVO.Date,
	}

	for placeholder, value := range placeholders {
		if value != "" {
			template = strings.Replace(template, placeholder, value, -1)
		}
	}

	return template, nil
}

func (ns *NotifyService) sendApp(ctx context.Context, sendVO *entities.SendVO, notifyVO *entities.NotifyVO, language common.Language) {
	var userID = sendVO.UserID
	var err error
	if sendVO.UserID == 0 {
		userIDString := ctx.(*gin.Context).Request.Header.Get(common.HEADER_X_UID)
		if userIDString == "" {
			logger.Error("sendApp parse header error, no X-Uid")
		}

		userID, err = strconv.ParseUint(userIDString, 10, 64)
		if err != nil {
			logger.Warn("X-Uid parse failed,", err)
		}
	}
	// 取得用戶 token
	tokenKey := utils.GetFirebaseTokenKey(userID)
	firebaseTokens, err := utils.RDB.HGetAll(ctx, tokenKey).Result()

	if err != nil || len(firebaseTokens) == 0 {
		return
	}

	if err := InitFirebase(); err != nil {
		logger.Warn("firebase init fail", err)
		return
	}

	for deviceID, firebaseTokenVO := range firebaseTokens {
		// 依緩存token 取得 firbase token 與 language
		if firebaseTokenVO == "" {
			logger.Warn("device : %v, no firebase token", deviceID)
			continue
		}

		tokenVO := &entities.FirebaseTokenVO{}
		if err := json.Unmarshal([]byte(firebaseTokenVO), &tokenVO); err != nil {
			logger.Warn("unmarshal token error,", err)
			continue
		}
		if tokenVO == nil {
			logger.Warn("tokenVO is nil")
			continue
		}
		// 推送語言以 緩存記錄為主
		language = tokenVO.Language
		firebaseToken := tokenVO.FirebaseToken

		notifyTemplate, err := ns.notifyTemplateDao.GetNotifyTemplate(ctx, sendVO.Code, language, common.NOTIFY_TYPE_APP)
		if err != nil {
			logger.Warn("Get notify template failed,", err)
		}

		if notifyTemplate == nil {
			logger.Warn("no APP notify template", language, sendVO)
			return
		}

		subject, err := ns.convertSubject(ctx, notifyTemplate, notifyVO, language)
		if err != nil {
			logger.Warn("Convert subject failed,", err)
		}

		template, err := ns.convertTemplate(ctx, notifyTemplate, notifyVO, language)
		if err != nil {
			logger.Warn("Convert template failed,", err)
		}

		resultMap := make(map[string]string)
		var extra string
		if notifyTemplate.Extra != "" {
			extra = ns.convertTemplateExtra(ctx, notifyTemplate, notifyVO)
			// 初始化 map
			var jsonMap map[string]interface{}
			// 解析 JSON 字串
			err = json.Unmarshal([]byte(extra), &jsonMap)
			if err != nil {
				logger.Warn("Error parsing JSON:", err)
			}
			// 將 JSON 轉換為 map[string]string
			for key, value := range jsonMap {
				// 特殊處理 `params` 保留原結構
				if key == "params" {
					if params, ok := value.(map[string]interface{}); ok {
						// 嵌套結構轉換成 JSON 字串
						paramsBytes, _ := json.Marshal(params)
						resultMap[key] = string(paramsBytes)
					}
				} else {
					resultMap[key] = fmt.Sprintf("%v", value)
				}
			}
		}

		resultMap["userID"] = strconv.FormatUint(userID, 10)

		message := &messaging.Message{
			Data: resultMap,
			Notification: &messaging.Notification{
				Title: "eUSD Card " + subject,
				Body:  template,
			},
			Token: firebaseToken,
		}

		response, err := client.Send(ctx, message)
		if err != nil {
			logger.Warn("send firebase message fail,", err)
		}

		extraRecord, err := json.Marshal(resultMap)
		if err != nil {
			logger.Warn("convert extra to json failed,", err)
			extraRecord = []byte("failed to convert extra")
		}
		if _, err = ns.notifyRecordDao.Save(ctx, &notifyDao.NotifyRecord{
			UserID:          userID,
			TemplateID:      notifyTemplate.ID,
			TemplateCode:    notifyTemplate.Code,
			Language:        notifyTemplate.Language,
			Subject:         "eUSD Card " + subject,
			Content:         template,
			Extra:           string(extraRecord),
			NotifyType:      common.NOTIFY_TYPE_APP,
			Status:          common.NOTIFY_STATUS_SUCCESS,
			ResponseContent: response,
		}); err != nil {
			logger.Warn("Save notify record failed, ", err)
		}
		logger.Infof("send firebase message reslt: %v", response)
	}

}

func (ns *NotifyService) convertTemplateExtra(ctx context.Context, notifyTemplate *systemDao.NotifyTemplate, notifyVO *entities.NotifyVO) string {
	extra := notifyTemplate.Extra

	// 依序檢查 notifyVO 並取代 extra
	if notifyVO.OrderNo != "" {
		extra = strings.Replace(extra, "{$ORDER_NO}", notifyVO.OrderNo, -1)
	}

	if notifyVO.Threeds != "" {
		extra = strings.Replace(extra, "{$THREEDS}", notifyVO.Threeds, -1)
	}

	if notifyVO.KycLevel != "" {
		extra = strings.Replace(extra, "{$KYC_LEVEL}", notifyVO.KycLevel, -1)
	}

	return extra
}

func (ns *NotifyService) RetryCallback(ctx context.Context, callbackType common.CallbackType) error {

	locker := utils.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_CALLBACK, callbackType.String()),
		utils.Config.TopDown.PreviewExpireSeconds*time.Second,
		utils.Config.System.LockWaitMicroseconds*time.Microsecond,
	); err != nil {
		logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CALLBACK, callbackType.String()), err)
		return err
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CALLBACK, callbackType.String()), err)
		}
	}()

	tasks, err := ns.callbackTaskDao.ListByTypeStatusOrderByNotifiedAt(ctx, callbackType, common.CALLBACK_STATUS_PENDING_RETRY)
	if err != nil {
		logger.Warn("get failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	var cancels sync.Map
	defer func() {
		cancels.Range(func(key, value interface{}) bool {
			cencel := value.(context.CancelFunc)
			cencel()
			return true
		})
	}()

	reqID, _ := utils.GetMDCValue("reqId")
	count := 0
	waitChan := make(chan struct{})
	go func() { // waitgroup用
		utils.SetMDCValue("reqId", reqID)
		wg := &sync.WaitGroup{}
		for _, task := range tasks {

			if task.Attempt >= task.MaxAttempt {
				logger.Warnf("task already exceeded max attempt: [%d]", task.ID)
				continue
			}

			req := task.OriginalRequest
			if task.RequestEncryptType != 0 {
				req = task.EncryptedRequest
			}

			cctx, cancel := context.WithCancel(ctx)
			cancels.Store(task.ID, cancel)

			wg.Add(1)
			go func(cctx context.Context) { // 併發回調用
				utils.SetMDCValue("reqId", reqID)
				defer func() { count++ }()
				defer wg.Done()

				var oriRes []byte
				var encRes []byte
				var statusCode int
				var success bool
				var err error
				status := common.CALLBACK_STATUS_PENDING_RETRY
				switch task.ResponseEncryptType {
				case 0:
					oriRes, statusCode, success, err = utils.CallbackPlain(ctx, req, task.URL)
					encRes = []byte{}
				default:
					oriRes, encRes, statusCode, success, err = utils.CallbackEncrypted(ctx, req, task.URL, task.Key)
				}

				defer func() {

					var terr error
					derr := utils.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
						var tctx = context.WithValue(cctx, "db", tx)

						var errCode int
						var errMsg string
						if err != nil {
							errCode = common.CODE_SYSTEM_ERROR
							errMsg = err.Error()
						}
						if berr, ok := err.(*common.BusinessError); ok {
							errCode = berr.Code()
						}

						now := time.Now()

						var rowsAffected int64
						rowsAffected, terr = ns.callbackTaskDao.Update(tctx, &notifyDao.CallbackTaskQuery{
							CallbackTask: notifyDao.CallbackTask{
								ID: task.ID,
							},
							Attrs: notifyDao.CallbackTask{
								OriginalResponse:  string(oriRes),
								EncryptedResponse: string(encRes),
								Attempt:           task.Attempt + 1,
								HTTPCode:          statusCode,
								ErrorCode:         errCode,
								ErrorMessage:      errMsg,
								Status:            status,
								NotifiedAt:        utils.DBQueryTime(now),
							},
						})
						if terr != nil {
							logger.Warn("update failed, ", terr)
							terr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
							return terr
						}
						if rowsAffected == 0 {
							logger.Warnf("update failed: [%d]", task.ID)
							return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
						}

						_, terr = ns.callbackLogDao.Save(ctx, &notifyDao.CallbackLog{
							TaskID:              task.ID,
							OrderNO:             task.OrderNO,
							ThreedsID:           task.ThreedsID,
							UserID:              task.UserID,
							MerchantID:          task.MerchantID,
							URL:                 task.URL,
							Type:                task.Type,
							Scene:               task.Scene,
							MaxAttempt:          task.MaxAttempt,
							Attempt:             task.Attempt + 1,
							Criteria:            task.Criteria,
							TimeoutSeconds:      task.TimeoutSeconds,
							OriginalRequest:     task.OriginalRequest,
							EncryptedRequest:    task.EncryptedRequest,
							OriginalResponse:    string(oriRes),
							EncryptedResponse:   string(encRes),
							Key:                 task.Key,
							RequestEncryptType:  task.RequestEncryptType,
							ResponseEncryptType: task.ResponseEncryptType,
							HTTPCode:            statusCode,
							ErrorCode:           errCode,
							ErrorMessage:        errMsg,
							Status:              status,
							NotifiedAt:          utils.DBQueryTime(now),
						})
						if terr != nil {
							logger.Warn("save failed, ", terr)
							terr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
							return terr
						}

						return nil
					})

					if terr != nil {
						logger.Error("transaction failed,", terr)
						return
					}

					if derr != nil {
						logger.Error("transaction failed,", derr)
					}
				}()

				status = common.CALLBACK_STATUS_SUCCESS
				defer func() {
					if status == common.CALLBACK_STATUS_PENDING_RETRY {
						if task.Attempt+1 >= task.MaxAttempt {
							status = common.CALLBACK_STATUS_FAILED
						}
					}
				}()
				switch task.Criteria {
				case common.CALLBACK_CRITERIA_STATUS_CODE_200:
					if statusCode != 200 || !success {
						status = common.CALLBACK_STATUS_PENDING_RETRY
					}
				case common.CALLBACK_CRITERIA_CODE_OK:
					if !success {
						status = common.CALLBACK_STATUS_PENDING_RETRY
						return
					}
					res := &utils.Response[interface{}]{}
					jerr := json.Unmarshal(oriRes, res)
					if jerr != nil {
						logger.Warn("unmarshal failed, ", jerr)
					}
					if res.Code != common.CODE_OK {
						status = common.CALLBACK_STATUS_PENDING_RETRY
						return
					}
				case common.CALLBACK_CRITERIA_MSG_OK:
					if !success {
						status = common.CALLBACK_STATUS_PENDING_RETRY
						return
					}
					res := &utils.Response[interface{}]{}
					jerr := json.Unmarshal(oriRes, res)
					if jerr != nil {
						logger.Warn("unmarshal failed, ", jerr)
					}
					if res.Msg != "ok" {
						status = common.CALLBACK_STATUS_PENDING_RETRY
						return
					}
				}

			}(cctx)

			time.Sleep(utils.Config.Notify.CallbackIntervalMilliseconds * time.Millisecond)
		}
		wg.Wait()
		waitChan <- struct{}{}
	}()

	select {
	case <-waitChan:
		logger.Infof("callback type: [%d] done [%d]", callbackType, len(tasks))
	case <-time.After(utils.Config.Notify.CallbackTaskTimeoutSeconds * time.Second):
		logger.Warnf("callback type: [%d] timeout: [%d]/[%d]", callbackType, count, len(tasks))
	}
	return nil
}

func (ns *NotifyService) GetTemplateList(ctx context.Context, form *entities.GetTemplateListForm) []*systemDao.NotifyTemplate {
	result, err := ns.notifyTemplateDao.GetNotifyTemplateList(ctx, form)
	if err != nil {
		logger.Warn("Get notify template List failed,", err)
	}
	return result
}

func (ns *NotifyService) getPassageWay(p common.Passageway) common.Passageway {
	if p == common.PASSAGEWAY_SENDGRID || p == common.PASSAGEWAY_MAILGUN {
		return p
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	if rnd.Intn(2) == 0 {
		return common.PASSAGEWAY_MAILGUN
	}
	return common.PASSAGEWAY_SENDGRID
}

func (ns *NotifyService) PublishMail(ctx *gin.Context, passageway common.Passageway, limit int) error {
	records, err := ns.notifyRecordDao.ListAnnouncements(ctx, "", common.NOTIFY_TYPE_EMAIL, common.NOTIFY_STATUS_SENDING, limit)
	if err != nil {
		logger.Warnf("get notify record failed: %v", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	for _, record := range records {
		if record == nil {
			continue
		}
		if record.Recipient == "" {
			logger.Warnf("recipient format incorrect: %v %v", record.ID, record.Recipient)
		}
		err = nil
		var responseContent string
		var responseCode string
		passageway = ns.getPassageWay(passageway)
		if passageway == common.PASSAGEWAY_MAILGUN {
			resp, tErr := utils.SendMailGun(ctx, utils.SendRequest{
				Sender:    utils.Email{Name: utils.Config.System.BrandName, Address: utils.Config.System.MailGunSendMail},
				Recipient: utils.Email{Name: "", Address: record.Recipient},
				Subject:   record.Subject,
				Content:   record.Content,
			})
			if tErr != nil {
				logger.Warnf("send mailgun failed: [%v][%v][%v]", record.ID, resp, tErr)
				err = tErr
			}
			responseContent = resp
		} else {
			code, body, tErr := utils.SendMail(ctx, utils.SendRequest{
				Sender:    utils.Email{Name: utils.Config.System.BrandName, Address: utils.Config.System.SendMail},
				Recipient: utils.Email{Name: "", Address: record.Recipient},
				Subject:   record.Subject,
				Content:   record.Content,
			})
			if tErr != nil {
				logger.Warnf("send sendgrid failed: [%v][%v][%v][%v]", record.ID, code, body, tErr)
				err = tErr
			}
			responseCode = strconv.Itoa(code)
			responseContent = body
		}
		if err != nil {
			logger.Warnf("send mail failed: %v %v", record.ID, err)
			_, err2 := ns.notifyRecordDao.Update(ctx, &notifyDao.NotifyRecordQuery{
				NotifyRecord: notifyDao.NotifyRecord{ID: record.ID},
				Attrs: notifyDao.NotifyRecord{
					Passageway:      passageway,
					Status:          common.NOTIFY_STATUS_FAILED,
					ResponseCode:    responseCode,
					ResponseContent: responseContent,
				},
			})
			if err2 != nil {
				logger.Warnf("update notify record failed: %v", err2)
			}
		} else {
			_, err2 := ns.notifyRecordDao.Update(ctx, &notifyDao.NotifyRecordQuery{
				NotifyRecord: notifyDao.NotifyRecord{ID: record.ID},
				Attrs: notifyDao.NotifyRecord{
					Passageway:      passageway,
					Status:          common.NOTIFY_STATUS_SUCCESS,
					ResponseCode:    responseCode,
					ResponseContent: responseContent,
				},
			})
			if err2 != nil {
				logger.Warnf("update notify record failed: %v", err2)
			}
		}
		time.Sleep(time.Second)
	}

	return nil
}

func (ns *NotifyService) PublishApp(ctx *gin.Context, limit int) error {
	if err := InitFirebase(); err != nil {
		logger.Warnf("firebase init fail: %v", err)
		return err
	}
	records, err := ns.notifyRecordDao.ListAnnouncements(ctx, "", common.NOTIFY_TYPE_APP, common.NOTIFY_STATUS_SENDING, limit)
	if err != nil {
		logger.Warnf("get minotify record failed: %v", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	for _, record := range records {
		if record == nil {
			continue
		}
		resultMap := make(map[string]string)
		err = json.Unmarshal([]byte(record.Extra), &resultMap)
		if err != nil {
			logger.Warnf("get notify record failed: %v", err)
		}
		message := &messaging.Message{
			Data: resultMap,
			Notification: &messaging.Notification{
				Title: "eUSD Card " + "Notification",
				Body:  record.Content,
			},
			Token: record.Token,
		}
		response, err := client.Send(ctx, message)
		if err != nil {
			logger.Warnf("send firebase message fail: %v", err)
			_, err2 := ns.notifyRecordDao.Update(ctx, &notifyDao.NotifyRecordQuery{
				NotifyRecord: notifyDao.NotifyRecord{ID: record.ID},
				Attrs:        notifyDao.NotifyRecord{Status: common.NOTIFY_STATUS_FAILED, ResponseContent: err.Error()},
			})
			if err2 != nil {
				logger.Warnf("update notify record failed: %v", err2)
			}
		} else {
			_, err2 := ns.notifyRecordDao.Update(ctx, &notifyDao.NotifyRecordQuery{
				NotifyRecord: notifyDao.NotifyRecord{ID: record.ID},
				Attrs:        notifyDao.NotifyRecord{Status: common.NOTIFY_STATUS_SUCCESS, ResponseContent: response},
			})
			if err2 != nil {
				logger.Warnf("update notify record failed: %v", err2)
			}
		}
	}
	return nil
}
