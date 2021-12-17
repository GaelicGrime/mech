package pandora

import (
   "encoding/hex"
   "encoding/json"
   "github.com/89z/mech"
   "golang.org/x/crypto/blowfish"
   "net/http"
   "net/url"
   "strings"
   "time"
)

type userLogin struct {
   Result struct {
      UserID string
      UserAuthToken string
   }
}

func (p partnerLogin) userLogin(username, password string) (*userLogin, error) {
   rUser := userLoginRequest{
      LoginType: "user",
      PartnerAuthToken: p.Result.PartnerAuthToken,
      Password: password,
      SyncTime: time.Now().Unix(),
      Username: username,
   }
   dec, err := json.Marshal(rUser)
   if err != nil {
      return nil, err
   }
   enc, err := encrypt(dec)
   if err != nil {
      return nil, err
   }
   req, err := http.NewRequest(
      "POST", origin + "/services/json/",
      strings.NewReader(hex.EncodeToString(enc)),
   )
   if err != nil {
      return nil, err
   }
   req.Header.Set("User-Agent", "Pandora/2110.1 Android/7.0 generic_x86")
   req.URL.RawQuery =
      "partner_id=42&method=auth.userLogin&auth_token=" +
      url.QueryEscape(p.Result.PartnerAuthToken)
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   user := new(userLogin)
   if err := json.NewDecoder(res.Body).Decode(user); err != nil {
      return nil, err
   }
   return user, nil
}

type partnerLogin struct {
   Result struct {
      PartnerAuthToken string
   }
}

const origin = "http://android-tuner.pandora.com"

var (
   LogLevel mech.LogLevel
   key = []byte("6#26FRL$ZWD")
)

func decrypt(src []byte) ([]byte, error) {
   dst := make([]byte, len(src))
   blow, err := blowfish.NewCipher(key)
   if err != nil {
      return nil, err
   }
   for low := 0; low < len(src); low += blowfish.BlockSize {
      blow.Decrypt(dst[low:], src[low:])
   }
   pad := dst[len(dst)-1]
   return dst[:len(dst)-int(pad)], nil
}

func encrypt(src []byte) ([]byte, error) {
   for len(src) % blowfish.BlockSize >= 1 {
      src = append(src, 0)
   }
   dst := make([]byte, len(src))
   blow, err := blowfish.NewCipher(key)
   if err != nil {
      return nil, err
   }
   for low := 0; low < len(src); low += blow.BlockSize() {
      blow.Encrypt(dst[low:], src[low:])
   }
   return dst, nil
}

func newPartnerLogin() (*partnerLogin, error) {
   body := `{"username":"android","password":"AC7IBG09A3DTSYM4R41UJWL07VLN8JI7"}`
   req, err := http.NewRequest(
      "POST", origin + "/services/json/", strings.NewReader(body),
   )
   if err != nil {
      return nil, err
   }
   req.URL.RawQuery = "method=auth.partnerLogin"
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   login := new(partnerLogin)
   if err := json.NewDecoder(res.Body).Decode(login); err != nil {
      return nil, err
   }
   return login, nil
}

type userLoginRequest struct {
   LoginType string `json:"loginType"`
   PartnerAuthToken string `json:"partnerAuthToken"`
   Password string `json:"password"`
   SyncTime int64 `json:"syncTime"`
   Username string `json:"username"`
}
