package main

import (
   "flag"
   "fmt"
   "github.com/89z/format"
   "github.com/89z/format/hls"
   "github.com/89z/mech/crackle"
   "io"
   "net/http"
   "os"
   "sort"
)

func download(stream hls.Stream, video *crackle.Video) error {
   fmt.Println("GET", stream.URI)
   res, err := http.Get(stream.URI.String())
   if err != nil {
      return err
   }
   seg, err := hls.NewScanner(res.Body).Segment(res.Request.URL)
   if err != nil {
      return err
   }
   if err := res.Body.Close(); err != nil {
      return err
   }
   file, err := os.Create(video.Base() + seg.Ext())
   if err != nil {
      return err
   }
   pro := format.ProgressChunks(file, len(seg.Info))
   for _, info := range seg.Info {
      res, err := http.Get(info.URI.String())
      if err != nil {
         return err
      }
      pro.AddChunk(res.ContentLength)
      if _, err := io.Copy(pro, res.Body); err != nil {
         return err
      }
      if err := res.Body.Close(); err != nil {
         return err
      }
   }
   return file.Close()
}

func doManifest(guid int64, bandwidth int, info bool) error {
   vod, err := crackle.NewAccessVOD(guid)
   if err != nil {
      return err
   }
   fmt.Println("GET", vod.ManifestPath)
   res, err := http.Get(vod.ManifestPath)
   if err != nil {
      return err
   }
   master, err := hls.NewScanner(res.Body).Master(res.Request.URL)
   if err != nil {
      return err
   }
   sort.Sort(hls.Bandwidth{master, bandwidth})
   for _, stream := range master.Stream {
      if info {
         fmt.Println(stream)
      } else {
         video, err := crackle.NewVideo(guid)
         if err != nil {
            return err
         }
         return download(stream, video)
      }
   }
   return res.Body.Close()
}

func main() {
   // b
   var guid int64
   flag.Int64Var(&guid, "b", 0, "GUID")
   // f
   var bandwidth int
   flag.IntVar(&bandwidth, "f", 5_581_000, "target bandwidth")
   // i
   var info bool
   flag.BoolVar(&info, "i", false, "information")
   // v
   var verbose bool
   flag.BoolVar(&verbose, "v", false, "verbose")
   flag.Parse()
   if verbose {
      crackle.LogLevel = 1
   }
   if guid >= 1 {
      err := doManifest(guid, bandwidth, info)
      if err != nil {
         panic(err)
      }
   } else {
      flag.Usage()
   }
}
