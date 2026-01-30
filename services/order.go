package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	coinsdoDao "api-server/dao/coinsdo"
	orderDao "api-server/dao/order"
	userDao "api-server/dao/user"
	walletDao "api-server/dao/wallet"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"golang.org/x/exp/slices"
)

type OrderService struct {
	transactionRecordDao *orderDao.TransactionRecordDao
	cardDao              *cardDao.CardDao
	categoryDao          *accountDao.CategoryDao
	walletDao            *walletDao.WalletDao
	userDao              *userDao.UserDao
	cryptoCurrencyDao    *coinsdoDao.CryptoCurrencyDao
}

func NewOrderService() *OrderService {
	return &OrderService{
		transactionRecordDao: orderDao.NewTransactionRecordDao(),
		cardDao:              cardDao.NewCardDao(),
		walletDao:            walletDao.NewWalletDao(),
		userDao:              userDao.NewUserDao(),
		cryptoCurrencyDao:    coinsdoDao.NewCryptoCurrencyDao(),
	}
}

type ListTransactionRecordsRequest struct {
	UserID          uint64
	AssetID         uint64
	FromCategoryID  uint64
	ToCategoryID    uint64
	FromUserID      uint64
	ToUserID        uint64
	Status          common.TransactionStatus
	ClearedAt       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	TransactionType common.TransactionType
	Mainnet         common.Mainnet
	Protocol        common.Protocol
	ReapChannel     common.ReapChannel
	FailReason      string
	ExchangeRate    decimal.Decimal
	ExchangeFee     decimal.Decimal
	DepositFee      decimal.Decimal
	TransferFee     decimal.Decimal
	CardFee         decimal.Decimal
	PhysicalCardFee decimal.Decimal
	DeliveryFee     decimal.Decimal
	ATMFee          decimal.Decimal
	FXFee           decimal.Decimal
	Page            int
	PageSize        int
}

type ListTransactionRecordsResponse struct {
	TransactionRecords []*orderDao.TransactionRecord
	Total              int64
}

func (os *OrderService) PageTransactionRecords(ctx context.Context, form *entities.PageTransactionRecordsForm, userIDs []uint64) (records []*orderDao.TransactionRecord, pageCurrent int, pageSize int, total int, err error) {

	var typeIn []common.TransactionRecordType
	for _, t := range form.Types {
		typeArray := common.TransactionRecordType(0).FromString(t)
		typeIn = append(typeIn, typeArray...)
	}

	if form.CardID != 0 {
		card, err := os.cardDao.GetByIDUserIDIn(ctx, form.CardID, userIDs)
		if err != nil {
			logger.Warn("get failed,", err)
			return nil, 0, 0, 0, err
		}
		if card == nil {
			return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_ORDER_USER_HAS_NO_SUCH_CARD)
		}
		records, pageCurrent, pageSize, total, err = os.transactionRecordDao.PageByUserIDInCardIDType(ctx, userIDs, form.CardID, typeIn, form.Current, form.PageSize)
		if err != nil {
			logger.Warn("get failed,", err)
			return nil, 0, 0, 0, err
		}

		return records, pageCurrent, pageSize, total, nil

	} else if form.Category != "" {
		currency := common.Currency(0).FromString(form.Category)
		if currency == 0 {
			return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_ORDER_NO_SUCH_CATEGORY)
		}
		form.CategoryID = uint64(currency)
	} else if form.CategoryID != 0 {
		// no-op
	} else {
		return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_MISSING_PARAMETER)
	}

	records, pageCurrent, pageSize, total, err = os.transactionRecordDao.PageByUserIDInCategoryIDType(ctx, userIDs, form.CategoryID, typeIn, form.Current, form.PageSize)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, 0, 0, 0, err
	}

	return records, pageCurrent, pageSize, total, nil
}

