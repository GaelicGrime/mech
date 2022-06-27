package main

import (
   "flag"
   "github.com/89z/format/dash"
   "github.com/89z/mech/paramount"
   "net/url"
   "os"
   "path/filepath"
)

type downloader struct {
   *paramount.Preview
   client string
   info bool
   key []byte
   pem string
   url *url.URL
   media dash.Media
}

func main() {
   home, err := os.UserHomeDir()
   if err != nil {
      panic(err)
   }
   var down downloader
   // b
   var guid string
   flag.StringVar(&guid, "b", "", "GUID")
   // c
   down.client = filepath.Join(home, "mech/client_id.bin")
   flag.StringVar(&down.client, "c", down.client, "client ID")
   // d
   var is_DASH bool
   flag.BoolVar(&is_DASH, "d", false, "DASH download")
   // f
   // paramountplus.com/shows/video/x6XrF8A_tiSDRwc4Rt349KFKnCZ8QmtY
   var video int64
   flag.Int64Var(&video, "f", 1611000, "video bandwidth")
   // g
   // paramountplus.com/shows/video/x6XrF8A_tiSDRwc4Rt349KFKnCZ8QmtY
   var audio int64
   flag.Int64Var(&audio, "g", 999999, "audio bandwidth")
   // i
   flag.BoolVar(&down.info, "i", false, "information")
   // k
   down.pem = filepath.Join(home, "mech/private_key.pem")
   flag.StringVar(&down.pem, "k", down.pem, "private key")
   // v
   var verbose bool
   flag.BoolVar(&verbose, "v", false, "verbose")
   flag.Parse()
   if verbose {
      paramount.Client.Log_Level = 2
   }
   if guid != "" {
      var err error
      down.Preview, err = paramount.New_Media(guid).Preview()
      if err != nil {
         panic(err)
      }
      if is_DASH {
         err := down.DASH(video, audio)
         if err != nil {
            panic(err)
         }
      } else {
         err := down.HLS(video)
         if err != nil {
            panic(err)
         }
      }
   } else {
      flag.Usage()
   }
}
