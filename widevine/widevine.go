package widevine

import (
   "crypto/rsa"
   "crypto/x509"
   "encoding/base64"
   "encoding/hex"
   "encoding/pem"
   "github.com/89z/format/http"
   "github.com/89z/format/protobuf"
   "strings"
)

func New_Module(private_key, client_id, key_id []byte) (*Module, error) {
   block, _ := pem.Decode(private_key)
   var (
      err error
      mod Module
   )
   mod.private_key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
   if err != nil {
      return nil, err
   }
   mod.license_request = protobuf.Message{
      1: protobuf.Bytes(client_id),
      2: protobuf.Message{ // ContentId
         1: protobuf.Message{ // CencId
            1: protobuf.Message{ // Pssh
               2: protobuf.Bytes(key_id),
            },
         },
      },
   }.Marshal()
   return &mod, nil
}

type no_operation struct{}

func (no_operation) Read(buf []byte) (int, error) {
   return len(buf), nil
}

func unpad(buf []byte) []byte {
   if len(buf) >= 1 {
      pad := buf[len(buf)-1]
      if len(buf) >= int(pad) {
         buf = buf[:len(buf)-int(pad)]
      }
   }
   return buf
}

type Content struct {
   Key []byte
   Type uint64
}

func (c Contents) Content() *Content {
   for _, con := range c {
      if con.Type == 2 {
         return &con
      }
   }
   return nil
}

func Key_ID(raw string) ([]byte, error) {
   raw = strings.ReplaceAll(raw, "-", "")
   return hex.DecodeString(raw)
}

func PSSH_Key_ID(raw string) ([]byte, error) {
   _, after, ok := strings.Cut(raw, "data:text/plain;base64,")
   if ok {
      raw = after
   }
   pssh, err := base64.StdEncoding.DecodeString(raw)
   if err != nil {
      return nil, err
   }
   cenc_header, err := protobuf.Unmarshal(pssh[32:])
   if err != nil {
      return nil, err
   }
   return cenc_header.Get_Bytes(2)
}

type Contents []Content

type Module struct {
   license_request []byte
   private_key *rsa.PrivateKey
}

var Client = http.Default_Client
