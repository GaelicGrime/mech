package pandora

import (
   "bytes"
   "encoding/hex"
   "encoding/json"
   "github.com/89z/format"
   "golang.org/x/crypto/blowfish" //lint:ignore SA1019 reason
   "net/http"
   "net/url"
   "strconv"
   "strings"
)

const (
   origin = "http://android-tuner.pandora.com"
   partnerPassword = "AC7IBG09A3DTSYM4R41UJWL07VLN8JI7"
   syncTime = 0x7FFF_FFFF
)

var blowfishKey = []byte("6#26FRL$ZWD")

type Cipher struct {
   Bytes []byte
}

func Decode(s string) *Cipher {
   buf, err := hex.DecodeString(s)
   if err != nil {
      return nil
   }
   return &Cipher{buf}
}

func (c Cipher) Decrypt() *Cipher {
   sLen := len(c.Bytes)
   if sLen < blowfish.BlockSize {
      return nil
   }
   block, err := blowfish.NewCipher(blowfishKey)
   if err != nil {
      return nil
   }
   for low := 0; low < sLen; low += blowfish.BlockSize {
      block.Decrypt(c.Bytes[low:], c.Bytes[low:])
   }
   return &c
}

func (c Cipher) Encode() string {
   return hex.EncodeToString(c.Bytes)
}

func (c Cipher) Encrypt() *Cipher {
   block, err := blowfish.NewCipher(blowfishKey)
   if err != nil {
      return nil
   }
   for low := 0; low < len(c.Bytes); low += blowfish.BlockSize {
      block.Encrypt(c.Bytes[low:], c.Bytes[low:])
   }
   return &c
}

func (c Cipher) Pad() Cipher {
   bLen := blowfish.BlockSize - len(c.Bytes) % blowfish.BlockSize
   for high := byte(bLen); bLen >= 1; bLen-- {
      c.Bytes = append(c.Bytes, high)
   }
   return c
}

func (c Cipher) Unpad() *Cipher {
   bLen := len(c.Bytes)
   if bLen == 0 {
      return nil
   }
   high := bLen - int(c.Bytes[bLen-1])
   if high <= -1 {
      return nil
   }
   c.Bytes = c.Bytes[:high]
   return &c
}

type PartnerLogin struct {
   Result struct {
      PartnerAuthToken string
   }
}

func NewPartnerLogin() (*PartnerLogin, error) {
   body := map[string]string{
      "deviceModel": "android-generic",
      "password": partnerPassword,
      "username": "android",
      "version": "5",
   }
   buf := new(bytes.Buffer)
   err := json.NewEncoder(buf).Encode(body)
   if err != nil {
      return nil, err
   }
   req, err := http.NewRequest("POST", origin + "/services/json/", buf)
   if err != nil {
      return nil, err
   }
   req.URL.RawQuery = "method=auth.partnerLogin"
   format.Log.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusOK {
      return nil, response{res}
   }
   part := new(PartnerLogin)
   if err := json.NewDecoder(res.Body).Decode(part); err != nil {
      return nil, err
   }
   return part, nil
}

func (p PartnerLogin) UserLogin(username, password string) (*UserLogin, error) {
   rUser := userLoginRequest{
      LoginType: "user",
      PartnerAuthToken: p.Result.PartnerAuthToken,
      Password: password,
      SyncTime: syncTime,
      Username: username,
   }
   buf, err := json.Marshal(rUser)
   if err != nil {
      return nil, err
   }
   body := Cipher{buf}.Pad().Encrypt().Encode()
   req, err := http.NewRequest(
      "POST", origin + "/services/json/", strings.NewReader(body),
   )
   if err != nil {
      return nil, err
   }
   // auth_token can be empty, but must be included:
   req.URL.RawQuery = url.Values{
      "auth_token": {""},
      "method": {"auth.userLogin"},
      "partner_id": {"42"},
   }.Encode()
   format.Log.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   user := new(UserLogin)
   if err := json.NewDecoder(res.Body).Decode(user); err != nil {
      return nil, err
   }
   return user, nil
}

type PlaybackInfo struct {
   Stat string
   Result *struct {
      AudioUrlMap struct {
         HighQuality struct {
            AudioURL string
         }
      }
   }
}

type notFound struct {
   input string
}

func (n notFound) Error() string {
   return strconv.Quote(n.input) + " not found"
}

type playbackInfoRequest struct {
   // this can be empty, but must be included:
   DeviceCode string `json:"deviceCode"`
   IncludeAudioToken bool `json:"includeAudioToken"`
   PandoraID string `json:"pandoraId"`
   SyncTime int `json:"syncTime"`
   UserAuthToken string `json:"userAuthToken"`
}

type response struct {
   *http.Response
}

func (r response) Error() string {
   return r.Status
}

type userLoginRequest struct {
   LoginType string `json:"loginType"`
   PartnerAuthToken string `json:"partnerAuthToken"`
   Password string `json:"password"`
   SyncTime int `json:"syncTime"`
   Username string `json:"username"`
}

type valueExchangeRequest struct {
   OfferName string `json:"offerName"`
   SyncTime int `json:"syncTime"`
   UserAuthToken string `json:"userAuthToken"`
}