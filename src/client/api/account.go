package api

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/skycoin/skycoin-exchange/src/client/account"
	"github.com/skycoin/skycoin-exchange/src/pp"
	"github.com/skycoin/skycoin-exchange/src/sknet"
	"github.com/skycoin/skycoin/src/cipher"
)

// CreateAccount handle the request of creating account.
func CreateAccount(se Servicer) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// generate account pubkey/privkey pair, pubkey is the account id.
		errRlt := &pp.EmptyRes{}
		for {
			a := account.New()
			r := pp.CreateAccountReq{
				Pubkey: pp.PtrString(a.Pubkey),
			}

			req, err := makeEncryptReq(&r, se.GetServKey().Hex(), a.Seckey)
			if err != nil {
				logger.Error(err.Error())
				errRlt = pp.MakeErrResWithCode(pp.ErrCode_WrongRequest)
				break
			}
			rsp, err := sknet.Get(se.GetServAddr(), "/auth/create/account", req)
			if err != nil {
				logger.Error(err.Error())
				errRlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}
			res, err := decodeRsp(rsp.Body, se.GetServKey().Hex(), a.Seckey, &pp.CreateAccountRes{})
			if err != nil {
				logger.Error(err.Error())
				errRlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}

			acntRes := res.(*pp.CreateAccountRes)
			if !acntRes.GetResult().GetSuccess() {
				sendJSON(w, res)
			} else {
				ret := struct {
					Result    pp.Result `json:"result"`
					Pubkey    string    `json:"pubkey"`
					CreatedAt int64     `json:"created_at"`
				}{
					Result:    *acntRes.Result,
					Pubkey:    a.Pubkey,
					CreatedAt: acntRes.GetCreatedAt(),
				}
				account.Set(a)
				sendJSON(w, &ret)
			}
			return
		}
		sendJSON(w, errRlt)
	}
}

// GetAccount get account that matchs the condition in url param.
// mode: GET
// url: /api/v1/account?active=[:active]
// params:
// 		active: optional condition, must be 1, if not exist, then retun all accounts.
func GetAccount(se Servicer) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// get active
		active := r.FormValue("active")
		switch active {
		case "1":
			res := struct {
				Result   *pp.Result `json:"result"`
				Accounts string     `json:"account,omitempty"`
			}{}
			a := account.GetActive()
			if a.Pubkey == "" && a.Seckey == "" {
				res.Result = pp.MakeResultWithCode(pp.ErrCode_NotExits)
			} else {
				res.Result = pp.MakeResultWithCode(pp.ErrCode_Success)
				res.Accounts = a.Pubkey
			}
			sendJSON(w, &res)
		case "":
			res := struct {
				Result   *pp.Result `json:"result"`
				Accounts []string   `json:"accounts,omitempty"`
			}{}
			accounts := account.GetAll()
			res.Result = pp.MakeResultWithCode(pp.ErrCode_Success)
			res.Accounts = func(accounts []account.Account) []string {
				as := make([]string, len(accounts))
				for i, a := range accounts {
					as[i] = a.Pubkey
				}
				return as
			}(accounts)
			sendJSON(w, &res)
		default:
			sendJSON(w, pp.MakeErrResWithCode(pp.ErrCode_WrongRequest))
		}
	}
}

// ActiveAccount active the specific account.
// mode: PUT
// url: /api/v1/account/state?pubkey=[:pubkey]
func ActiveAccount(se Servicer) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var rlt *pp.EmptyRes
		for {
			// get pubkey
			pk := r.FormValue("pubkey")
			if pk == "" {
				logger.Error("pubkey is empty")
				rlt = pp.MakeErrRes(errors.New("pubkey is empty"))
				break
			}

			// validate the pubkey
			if _, err := cipher.PubKeyFromHex(pk); err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrRes(errors.New("invalid pubkey"))
				break
			}

			// active the account
			if err := account.SetActive(pk); err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrRes(err)
				break
			}
			res := struct {
				Result *pp.Result
			}{
				pp.MakeResultWithCode(pp.ErrCode_Success),
			}
			sendJSON(w, &res)
			return
		}
		sendJSON(w, rlt)
	}
}
