package instagram

import (
   "encoding/json"
   "encoding/xml"
   "net/http"
   "strconv"
   "strings"
   "time"
)

func appendImage(dst []string, src ImageVersion) []string {
   var (
      addr string
      max int
   )
   for _, can := range src.Candidates {
      if can.Height > max {
         addr = can.URL
         max = can.Height
      }
   }
   return append(dst, addr)
}

func appendManifest(dst []string, src string) ([]string, error) {
   var video dashManifest
   err := xml.Unmarshal([]byte(src), &video)
   if err != nil {
      return nil, err
   }
   for _, ada := range video.Period.AdaptationSet {
      var (
         addr string
         max int
      )
      for _, rep := range ada.Representation {
         if rep.Bandwidth > max {
            addr = rep.BaseURL
            max = rep.Bandwidth
         }
      }
      dst = append(dst, addr)
   }
   return dst, nil
}

func appendVideo(dst []string, src []VideoVersion) []string {
   var (
      addr string
      max int
   )
   for _, ver := range src {
      if ver.Type > max {
         addr = ver.URL
         max = ver.Type
      }
   }
   return append(dst, addr)
}

type EdgeText struct {
   Edges []struct {
      Node struct {
         Text string
      }
   }
}

type EdgeURL struct {
   Edges []struct {
      Node struct {
         Display_URL string
         Video_URL string
      }
   }
}

type GraphMedia struct {
   Edge_Media_To_Caption EdgeText
   Owner struct {
      Username string
   }
   Display_URL string
   Video_URL string
   Edge_Sidecar_To_Children EdgeURL
   Taken_At_Timestamp int64
   Edge_Media_To_Parent_Comment EdgeText
}

func NewGraphMedia(shortcode string) (*GraphMedia, error) {
   var buf strings.Builder
   buf.WriteString("https://www.instagram.com/p/")
   buf.WriteString(shortcode)
   buf.WriteByte('/')
   req, err := http.NewRequest("GET", buf.String(), nil)
   if err != nil {
      return nil, err
   }
   req.Header.Set("User-Agent", Android.String())
   req.URL.RawQuery = "__a=1"
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusOK {
      return nil, errorString(res.Status)
   }
   var post struct {
      GraphQL struct {
         Shortcode_Media GraphMedia
      }
   }
   if err := json.NewDecoder(res.Body).Decode(&post); err != nil {
      return nil, err
   }
   return &post.GraphQL.Shortcode_Media, nil
}

func (g GraphMedia) String() string {
   var buf []byte
   buf = append(buf, "Taken: "...)
   buf = append(buf, g.Time().String()...)
   buf = append(buf, "\nOwner: "...)
   buf = append(buf, g.Owner.Username...)
   for _, edge := range g.Edge_Media_To_Caption.Edges {
      buf = append(buf, "\nCaption: "...)
      buf = append(buf, edge.Node.Text...)
   }
   for _, edge := range g.Edge_Media_To_Parent_Comment.Edges {
      buf = append(buf, "\nComment: "...)
      buf = append(buf, edge.Node.Text...)
   }
   for _, addr := range g.URLs() {
      buf = append(buf, "\nURL: "...)
      buf = append(buf, addr...)
   }
   return string(buf)
}

func (g GraphMedia) Time() time.Time {
   return time.Unix(g.Taken_At_Timestamp, 0)
}

func (g GraphMedia) URLs() []string {
   src := make(map[string]bool)
   src[g.Display_URL] = true
   src[g.Video_URL] = true
   for _, edge := range g.Edge_Sidecar_To_Children.Edges {
      src[edge.Node.Display_URL] = true
      src[edge.Node.Video_URL] = true
   }
   var dst []string
   for key := range src {
      if key != "" {
         dst = append(dst, key)
      }
   }
   return dst
}

type ImageVersion struct {
   Candidates []struct {
      Width int
      Height int
      URL string
   }
}

type Item struct {
   Caption struct {
      Text string
   }
   User struct {
      Username string
   }
   Video_DASH_Manifest string
   Image_Versions2 ImageVersion
   Video_Versions []VideoVersion
   Carousel_Media []struct {
      Video_DASH_Manifest string
      Image_Versions2 ImageVersion
      Video_Versions []VideoVersion
   }
   Taken_At int64
}