func (os *OrderService) GetTransactionRecord(ctx context.Context, form *entities.GetTransactionRecordForm, userIDs []uint64) (record *orderDao.TransactionRecord, err error) {

	side := common.TransactionSide(0).FromString(form.Side)

	record, err = os.transactionRecordDao.GetByTransactionNOSideCardID(ctx, form.OrderNO, side, form.CardID, userIDs)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, err
	}
	if !slices.Contains(userIDs, record.UserID) {
		return nil, nil
	}
	return record, nil
}

func (os *OrderService) Title(ctx context.Context, record *orderDao.TransactionRecord, cardMap map[uint64]*cardDao.Card) (string, error) {

	switch record.Type {
	case common.TRANSACTION_RECORD_TYPE_APPLY:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_APPLY), nil
	case common.TRANSACTION_RECORD_TYPE_TRANSFER:
		switch record.TransferChannel {
		case common.TRANSFER_CHANNEL_USER_ID:
			if record.Side == common.TRANSACTION_SIDE_FROM {
				return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TRANSFER_TO, strconv.FormatUint(record.ToUserID, 10)), nil
			}
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TRANSFER_FROM, strconv.FormatUint(record.FromUserID, 10)), nil
		case common.TRANSFER_CHANNEL_EMAIL:
			if record.Side == common.TRANSACTION_SIDE_FROM {
				return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TRANSFER_TO, record.ToEmail), nil
			}
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TRANSFER_FROM, record.FromEmail), nil
		case common.TRANSFER_CHANNEL_CARD_ID:
			if record.Side == common.TRANSACTION_SIDE_FROM {
				return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TRANSFER_TO, strconv.FormatUint(record.ToCardID, 10)), nil
			}
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TRANSFER_FROM, strconv.FormatUint(record.FromCardID, 10)), nil
		case common.TRANSFER_CHANNEL_ADDRESS:
			if record.Side == common.TRANSACTION_SIDE_FROM {
				return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TRANSFER_TO, record.ToAddress), nil
			}
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TRANSFER_FROM, record.FromAddress), nil
		}
		return "", nil
	case common.TRANSACTION_RECORD_TYPE_EXCHANGE:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_EXCHANGE, strings.ToUpper(record.FromCurrency.String()), strings.ToUpper(record.ToCurrency.String())), nil
	case common.TRANSACTION_RECORD_TYPE_WITHDRAW:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_WITHDRAW, record.ToAddress), nil
	case common.TRANSACTION_RECORD_TYPE_PAY:
		if record.ReapChannel == common.REAP_CHANNEL_VISA_DIRECT {
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_DEPOSIT, record.MerchantName), nil
		}
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_PAY, strings.ToUpper(record.MerchantName[:1])+record.MerchantName[1:]), nil
	case common.TRANSACTION_RECORD_TYPE_DEPOSIT:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_DEPOSIT, record.FromAddress), nil
	case common.TRANSACTION_RECORD_TYPE_TOP_UP:
		if record.Side == common.TRANSACTION_SIDE_TO {
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TOP_UP_TO, strings.ToUpper(record.FromCurrency.String())), nil
		}
		panNumber := "••••"
		if c, ok := cardMap[record.ToCardID]; ok && len(c.PANNumber) > 4 {
			panNumber += c.PANNumber[len(c.PANNumber)-4:]
		}
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TOP_UP_FROM, panNumber), nil
	case common.TRANSACTION_RECORD_TYPE_TOP_DOWN:
		if record.Side == common.TRANSACTION_SIDE_FROM {
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TOP_DOWN_FROM, strings.ToUpper(record.ToCurrency.String())), nil
		}
		panNumber := "••••"
		if c, ok := cardMap[record.FromCardID]; ok && len(c.PANNumber) > 4 {
			panNumber += c.PANNumber[len(c.PANNumber)-4:]
		}
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_TOP_DOWN_TO, panNumber), nil
	case common.TRANSACTION_RECORD_TYPE_REFUND:
		switch record.ReapChannel {
		case common.REAP_CHANNEL_VISA_DIRECT:
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_DEPOSIT, record.MerchantName), nil
		default:
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_REFUND, strings.ToUpper(record.MerchantName[:1])+record.MerchantName[1:]), nil
		}
	case common.TRANSACTION_RECORD_TYPE_VERIFY:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_VERIFY, strings.ToUpper(record.MerchantName[:1])+record.MerchantName[1:]), nil
	case common.TRANSACTION_RECORD_TYPE_CARD_TO_CARD:
		switch record.TransferChannel {
		case common.TRANSFER_CHANNEL_USER_ID:
			if record.Side == common.TRANSACTION_SIDE_FROM {
				return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_CARD_TO_CARD_TO, strconv.FormatUint(record.ToUserID, 10)), nil
			}
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_CARD_TO_CARD_FROM, strconv.FormatUint(record.FromUserID, 10)), nil
		case common.TRANSFER_CHANNEL_EMAIL:
			if record.Side == common.TRANSACTION_SIDE_FROM {
				return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_CARD_TO_CARD_TO, record.ToEmail), nil
			}
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_CARD_TO_CARD_FROM, record.FromEmail), nil
		case common.TRANSFER_CHANNEL_CARD_ID:
			if record.Side == common.TRANSACTION_SIDE_FROM {
				panNumber := "••••"
				if c, ok := cardMap[record.ToCardID]; ok && len(c.PANNumber) > 4 {
					panNumber += c.PANNumber[len(c.PANNumber)-4:]
				}
				return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_CARD_TO_CARD_TO, panNumber), nil
			}
			panNumber := "••••"
			if c, ok := cardMap[record.FromCardID]; ok && len(c.PANNumber) > 4 {
				panNumber += c.PANNumber[len(c.PANNumber)-4:]
			}
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_CARD_TO_CARD_FROM, panNumber), nil
		}
	case common.TRANSACTION_RECORD_TYPE_MANUAL:
		if record.Side == common.TRANSACTION_SIDE_FROM {
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_CARD_TO_CARD_TO, strconv.FormatUint(record.ToCardID, 10)), nil
		}
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_CARD_TO_CARD_FROM, strconv.FormatUint(record.FromCardID, 10)), nil
	case common.TRANSACTION_RECORD_TYPE_CHIPPAY_EXPRESS:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_CP_EXPRESS, strings.ToUpper(record.FromCurrency.String()), strings.ToUpper(record.ToCurrency.String())), nil
	case common.TRANSACTION_RECORD_TYPE_POINT_ACCURAL:
		switch record.PointSource {
		case common.POINT_SOURCE_TOP_UP:
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_POINT_ACCURAL), nil
		}
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_POINT_ACCURAL), nil
	case common.TRANSACTION_RECORD_TYPE_INTEREST:
		switch record.FinancialCode {
		case common.FINANCIAL_CODE_AUTO_YIELD:
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_INTEREST_AUTO_YIELD), nil
		}
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_INTEREST), nil
	case common.TRANSACTION_RECORD_TYPE_FINANCIAL_TRANSFER:
		if record.Side == common.TRANSACTION_SIDE_FROM {
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_FINANCIAL_TRANSFER_FROM, record.ToAlias), nil
		}
		switch record.FromType {
		case common.ASSET_TYPE_AUTO_YIELD:
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_FINANCIAL_TRANSFER_TO_AUTO_YIELD), nil
		}
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_FINANCIAL_TRANSFER_TO, record.FromAlias), nil
	case common.TRANSACTION_RECORD_TYPE_WHALE_PAY:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_PAY, record.WhaleDetail), nil
	case common.TRANSACTION_RECORD_TYPE_WHALE_REFUND:
		switch record.ReapChannel {
		case common.REAP_CHANNEL_VISA_DIRECT:
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_DEPOSIT, record.MerchantName), nil
		default:
			return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_REFUND, strings.ToUpper(record.WhaleDetail)), nil
		}
	case common.TRANSACTION_RECORD_TYPE_DISTRIBUTE_APPLY:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_DISTRIBUTE_APPLY), nil
	case common.TRANSACTION_RECORD_TYPE_PAYCRYPTO_PAY:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_PAY, strings.ToUpper(record.MerchantName)), nil
	case common.TRANSACTION_RECORD_TYPE_PAYCRYPTO_REFUND:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_REFUND, strings.ToUpper(record.MerchantName)), nil
	case common.TRANSACTION_RECORD_TYPE_PAYCRYPTO_VERIFY:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_VERIFY, strings.ToUpper(record.MerchantName)), nil

	case common.TRANSACTION_RECORD_TYPE_BINANCE_PAY:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_DEPOSIT, record.MerchantName), nil
	case common.TRANSACTION_RECORD_TYPE_SYSTEM_CHARGE:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_SYSTEM_CHARGE_DECLINE_FINE), nil
	case common.TRANSACTION_RECORD_TYPE_ETHERFI:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_PAY, strings.ToUpper(record.MerchantName)), nil
	case common.TRANSACTION_RECORD_TYPE_MERCHANT_ADJUST_BALANCE:
		return utils.Translate(ctx, common.TRANSLATE_MSG_ORDER_TITLE_MERCHANT_ADJUST_BALANCE), nil
	}

	return "", nil
}

