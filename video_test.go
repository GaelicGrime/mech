package youtube

import (
   "net/http"
   "testing"
)

var tests = []struct{id, desc string}{
   {
      "XeojXq6ySs4",
      "Provided to YouTube by Epitaph\n\nSnowflake · Kate Bush\n\n" +
      "50 Words For Snow\n\n" +
      "℗ Noble & Brite Ltd. trading as Fish People, under exclusive license to Anti Inc.\n\n" +
      "Released on: 2011-11-22\n\nMusic  Publisher: Noble and Brite Ltd.\n" +
      "Composer  Lyricist: Kate Bush\n\nAuto-generated by YouTube.",
   }, {
      "ClYg-0-z_ds",
      "",
   },
}

func TestDesc(t *testing.T) {
   for _, test := range tests {
      v, err := NewVideo(test.id)
      if err != nil {
         t.Error(err)
      }
      if v.Description() != test.desc {
         t.Errorf("%+v\n", v)
      }
   }
}

func TestStream(t *testing.T) {
   v, e := NewVideo("GiNR187EMd4")
   if e != nil {
      t.Error(e)
   }
   s, e := v.GetStream(251)
   if e != nil {
      t.Error(e)
   }
   return
   r, e := http.Head(s)
   if e != nil {
      t.Error(e)
   }
   if r.StatusCode != 200 {
      t.Error(r.StatusCode)
   }
}