func (i Item) Format() (string, error) {
   var buf []byte
   buf = append(buf, "Taken: "...)
   buf = append(buf, i.Time().String()...)
   buf = append(buf, "\nUser: "...)
   buf = append(buf, i.User.Username...)
   buf = append(buf, "\nCaption: "...)
   buf = append(buf, i.Caption.Text...)
   addrs, err := i.URLs()
   if err != nil {
      return "", err
   }
   for _, addr := range addrs {
      buf = append(buf, "\nURL: "...)
      buf = append(buf, addr...)
   }
   return string(buf), nil
}

func (i Item) Time() time.Time {
   return time.Unix(i.Taken_At, 0)
}

func (i Item) URLs() ([]string, error) {
   var (
      dst []string
      err error
   )
   if i.Video_DASH_Manifest != "" {
      dst, err = appendManifest(dst, i.Video_DASH_Manifest)
      if err != nil {
         return nil, err
      }
   } else if i.Video_Versions != nil {
      dst = appendVideo(dst, i.Video_Versions)
   } else if i.Image_Versions2.Candidates != nil {
      dst = appendImage(dst, i.Image_Versions2)
   }
   for _, med := range i.Carousel_Media {
      if med.Video_DASH_Manifest != "" {
         dst, err = appendManifest(dst, med.Video_DASH_Manifest)
         if err != nil {
            return nil, err
         }
      } else if med.Video_Versions != nil {
         dst = appendVideo(dst, med.Video_Versions)
      } else if med.Image_Versions2.Candidates != nil {
         dst = appendImage(dst, med.Image_Versions2)
      }
   }
   return dst, nil
}

func (l Login) Items(shortcode string) ([]Item, error) {
   var buf strings.Builder
   buf.WriteString("https://www.instagram.com/p/")
   buf.WriteString(shortcode)
   buf.WriteByte('/')
   req, err := http.NewRequest("GET", buf.String(), nil)
   if err != nil {
      return nil, err
   }
   req.Header = http.Header{
      "Authorization": {l.Authorization},
      "User-Agent": {Android.String()},
   }
   req.URL.RawQuery = "__a=1"
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusOK {
      return nil, errorString(res.Status)
   }
   var post struct {
      Items []Item
   }
   if err := json.NewDecoder(res.Body).Decode(&post); err != nil {
      return nil, err
   }
   return post.Items, nil
}

func (l Login) User(username string) (*User, error) {
   var buf strings.Builder
   buf.WriteString("https://www.instagram.com/")
   buf.WriteString(username)
   buf.WriteByte('/')
   req, err := http.NewRequest("GET", buf.String(), nil)
   if err != nil {
      return nil, err
   }
   req.Header.Set("User-Agent", Android.String())
   if l.Authorization != "" {
      req.Header.Set("Authorization", l.Authorization)
   }
   req.URL.RawQuery = "__a=1"
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   var profile struct {
      GraphQL struct {
         User User
      }
   }
   if err := json.NewDecoder(res.Body).Decode(&profile); err != nil {
      return nil, err
   }
   return &profile.GraphQL.User, nil
}

type User struct {
   Edge_Followed_By struct {
      Count int64
   }
   Edge_Follow struct {
      Count int64
   }
   Edge_Owner_To_Timeline_Media EdgeURL
}

func NewUser(username string) (*User, error) {
   return Login{}.User(username)
}

func (u User) String() string {
   buf := []byte("Followers: ")
   buf = strconv.AppendInt(buf, u.Edge_Followed_By.Count, 10)
   buf = append(buf, "\nFollowing: "...)
   buf = strconv.AppendInt(buf, u.Edge_Follow.Count, 10)
   return string(buf)
}

type VideoVersion struct {
   Type int
   Width int
   Height int
   URL string
}

type dashManifest struct {
   Period struct {
      AdaptationSet []struct { // one video one audio
         Representation []struct {
            Width int `xml:"width,attr"`
            Height int `xml:"height,attr"`
            Bandwidth int `xml:"bandwidth,attr"`
            BaseURL string
         }
      }
   }
}