func (os *OrderService) TransactionRecordToVO(
	ctx context.Context,
	record *orderDao.TransactionRecord,
	mainnetNames map[common.Mainnet]string,
	decimals map[common.Currency]int,
	cardMap map[uint64]*cardDao.Card,
) *entities.TransactionRecordVO {

	recordsVO := &entities.TransactionRecordVO{}
	var err error
	if err := copier.Copy(recordsVO, record); err != nil {
		logger.Warnf("copy failed [%+v],", record, err)
	}
	recordsVO.Title, err = os.Title(ctx, record, cardMap)
	if err != nil {
		logger.Warnf("get title failed [%+v],", record, err)
	}
	//  1-APL 2-TRF 3-EXG 7-TPU 8-TOD 11-CTC 歷史資料，display_rate 兼容
	if (record.Type == common.TRANSACTION_RECORD_TYPE_APPLY || record.Type == common.TRANSACTION_RECORD_TYPE_TRANSFER || record.Type == common.TRANSACTION_RECORD_TYPE_EXCHANGE ||
		record.Type == common.TRANSACTION_RECORD_TYPE_TOP_UP || record.Type == common.TRANSACTION_RECORD_TYPE_TOP_DOWN || record.Type == common.TRANSACTION_RECORD_TYPE_CARD_TO_CARD) && recordsVO.DisplayRate == nil {
		cryptos, err := os.cryptoCurrencyDao.ListByCurrencies(ctx, []common.Currency{
			record.FromCurrency,
			record.ToCurrency,
		})
		if err != nil {
			logger.Warnf("get cryptocurrency failed [%+v],", cryptos, err)
		}

		var toPlaces int
		for _, crypto := range cryptos {
			if crypto.CurrencyType == record.ToCurrency {
				toPlaces = crypto.DisplayDecimals
			}
		}

		if record.ToAmount.Decimal.IsZero() {
			recordsVO.DisplayRate = &decimal.Zero
		} else {
			var transsferFee = decimal.Zero
			if record.TransferFee != nil {
				transsferFee = *record.TransferFee
			}
			inverseRate := record.ToAmount.Decimal.Round(int32(toPlaces)).Div(record.FromAmount.Decimal.Sub(transsferFee)).Round(int32(utils.Config.System.RatePrecision))
			recordsVO.DisplayRate = &inverseRate
		}

	}
	recordsVO.Type = record.Type.String()
	if record.ReapChannel == common.REAP_CHANNEL_VISA_DIRECT {
		recordsVO.Type = common.TRANSACTION_RECORD_TYPE_DEPOSIT.String()
	}
	recordsVO.Income = record.Income.Decimal.RoundFloor(int32(decimals[record.IncomeCurrency])).StringFixed(int32(decimals[record.IncomeCurrency]))
	recordsVO.IncomeCurrency = record.IncomeCurrency.String()
	recordsVO.AgainstIncome = ""
	recordsVO.AgainstIncomeCurrency = ""
	if record.AgainstIncome.Valid {
		recordsVO.AgainstIncome = record.AgainstIncome.Decimal.String()
		recordsVO.AgainstIncomeCurrency = record.AgainstIncomeCurrency.String()
	}
	recordsVO.Side = record.Side.String()
	recordsVO.FromCategory = common.Currency(int(record.FromCategoryID)).String()
	recordsVO.ToCategory = common.Currency(int(record.ToCategoryID)).String()
	recordsVO.FromCurrency = record.FromCurrency.String()
	recordsVO.ToCurrency = record.ToCurrency.String()
	recordsVO.FromAmount = ""
	if record.FromAmount.Valid {
		recordsVO.FromAmount = record.FromAmount.Decimal.String()
	}
	recordsVO.ToAmount = ""
	if record.ToAmount.Valid {
		recordsVO.ToAmount = record.ToAmount.Decimal.String()
	}
	recordsVO.FromDiscount = ""
	if record.FromDiscount.Valid {
		recordsVO.FromDiscount = record.FromDiscount.Decimal.String()
	}
	recordsVO.ToBonus = ""
	if record.ToBonus.Valid {
		recordsVO.ToBonus = record.ToBonus.Decimal.String()
	}
	recordsVO.TransferToType = record.TransferToType.String()
	recordsVO.TransferToCurrency = record.TransferToCurrency.String()
	recordsVO.Mainnet = record.Mainnet.String()
	if record.Mainnet != 0 {
		recordsVO.MainnetName = mainnetNames[record.Mainnet]
	}
	recordsVO.Protocol = record.Protocol.String()
	recordsVO.TransferChannel = record.TransferChannel.String()
	recordsVO.ReapChannel = record.ReapChannel.String()
	recordsVO.ClearAmount = ""
	if record.ClearAmount.Valid {
		recordsVO.ClearAmount = record.ClearAmount.Decimal.StringFixed(int32(decimals[record.FromCurrency]))
	}
	recordsVO.ReversalAmount = ""
	if record.ReversalAmount.Valid {
		recordsVO.ReversalAmount = record.ReversalAmount.Decimal.StringFixed(int32(decimals[record.FromCurrency]))
	}
	recordsVO.ReversalTransactionAmount = ""
	if record.ReversalTransactionAmount.Valid {
		recordsVO.ReversalTransactionAmount = record.ReversalTransactionAmount.Decimal.StringFixed(int32(decimals[record.FromCurrency]))
	}
	recordsVO.ReversalExchangeRate = nil
	if record.ReversalExchangeRate.Valid {
		recordsVO.ReversalExchangeRate = &record.ReversalExchangeRate.Decimal
	}
	recordsVO.RefundAmount = ""
	if record.RefundAmount.Valid {
		recordsVO.RefundAmount = record.RefundAmount.Decimal.StringFixed(int32(decimals[record.FromCurrency]))
	}
	recordsVO.RefundTransactionAmount = ""
	if record.RefundTransactionAmount.Valid {
		recordsVO.RefundTransactionAmount = record.RefundTransactionAmount.Decimal.StringFixed(int32(decimals[record.FromCurrency]))
	}
	recordsVO.RefundExchangeRate = nil
	if record.RefundExchangeRate.Valid {
		recordsVO.RefundExchangeRate = &record.RefundExchangeRate.Decimal
	}

	recordsVO.ExchangeFee = utils.DecPtrToStr(record.ExchangeFee, decimals[record.ExchangeFeeCurrency])
	recordsVO.ExchangeFeeCurrency = record.ExchangeFeeCurrency.String()
	recordsVO.DepositFee = utils.DecPtrToStr(record.DepositFee, decimals[record.DepositFeeCurrency])
	recordsVO.DepositFeeCurrency = record.DepositFeeCurrency.String()
	recordsVO.TransferFee = utils.DecPtrToStr(record.TransferFee, decimals[record.TransferFeeCurrency])
	recordsVO.TransferFeeCurrency = record.TransferFeeCurrency.String()
	recordsVO.WithdrawFee = utils.DecPtrToStr(record.WithdrawFee, decimals[record.WithdrawFeeCurrency])
	recordsVO.WithdrawFeeCurrency = record.WithdrawFeeCurrency.String()
	recordsVO.CardFee = utils.DecPtrToStr(record.CardFee, decimals[record.CardFeeCurrency])
	recordsVO.CardFeeCurrency = record.CardFeeCurrency.String()
	recordsVO.PhysicalCardFee = utils.DecPtrToStr(record.PhysicalCardFee, decimals[record.PhysicalCardFeeCurrency])
	recordsVO.PhysicalCardFeeCurrency = record.PhysicalCardFeeCurrency.String()
	recordsVO.DeliveryFee = utils.DecPtrToStr(record.DeliveryFee, decimals[record.DeliveryFeeCurrency])
	recordsVO.DeliveryFeeCurrency = record.DeliveryFeeCurrency.String()
	recordsVO.TopUpFee = utils.DecPtrToStr(record.TopUpFee, decimals[record.TopUpFeeCurrency])
	recordsVO.TopUpFeeCurrency = record.TopUpFeeCurrency.String()
	recordsVO.TopDownFee = utils.DecPtrToStr(record.TopDownFee, decimals[record.TopDownFeeCurrency])
	recordsVO.TopDownFeeCurrency = record.TopDownFeeCurrency.String()
	recordsVO.CardToCardFee = utils.DecPtrToStr(record.CardToCardFee, decimals[record.CardToCardFeeCurrency])
	recordsVO.CardToCardFeeCurrency = record.CardToCardFeeCurrency.String()
	recordsVO.ATMFee = utils.DecPtrToStr(record.ATMFee, decimals[record.ATMFeeCurrency])
	recordsVO.ATMFeeCurrency = record.ATMFeeCurrency.String()
	recordsVO.FXFee = utils.DecPtrToStr(record.FXFee, decimals[record.FXFeeCurrency])
	recordsVO.FXFeeCurrency = record.FXFeeCurrency.String()
	recordsVO.ReapAuthorizationFee = utils.DecPtrToStr(record.ReapAuthorizationFee, decimals[record.ReapAuthorizationFeeCurrency])
	recordsVO.ReapAuthorizationFeeCurrency = record.ReapAuthorizationFeeCurrency.String()
	recordsVO.CPExpressFee = utils.DecPtrToStr(record.CPExpressFee, decimals[record.FXFeeCurrency])
	recordsVO.CPExpressFeeCurrency = record.CPExpressFeeCurrency.String()
	recordsVO.PromotionType = record.PromotionType.String()
	recordsVO.Status = record.Status.String()
	recordsVO.ReapDataLoss = record.ReapDataLoss.String()
	if record.ReversalAt != nil {
		recordsVO.ReversalAt = utils.TimeFromDB(*record.ReversalAt).UnixMilli()
	}
	if record.ClearedAt != nil {
		recordsVO.ClearedAt = utils.TimeFromDB(*record.ClearedAt).UnixMilli()
	}
	if record.RefundedAt != nil {
		recordsVO.RefundedAt = utils.TimeFromDB(*record.RefundedAt).UnixMilli()
	}
	if record.CoinsdoRepliedAt != nil {
		recordsVO.CoinsdoRepliedAt = utils.TimeFromDB(*record.CoinsdoRepliedAt).UnixMilli()
	}
	recordsVO.CreatedAt = utils.TimeFromDB(record.CreatedAt).UnixMilli()
	recordsVO.UpdatedAt = utils.TimeFromDB(record.UpdatedAt).UnixMilli()

	recordsVO.Remark = record.Remark

	return recordsVO
}

func (os *OrderService) getMaskMail(ctx context.Context, email string) string {
	parts := strings.Split(email, "@")
	// 本地部分的處理，只保留第一個字母
	localPart := parts[0]
	maskedLocal := ""
	if len(localPart) > 2 {
		maskedLocal = localPart[:2] + strings.Repeat("*", len(localPart)-2)
	} else {
		maskedLocal = strings.Repeat("*", len(localPart))
	}
	// 返回完整的結果
	return maskedLocal + "@" + parts[1]
}
