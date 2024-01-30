package e

const (
	SUCCESS                 = 200
	ERROR                   = 500
	NOT_MODIFIED            = 304
	INVALID_PARAMS          = 400
	UNAUTHORIZED            = 401
	NOT_FOUND               = 404
	UNSUPPORTED_MEDIA_TYPE  = 415
	UNPROCESSABLE_ENTITY    = 422
	INSUFFICIENT_BALANCE    = 423
	MEMBER_NOT_FOUND        = 424
	MEMBER_ACCOUNT_INACTIVE = 425

	// auth, otp
	ACCESS_TOKEN_NOT_FOUND                  = 10001
	INVALID_PASSWORD                        = 10002
	PUBLIC_KEY_MISSING                      = 10003
	PRIVATE_KEY_MISSING                     = 10004
	TOKEN_EXPIRED                           = 10005
	SIGNATURE_VERIFICATION_FAIL             = 10006
	INVAID_TOKEN_ISSUE_DATE                 = 10007
	INVAID_REFRESH_TOKEN                    = 10008
	OTP_EXPIRED                             = 10009
	OTP_EXCEED_MAX_ATTEMPTS                 = 10010
	INVALID_OTP                             = 10011
	INVALID_USERNAME_OR_PASSWORD            = 10012
	INVALID_USER                            = 10013
	PLEASE_REQUEST_OTP                      = 10014
	INVALID_OTP_TYPE                        = 10015
	INVALID_TOKEN_TYPE                      = 10016
	INVALID_OTP_SEND_TYPE                   = 10017
	INVALID_EMAIL_FORMAT                    = 10018
	EMAIL_IS_REQUIRED                       = 10019
	COUNTRY_CODE_IS_REQUIRED                = 10020
	MOBILE_NO_IS_REQUIRE                    = 10021
	EMPTY_USER_EMAIL_AND_MOBILE             = 10022
	INVALID_JWT_TOKEN_TYPE                  = 10023
	INVAID_SLOT_REFRESH_TOKEN               = 10024
	SLOT_ACCESS_TOKEN_NOT_FOUND             = 10025
	GENERATE_ACCESS_TOKEN_ID_ERROR          = 10026
	GENERATE_REFRESH_TOKEN_ID_ERROR         = 10027
	GENERATE_SLOT_ACCESS_TOKEN_ID_ERROR     = 10028
	GENERATE_SLOT_REFRESH_TOKEN_ID_ERROR    = 10029
	PASSWORD_VALIDATION_ERROR               = 10030
	ETAG_RESPONSE_DATA_NOT_FOUND            = 10031
	ETAG_INVALID_RESPONSE_DATA              = 10032
	PLEASE_CHECK_EMAIL_FOR_ACTIVATION_LINK  = 10033
	PLEASE_CEHCK_MOBILE_FOR_ACTIVATION_CODE = 10034
	YOUR_ACCOUNT_IS_LOGIN_AT_ANOTHER_DEVICE = 10035
	OTP_SETTING_NOT_FOUND                   = 10036
	OTP_MAX_REQUEST_SETTING_NOT_FOUND       = 10037
	OTP_MAX_REQUEST_SETTING_TIME_NOT_FOUND  = 10038
	OTP_EXCEED_MAX_REQUEST                  = 10039
	PLEASE_CHECK_EMAIL_FOR_OTP_CODE         = 10040
	PLEASE_CEHCK_MOBILE_FOR_OTP_CODE        = 10041
	INVALID_SECONDARY_PIN                   = 10042
	SECONDARY_PIN_VALIDATION_ERROR          = 10043
	RSA_PUBLIC_KEY_MISSING                  = 10044
	RSA_PRIVATE_KEY_MISSING                 = 10045
	INVALID_RSA_PRIVATE_KEY                 = 10046
	INVALID_RSA_PUBLIC_KEY                  = 10047
	DECREPT_PEM_PUBLIC_ERROR                = 10048
	DECREPT_PEM_PRIVATE_ERROR               = 10049
	RSA_INVALID_TEXT                        = 10050
	RSA_ENCRYPT_ERROR                       = 10051
	RSA_DECRYPT_ERROR                       = 10052

	// email / notification / sms
	TO_NAME_LENGHT_MUST_SAME_WITH_TO_EMAIL = 20001
	CC_NAME_LENGHT_MUST_SAME_WITH_CC_EMAIL = 20002
	SEND_MAIL_SUBJECT_CANNOT_EMPTY         = 20003
	SEND_MAIL_MESSAGE_CANNOT_EMPTY         = 20004
	SEND_EMAIL_FAIL                        = 20005
	SYS_EMAIL_NOT_FOUND                    = 20006
	INVALID_EMAIL_SETTING                  = 20007
	SEND_NOTIFICATION_FAILED               = 20008
	JPUSH_APP_KEY_MISSING                  = 20009
	JPUSH_MASTER_SECRET_MISSING            = 20010
	JPUSH_URL_MISSING                      = 20011
	NOTFICATION_LABEL_NOT_FOUND            = 20012
	INVALID_NOTIFICATION_STATUS            = 20013
	INVALID_NOTIFICATION_ID                = 20014
	INVALID_NOTIFICATION_MODULE            = 20015
	INVALID_NOTIFICATION_MODULE_ID         = 20016
	NOTIFICATION_LABEL_IS_REQUIRED         = 20017
	INVALID_NOTIFICATION_LABEL             = 20018
	INVALID_NOTIFICATION_MODULE_TYPE       = 20019
	EMPTY_NOTIFICATION_LABEL               = 20020
	NOTIFICATION_ALREADY_SEND              = 20021
	SMS_URL_SETTING_NOT_FOUND              = 20022
	SMS_CLIENT_ID_SETTING_NOT_FOUND        = 20023
	SMS_PRIVATE_KEY_SETTING_NOT_FOUND      = 20024
	SMS_USERNAME_SETTING_NOT_FOUND         = 20025
	SMS_MEMBER_MOBILE_NO_NOT_FOUND         = 20026
	SMS_MEMBER_MOBILE_COUNTRY_NOT_FOUND    = 20027
	SMS_LABEL_NOT_FOUND                    = 20028
	SMS_INVALID_MODULE                     = 20029
	SMS_LABEL_IS_REQUIRED                  = 20030
	SMS_LABEL_INVALID                      = 20031
	SMS_MODULE_ID_INVALID                  = 20032
	SMS_ID_INVALID                         = 20033
	SMS_LOG_ID_NOT_FOUND                   = 20034

	// user
	MEMBER_EMAIL_EXISTS                        = 30001
	MEMBER_ALREADY_ACTIVATED                   = 30002
	INVALID_MEMBER                             = 30003
	PASSWORD_MUST_MATCH_CONFIRM_PASS           = 30004
	INVALID_COUNTRY_CODE                       = 30005
	INVALID_MOBILE_NO                          = 30006
	MEMBER_MOBILE_EXISTS                       = 30007
	MEMBER_EMAIL_NOT_FOUND                     = 30008
	MEMBER_MOBILE_NOT_FOUND                    = 30009
	INVALID_USERNAME                           = 30010
	CANNOT_ENTER_SAME_EMAIL                    = 30011
	CANNOT_ENTER_SAME_MOBILE                   = 30012
	MEMBER_USERNAME_ALREADY_EXISTS             = 30013
	MEMBER_ALREADY_SET_USER_ID                 = 30014
	INVALID_SPONSOR                            = 30015
	INVALID_MACH_CODE                          = 30016
	INVALID_OLD_PASSWORD                       = 30017
	CHANGE_PASSWORD_OVER_ATTEMPT               = 30018
	RESET_PASSWORD_OVER_ATTEMPT                = 30019
	COM_NOT_FOUND                              = 30020
	SPONSOR_ID_NOT_DOWNLINE                    = 30021
	USERNAME_IS_REQUIRED                       = 30022
	MOBILE_IS_REQUIRED                         = 30023
	INVALID_REFERRAL_CODE                      = 30024
	GENERATE_REFERRAL_CODE_ERROR               = 30025
	GENERATE_MEMBER_SUB_ID_ERROR               = 30026
	PROFILE_PHOTO_IS_REQUIRED                  = 30027
	IMAGE_ONLY_SUPPORT_JPG_AND_PNG_ONLY        = 30028
	CREATE_FILE_ERROR                          = 30029
	COPY_FILE_ERROR                            = 30030
	INVALID_USER_ID                            = 30031
	USER_ID_MUST_CONTAIN_AT_LEAST_ONE_ALPHABET = 30032
	USER_ID_ONLY_ACCEPT_ALPHABET_AND_NUMBER    = 30033
	USER_ID_CANNOT_CONTAIN_SPACE               = 30034
	OS_VERSION_RECORD_NOT_FOUND                = 30035
	PROFILE_PIC_SIZE_LIMIT_IS                  = 30036
	CONNECTION_NOT_FOUND                       = 30037
	ACCOUNT_TERMINATED                         = 30038
	INVALID_MEMBER_STATUS                      = 30039

	// admin
	GENERATE_ADMIN_SUB_ID_ERROR   = 40001
	INVALID_DATE_FROM_TO_RANGE    = 40002
	INVALID_DATE_FORMAT           = 40003
	INVALID_SLOT_ROOM_STATUS      = 40004
	DUPLICATE_SLOT_ROOM           = 40005
	DUPLICATE_SLOT_DESC           = 40006
	INVALID_SLOT_ROOM             = 40007
	DUPLICATE_SECTION_DESC        = 40008
	INVALID_SLOT_ROW              = 40009
	INVALID_SLOT_COL              = 40010
	INVALID_WIN_RATE              = 40011
	INVALID_WIN_BIG_RATE          = 40012
	DUPLICATE_SLOT_MACH           = 40013
	INVALID_SECTION               = 40014
	INVALID_SLOT_MACH_TPYE        = 40015
	INVALID_SYMBOL_CODE           = 40016
	INVALID_FILTER_RANGE          = 40017
	APK_FILE_IS_REQUIRED          = 40018
	INVALID_VERSION_NUMBER_FORMAT = 40019
	VERSION_NUMBER_ALREADY_EXISTS = 40020
	INVALID_PAY_TOKEN_RANGE       = 40021
	INVALID_PAY_TOKEN_VAL_RANGE   = 40022
	INVALID_PAY_TOKEN_COMBINATION = 40023
	INVALID_PAY_TOKEN_VAL         = 40024
	INVALID_DATE_RANGE            = 40025

	// slot
	INVALID_TOTAL_BET_LINE                             = 50001
	INVALID_CONFIG_ROW_AND_COLUMN                      = 50002
	INVALID_CURRBETLINE_NOT_MATCH_PAYTOKEN_AND_PAYLINE = 50003
	PAYLINE_OUT_OF_RANGE                               = 50004
	INVALID_AT_STOP                                    = 50005
	INVALID_BET_TOKEN                                  = 50006
	FREELINE_OUT_OF_RANGE                              = 50007
	INVALID_STEP_RANGE                                 = 50008
	INVALID_STEP_POSITION                              = 50009
	REPEATED_PAY_LINE_RULES                            = 50010
	INVALID_ORDER_TYPE                                 = 50011
	INVALID_PAY_WIN_RATE_RANGE                         = 50012
	NOT_FOUND_BET_SIZE                                 = 50013

	// language / translation
	INVALID_LANGUAGE_CODE           = 60001
	TRANSLATION_NAME_ALREADY_EXISTS = 60002
	INVALID_TRANSLATION_ID          = 60003
	DATA_FILE_IS_REQUIRED           = 60004
	DATA_FILE_SHOULD_BE_CSV_FILE    = 60005
	DATA_FILE_READ_FILE_ERROR       = 60006

	// wallet
	NO_DATA_FOUND = 70001

	// base / media
	REQUEST_API_INVALID_METHOD   = 80001
	TEXT_PARSE_FAILED            = 80002
	TEXT_EXECUTE_FAILED          = 80003
	MEDIA_URL_NOT_FOUND          = 80004
	MEDIA_PROJECT_CODE_NOT_FOUND = 80005
	MEDIA_UPLOAD_FILE_ERROR      = 80006
	INVALID_FLOAT_STRING         = 80007

	//sicbo
	SICBO_TABLE_NOT_FOUND               = 90001
	TABLE_NO_REACH_MIN_BET              = 90002
	TABLE_EXCEED_MAX_BET                = 90003
	BET_INVALID_TOTAL                   = 90004
	CELL_BET_REPEATED                   = 90005
	CELL_NOT_UNDER_THIS_TABLE           = 90006
	ONLY_ALLOW_BET_ONE_SMALL_OR_BIG     = 90007
	SICBO_TABLE_INACTIVE                = 90008
	SICBO_TABLE_EXIST                   = 90009
	SICBO_SECTION_INVALID               = 90010
	SICBO_CELL_SETTING_NOT_VALID_LENGTH = 90011
	SICBO_CELL_SETTING_NOT_VALID        = 90012
	ONLY_ALLOW_BET_ONE_ODD_OR_EVEN      = 90013

	//roulette
	ROULETTE_WHEEL_EXIST                       = 100001
	ROULETTE_SECTION_INVALID                   = 100002
	ROULETTE_WHEEL_INVALID                     = 100003
	ROULETTE_WHEEL_INACTIVE                    = 100004
	ROULETTE_INVALID_TOTAL_COMBINATION_SETTING = 100005

	//baccarat
	BACCARAT_TABLE_INVALID   = 110001
	BACCARAT_TABLE_INACTIVE  = 110002
	BACCARAT_TABLE_EXIST     = 110003
	BACCARAT_SECTION_INVALID = 110004
	BACCARAT_CELL_INVALID    = 110005
)
