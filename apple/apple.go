package apple

import (
   "encoding/json"
   "github.com/89z/mech"
   "github.com/89z/parse/html"
   "net/http"
   "strconv"
   "strings"
)

const podcast = "\uf8ff.v1.catalog.us.podcast-episodes."

type Attributes struct {
   ArtistName string
   AssetURL string
   DurationInMilliseconds int
   Name string
   ReleaseDateTime string
}

type Audio struct {
   D []struct {
      Attributes Attributes
   }
}

func NewAudio(addr string) (*Audio, error) {
   req, err := http.NewRequest("GET", addr, nil)
   if err != nil {
      return nil, err
   }
   res, err := mech.RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   for _, node := range html.Parse(res.Body, "script") {
      if node.Attr["id"] == "shoebox-media-api-cache-amp-podcasts" {
         var raw map[string]json.RawMessage
         if err := json.Unmarshal(node.Data, &raw); err != nil {
            return nil, err
         }
         for key, val := range raw {
            if strings.HasPrefix(key, podcast) {
               unq, err := strconv.Unquote(string(val))
               if err != nil {
                  return nil, err
               }
               aud := new(Audio)
               if err := json.Unmarshal([]byte(unq), aud); err != nil {
                  return nil, err
               }
               return aud, nil
            }
         }
      }
   }
   return nil, mech.NotFound{podcast}
}