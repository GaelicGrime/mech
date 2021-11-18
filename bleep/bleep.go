package bleep

import (
   "encoding/json"
   "fmt"
   "github.com/89z/mech"
   "github.com/89z/parse/html"
   "io"
   "net/http"
   "net/url"
   "strings"
   "time"
)

const Origin = "https://bleep.com"

type Date string

// can be either one of these:
//  2001-05-01 00:00:00.0
//  Tue May 01 00:00:00 UTC 2001
func (d Date) Parse() (time.Time, error) {
   value := string(d)
   date, err := time.Parse(time.UnixDate, value)
   if err != nil {
      return time.Parse("2006-01-02 15:04:05.9", value)
   }
   return date, nil
}

type Meta struct {
   Image string `json:"og:image"`
   Release_Date Date `json:"music:release_date"`
}

func NewMeta(releaseID int) (*Meta, error) {
   req, err := http.NewRequest(
      "GET", fmt.Sprint(Origin, "/release/", releaseID), nil,
   )
   if err != nil {
      return nil, err
   }
   if mech.Verbose {
      fmt.Println(req.Method, req.URL)
   }
   // this redirects, so we cannot use RoundTrip
   res, err := new(http.Client).Do(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   met := new(Meta)
   if err := html.NewMap(res.Body).Struct(met); err != nil {
      return nil, err
   }
   return met, nil
}

type Track struct {
   Artist string
   Disc int
   Number int
   ReleaseID int
   Title string
}

// 8728-1-1
func Parse(track string) (*Track, error) {
   var t Track
   _, err := fmt.Sscanf(track, "%v-%v-%v", &t.ReleaseID, &t.Disc, &t.Number)
   if err != nil {
      return nil, err
   }
   return &t, nil
}

func Release(releaseID int) ([]Track, error) {
   val := url.Values{
      "id": {fmt.Sprint(releaseID)}, "type": {"ReleaseProduct"},
   }
   req, err := http.NewRequest(
      "POST", Origin + "/player/addToPlaylist", strings.NewReader(val.Encode()),
   )
   if err != nil {
      return nil, err
   }
   req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
   res, err := mech.RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   var rel []Track
   if err := json.NewDecoder(res.Body).Decode(&rel); err != nil {
      return nil, err
   }
   return rel, nil
}

func (t Track) Resolve() (string, error) {
   req, err := http.NewRequest(
      "GET", Origin + "/player/resolve/" + t.String(), nil,
   )
   if err != nil {
      return "", err
   }
   res, err := mech.RoundTrip(req)
   if err != nil {
      return "", err
   }
   defer res.Body.Close()
   dst, err := io.ReadAll(res.Body)
   if err != nil {
      return "", err
   }
   return string(dst), nil
}

func (t Track) String() string {
   return fmt.Sprint(t.ReleaseID, "-", t.Disc, "-", t.Number)
}