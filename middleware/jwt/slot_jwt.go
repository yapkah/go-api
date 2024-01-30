package jwt

// SlotJWT is jwt middleware for slot token
// func SlotJWT() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var code int
// 		var data interface{}
// 		var userid int
// 		var usertype string
// 		var tokenid string
// 		code = e.SUCCESS

// 		// get header token
// 		token := c.GetHeader("Authorization")

// 		if token == "" { // empty token
// 			code = e.UNAUTHORIZED
// 		} else {
// 			// parse token
// 			claim, err := util.ParseSlotToken(token)

// 			if err != nil {
// 				if err.Error() == "public_key_missing" {
// 					code = e.PUBLIC_KEY_MISSING
// 				} else {
// 					code = e.UNAUTHORIZED
// 				}
// 			}

// 			if claim != nil {
// 				tokenid = claim.Id
// 				// set token claim
// 				c.Set("token_claim", claim)

// 				// check token id in db
// 				at, err := models.GetAllStatusSlotAccessTokenByID(claim.Id)
// 				if err != nil {
// 					code = e.UNAUTHORIZED
// 				}

// 				// check token mach code
// 				var tkData util.SlotTokenData
// 				if claim.Data != nil {
// 					data, _ := json.Marshal(claim.Data)
// 					json.Unmarshal(data, &tkData)
// 				}

// 				// if at != nil && at.Status == models.GetSlotAccesTokenReplaceStatus(){
// 				// 	code = e.YOUR_ARE_PLAYING_ANOTHER_GAME
// 				// }

// 				if at != nil && at.Status == "A" && at.SubID == claim.Subject && tkData.MachCode == at.MachCode {

// 					// set access token
// 					c.Set("access_token", at)

// 					// check user linked to token
// 					user, err := at.GetUser()
// 					if err != nil {
// 						code = e.UNAUTHORIZED
// 					}

// 					if user != nil {
// 						userid = user.GetUserID()
// 						usertype = user.GetUserType()
// 						// set access user
// 						c.Set("access_user", user)

// 					} else {
// 						code = e.UNAUTHORIZED
// 					}
// 				} else {
// 					code = e.UNAUTHORIZED
// 				}

// 			} else {
// 				code = e.UNAUTHORIZED
// 			}

// 		}

// 		l, ok := c.Get("api_log")
// 		if ok {
// 			log := l.(*models.ApiLog)
// 			log.UpdateUser(userid, usertype, tokenid)
// 		}

// 		if code != e.SUCCESS { // error return
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"status":     "error",
// 				"statusCode": code,
// 				"msg":        trans(c, e.GetMsg(code)),
// 				"data":       data,
// 			})
// 			c.Abort()
// 			return
// 		}

// 		c.Next()
// 	}
// }

// // CheckSlotDuplicate prevent duplicate play for slot token
// func CheckSlotDuplicate() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var code int
// 		var data interface{}

// 		code = e.UNAUTHORIZED

// 		tc, tok := c.Get("token_claim")
// 		u, cok := c.Get("access_user")

// 		if tok && cok {
// 			claim := tc.(*util.Claims)
// 			user := u.(*models.Members)

// 			if claim.Id == user.SlotTokenID {
// 				code = e.SUCCESS
// 			}
// 		}

// 		if code != e.SUCCESS { // error return
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"status":     "error",
// 				"statusCode": code,
// 				"msg":        e.GetMsg(code),
// 				"data":       data,
// 			})

// 			c.Abort()
// 			return
// 		}

// 		c.Next()
// 	}
// }
