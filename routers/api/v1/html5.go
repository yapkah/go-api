package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/controllers/member"
	"github.com/yapkah/go-api/controllers/member/product"
	"github.com/yapkah/go-api/controllers/member/report"
	"github.com/yapkah/go-api/controllers/member/wallet"
	"github.com/yapkah/go-api/middleware/api"
	"github.com/yapkah/go-api/middleware/jwt"
)

// Html5 func
func Html5(route *gin.RouterGroup) {

	route.POST("/translation/add", member.AddTranslation)         // add frontend translation
	route.GET("/translation/update", member.UpdateAppTranslation) // update frontend translation
	route.GET("/country/list", member.CountryList)                // get country list
	route.GET("/language/list", member.LanguageList)              // get langugage list
	route.GET("/locales/:lang/:namespace", member.GetTranslation) // get frontend translation
	route.GET("/locales/:lang", member.GetAppTranslation)         // get app frontend translation
	route.GET("/app-version/list", member.GetAppVersListv1)       // get app version
	route.POST("/app-version/check", member.CheckAppVersv1)       // check app version
	route.POST("/app-version/process", member.ProcessAppVersv1)   // process app version

	route.GET("/referral/validate", member.ValidateReferralCode) // validate encrypted referral code

	group := route.Group("/member")
	{
		// admin login member - this api will kickout current login member
		group.POST("/admin/login/access/generate", member.AdminGenerateMemberAccess) // generate login access
		group.POST("/admin/login/gateway", member.AdminLoginGateway)                 // bypass admin login member

		// admin login member - this api will not kickout current login member
		group.POST("/admin/login-gateway/tmp-password", member.AdminLoginGatewayTmpPassword) // generate login access

		// address
		group.POST("/address/login/access/generate", member.AddressGenerateMemberAccess) // generate login access

		// pre-login
		group.GET("/pre-login/document/list", member.GetPreloginDocumentList) // get prelogin document list

		group.Use(jwt.JWT())
		group.Use(api.CheckForMaintenanceMode())
		group.POST("/login", member.Login)
		group.POST("/register", member.Registerv2)            // create user login
		group.GET("/username/random", member.GetRandUsername) // get random username
		// group.GET("/profile/:name", member.GetMemberByUsername)              // get by username
		// group.POST("/register/email", member.RegisterByEmail)                // register with email
		// group.POST("/translation/list", member.TranslationList, etag.Etag()) // get translation list

		// otp
		group.POST("/otp/request", member.RequestOTP) // request activation otp
		// group.POST("/otp/validate", member.ValidateOTP) // validate otp for register/reset password

		// group.GET("/forget-password/request", member.RequestForgetPassword) // request forget password

		// reser password/transaction pin
		reset := group.Group("/reset")
		{
			reset.POST("/password", member.ResetPassword)                         // reset security setting
			reset.POST("/password/key", member.ResetPasswordWithHashedPrivateKey) // reset security setting
			reset.POST("/secondary-pin", member.ResetSecondaryPin)                // reset security setting
		}

		// mnemonic
		// group.GET("/mnemonic/request", member.RequestMnemonicv1) // request new mnemonic

		// token
		// group.POST("/slot-token/refresh", member.RefreshSlotToken)

		// version
		// group.POST("/version/check", member.VersionChecking)

		auth := group.Group("/")
		auth.Use(api.UpdateMemberLatestLangCode())
		// auth.Use(jwt.CheckScopeOr([]string{"ACCESS"})) // check if it is an access token
		{
			// auth.POST("/account/create", member.CreateAccountv2)       // create account
			auth.POST("/logout", member.Logout)                        // put it here to prevent special case happend like duplicate call and return failed
			auth.GET("/setting/status", member.GetMemberSettingStatus) // return user account related status
			// auth.POST("/mobile/update", member.UpdateMobile)           // update mobile
			// auth.POST("/mnemonic/bind", member.BindMnemonicv1)  // bind mnemonic
			// auth.POST("/pk/info/update", member.UpdatePKInfov1) // update mnemonic / private key

			// 	// with otp specific scope
			// 	auth.POST("/activate", jwt.CheckScopeOr([]string{"MEM-REG", "STAT-I"}), member.ActivateMember)               // member activate with token
			// 	auth.POST("/password/reset", jwt.CheckScopeOr([]string{"MEM-RP"}), member.ResetPassword)                     // member reset password with token
			// 	auth.POST("/email/update", jwt.CheckScopeOr([]string{"MEM-UE"}), member.UpdateEmail)                         // update email
			// 	auth.POST("/otp/email/update", jwt.CheckScopeOr([]string{"MEM-VM", "MEM-VE"}), member.UpdateEmailOTP)        // verify new email
			// 	auth.POST("/otp/mobile/update", jwt.CheckScopeOr([]string{"MEM-VM", "MEM-VE"}), member.UpdateMobileOTP)      // verify new mobile no
			// 	auth.POST("/password/secondary/forget", jwt.CheckScopeOr([]string{"MEM-FTP"}), member.ForgetTradingPassword) // member activate with token

			memberGroup := auth.Group("/")
			memberGroup.Use(api.CheckActiveMember())
			memberGroup.Use(api.MemberAccessPermission())
			// memberGroup.Use(jwt.CheckScopeAnd([]string{"MEM", "STAT-A"})) // check token scope
			// memberGroup.Use(jwt.CheckDuplicateLogin())                    // prevent duplication login
			{
				// 		memberGroup.POST("/userid/add", member.AddUserID) // user id only can add once (cannot be changed)

				// 		// otp
				// 		memberGroup.POST("/otp/request", member.RequestOTP)
				// 		memberGroup.POST("/otp/email/verify", member.VerifyEmailOTP)
				// 		memberGroup.POST("/otp/mobile/verify", member.VerifyMobileOTP)
				// 		memberGroup.POST("/otp/trading-password/forget", member.ForgetTradingPasswordOTP)

				//Dashboard
				memberGroup.GET("/dashboard", member.GetDashboard)

				// Profile
				memberGroup.GET("/profile", member.GetProfile)
				memberGroup.POST("/profile/update", member.UpdateProfile)
				memberGroup.POST("/password/update", member.UpdatePassword)
				memberGroup.POST("/secondary-pin/update", member.UpdateSecondaryPassword)
				memberGroup.POST("/secondary-pin/check", member.CheckSecondaryPasswordv1)
				memberGroup.POST("/mobile/bind", member.BindMobile)
				memberGroup.POST("/email/bind", member.BindEmail)

				// Account
				// accountGroup := memberGroup.Group("/account")
				// {
				// 	accountGroup.GET("/list", member.GetMemberAccountListv1)
				// 	accountGroup.POST("/switch", member.SwitchCurrentActiveMemberAccountv1)
				// 	accountGroup.POST("/deactivate", member.DeactivateMemberAccountv1) // so far no use. [wai kit said it is handle by front-end]
				// 	accountGroup.POST("/unbind", member.UnbindMemberAccountv1)
				// 	accountGroup.POST("/import/check", member.CheckImportMemberAccountv1)
				// 	accountGroup.POST("/tag", member.TagMemberAccountv1)
				// 	accountGroup.GET("/transfer-exchange/batch/assets", member.GetMemberAccountTransferExchangeBatchAssetsv1)
				// 	accountGroup.POST("/transfer-exchange/batch/setup", member.GetMemberAccountTransferExchangeBatchSetupv1)
				// }

				// KYC
				memberGroup.POST("/kyc/create", member.CreateMemberKYCv1)
				memberGroup.GET("/kyc/details", member.GetMemberKYCv1)

				// tree
				memberGroup.GET("/tree/list", member.GetMemberTreev1)
				// memberGroup.GET("/tree/list2", member.GetMemberTreev2)

				// 		// Secondary Password
				// 		memberGroup.POST("/password/secondary", member.CreateTradingPassword)      //  create trading password
				// 		memberGroup.POST("/password/secondary/reset", member.ResetTradingPassword) //  reset trading password

				//  get profile
				// 		memberGroup.POST("/profile/photo", member.UploadProfilePhoto) // upload profile photo
				// 		//memberGroup.GET("/profile", member.GetMember)                 // for mobile

				// 		memberGroup.POST("/memberships", member.GetMemberships) // get memberships

				// 		// language
				// 		memberGroup.POST("/language/update", member.UpdateLanguage)

				// member device info
				memberGroup.POST("/device-info/update", member.UpdateDeviceInfo)

				// push notification
				// Account
				memberPNGroup := memberGroup.Group("/push-notification")
				{
					memberPNGroup.GET("/list", member.GetMemberPushNotificationListv1)              // get member push notification list
					memberPNGroup.POST("/msg/remove", member.ProcessRemoveMemberPushNotificationv1) // remove member push notification list
					// 		memberPNGroup.POST("/notification/status", member.UpdateNotificationStatus)
					// 		memberPNGroup.POST("/notification/receive", member.UpdateMemberNotiStatus)
					// 		memberPNGroup.POST("/notification/receive/status", member.GetMemberNotiStatus)
				}
				memberGroup.GET("/blockchain-explorer/list", member.GetMemberBlockChainExplorerListv1)

				memberGroup.POST("/file/upload", member.UploadMemberFile)
				// Wallet
				// memberGroup.GET("/address", member.GetCryptoAddress) //eddie request fmt

				// statementGroup := memberGroup.Group("/statement")
				// {
				// statementGroup.GET("/list", member.GetMemberStatementListv1)
				// statementGroup.GET("/listV2", member.GetMemberStatementListv2) //ui-v2
				// }

				walletGroup := memberGroup.Group("/wallet")
				{
					walletGroup.POST("/withdraw", member.PostWithdraw)
					walletGroup.GET("/withdraw-setting", member.GetWithdrawSetting)
					walletGroup.GET("/transfer-setting", member.GetTransferSetting)
					walletGroup.GET("/transfer-exchange-setting", member.GetTransferExchangeSetting)
					// walletGroup.POST("/transfer", member.PostTransfer)
					// walletGroup.POST("/transfer-exchange", member.PostTransferExchangeV2)
					// walletGroup.POST("/transfer-exchange/batch", member.PostTransferExchangeBatch)
					walletGroup.GET("/balance", member.GetMemberBalanceListv1)
					// walletGroup.POST("/exchange", member.PostWalletExchange)
					// walletGroup.GET("/exchange/setting", member.GetExchangeSetting)
					// walletGroup.GET("/sign-key/setting", member.GetWalletSigningKey)
					// walletGroup.POST("/setting", member.GetWalletSetting)
					// walletGroup.GET("/transaction", member.GetWalletTransactionv1)
					walletGroup.GET("/withdraw/statement", member.GetWithdrawStatement)
					// walletGroup.GET("/transfer/statement", member.GetTransferStatement)
					// walletGroup.GET("/withdraw/transactionFee", member.GetWithdrawTransactionFee)

					walletGroup.GET("/transaction", member.GetWalletStatement)
					// walletGroup.GET("/summary-detail", member.GetWalletSummaryDetails)
					// walletGroup.GET("/withdraw/detail", member.GetWithdrawDetail)              // for new ui wallet statement
					// walletGroup.GET("/transfer/detail", member.GetTransferDetail)              // for new ui wallet statement
					// walletGroup.POST("/convert", member.PostConvert)

					// walletGroup.GET("/exchange/setting", product.GetExchangeSetting)
					// walletGroup.POST("/exchange", product.PostExchange)

					// get pending adjustment
					walletGroup.GET("/pending-transfer-out", member.GetPendingTransferOut)
					walletGroup.POST("/adjust-out", member.AdjustOut)
				}

				// Sales
				salesGroup := memberGroup.Group("/sales")
				{
					salesGroup.GET("/list", member.GetMemberSalesListv1)
					salesGroup.GET("/list/summary", member.GetMemberSalesListSummary)
					salesGroup.GET("/topup/list", member.GetMemberSalesTopupListv1)
					// salesGroup.POST("/package/purchase", member.PurchasePackageV2)
					// salesGroup.POST("/package/topup", member.TopupPackage)
					// salesGroup.POST("/package/test", member.DailyTokenDistribution)

					// salesGroup.POST("/ballot", member.PostBallot)
					// salesGroup.GET("/ballot/setting", member.GetBallotSetting)
					// salesGroup.GET("/ballot/list", member.GetMemberBallotList)
					// salesGroup.POST("/ballot/winner", member.PostBallotWinner)

					miningNodeGroup := salesGroup.Group("/mining/node")
					{
						miningNodeGroup.GET("/list", member.GetMemberMiningNodeListV1)
						miningNodeGroup.GET("/list/card/update", member.GetMemberMiningNodeListUpdateV1)
						miningNodeGroup.POST("/topup", product.TopupMiningNode)
						miningNodeGroup.GET("/topup/list", member.GetMemberMiningNodeTopupListV1)
					}
				}

				// Product
				memberGroup.GET("/products", product.GetProductsv1)

				nftGroup := memberGroup.Group("/nft")
				{
					nftGroup.GET("/series", product.GetNftSeries)
				}

				// Contract
				contractGroup := memberGroup.Group("/contracts")
				{
					contractGroup.POST("/purchase", product.PurchaseContract) // purchase contract
					// contractGroup.POST("/topup", product.TopupContract)       // topup contract
				}

				// staking
				// stakingGroup := memberGroup.Group("/staking")
				// {
				// stakingGroup.POST("/purchase", product.PostStaking) // post staking
				// stakingGroup.POST("/refund", product.PostUnstake)   // post unstake
				// }

				// Mining
				miningGroup := memberGroup.Group("/mining-action")
				{
					miningGroup.GET("/list", product.GetMemberMiningActionListv1)
					miningGroup.GET("/contract", product.GetMemberContractMiningActionDetailsv1)
					miningContractGroup := miningGroup.Group("/contract")
					{
						miningContractGroup.GET("/history", product.GetContractMiningActionHistoryList)
						miningContractGroup.GET("/ranking", product.GetContractMiningActionRankingList)
					}

					miningGroup.GET("/staking", product.GetMemberStakingMiningActionDetailsv1)
					miningGroup.GET("/mining", product.GetMemberMiningMiningActionDetailsv1)
					miningGroup.GET("/mining/list", product.GetMemberMiningMiningActionListv1)
					miningGroup.GET("/pool", product.GetMemberPoolMiningActionDetailsv1)
				}

				memberGroup.GET("/exchange-price/list", wallet.GetWSExchangePriceRateList)

				// Deposit
				cryptoGroup := memberGroup.Group("/crypto")
				{
					cryptoGroup.GET("/get", member.GetCryptoDetail)
					cryptoGroup.POST("/purchase", member.AddCryptoPurchase)
					cryptoGroup.POST("/cancel", member.CancelCryptoPurchase)
				}

				// Reward
				rewardGroup := memberGroup.Group("/reward")
				{
					// rewardGroup.POST("/statement", member.GetRewardList)
					// rewardGroup.POST("/setting", member.GetRewardSetting)
					// rewardGroup.GET("/detail", member.GetRewardStatements)
					rewardGroup.GET("/summary", member.GetRewardSummary)
					rewardGroup.GET("/statement", member.GetRewardStatement)
					rewardGroup.GET("/history", member.GetRewardHistory)
					// rewardGroup.GET("/detail", member.GetRewardDetail)

				}

				// announcement
				memberGroup.GET("/announcement/list", member.GetMemberAnnouncementListv1)

				// event
				memberGroup.GET("/event/list", member.GetMemberEventListv1)

				// pool
				memberGroup.GET("/pool/list", member.GetMemberPoolListv1)

				// Trading
				// tradingGroup := memberGroup.Group("/trading")
				// {
				// 	tradingGroup.GET("/market/list", trading.GetMemberTradingMarketListv1)
				// 	tradingGroup.GET("/available-market-price/buy/list", trading.GetMemberAvailableTradingSellListv1)
				// 	tradingGroup.GET("/available-market-price/sell/list", trading.GetMemberAvailableTradingBuyListv1)
				// 	tradingGroup.GET("/selection/list", trading.GetMemberTradingSelectionListv1)
				// 	tradingGroup.GET("/setup", trading.GetMemberTradingSetupv2)
				// 	tradingGroup.GET("/buy/list", trading.GetMemberTradingBuyListv1)
				// 	tradingGroup.GET("/sell/list", trading.GetMemberTradingSellListv1)
				// 	// tradingGroup.POST("/buy", trading.MemberBuyTradingv1)
				// 	tradingGroup.POST("/buy/request", trading.MemberBuyTradingRequestv2)
				// 	// tradingGroup.POST("/sell", trading.MemberSellTradingv1)
				// 	tradingGroup.POST("/sell/request", trading.MemberSellTradingRequestv2)
				// 	tradingGroup.POST("/request/cancel", trading.MemberCancelTradingRequestv1)
				// 	tradingGroup.GET("/open-order/list", trading.GetMemberTradingOpenOrderTransListv1)
				// 	tradingGroup.GET("/order-history/list", trading.GetMemberTradingOrderHistoryTransListv1)
				// 	tradingGroup.GET("/order-history/details", trading.GetMemberTradingOrderHistoryTransDetailsv1)
				// 	tradingGroup.GET("/trade-history/list", trading.GetMemberTradingHistoryTransListv1)
				// 	// tradingGroup.GET("/statement/:reward_type_code", member.GetRewardStatementV2) // for new ui reward statement
				// 	tradingGroup.GET("/trading-view/setup", trading.GetMemberTradingViewSetupv1)
				// }

				// Language
				memberGroup.POST("/language/update", member.UpdateMemberDeviceLanguagev1)

				// //Support Ticket
				// supportGroup := memberGroup.Group("/support")
				// {
				// 	supportGroup.POST("/ticket", member.PostSupportTicket)
				// 	supportGroup.GET("/ticket/details", member.GetMemberSupportTicketList)
				// 	supportGroup.GET("/ticket/history", member.GetMemberSupportTicketHistoryList)
				// 	supportGroup.POST("/ticket-reply", member.PostSupportTicketReply)
				// 	supportGroup.POST("/ticket-close", member.PostSupportTicketClose)
				// }

				// route.GET("/general/faq", member.GetFaq) //get faq

				// route.GET("/general/nftImg", member.GetNftImg)

				// Event
				// eventGroup := memberGroup.Group("/event")
				// {
				// 	// Ranking Event
				// 	rankingGroup := eventGroup.Group("/ranking")
				// 	{
				// 		rankingGroup.GET("/sponsor", event.GetEventSponsorRankingList)
				// 		rankingGroup.GET("/sponsor/setting", event.GetEventSponsorRankingSetting)
				// 	}

				// 	// Auction Event
				// 	auctionGroup := eventGroup.Group("/auction")
				// 	{
				// 		auctionGroup.GET("/lucky_number", event.GetAuctionLuckyNumberList)
				// 		auctionGroup.GET("/lucky_number/history", event.GetAuctionLuckyNumberHistoryList)
				// 	}

				// }

				// Report
				reportGroup := memberGroup.Group("/report")
				{
					reportGroup.GET("/setup", report.GetReportSetup)
					reportGroup.GET("/list", report.GetReportList)
				}
			}
		}
	}
}
