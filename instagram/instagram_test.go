package instagram

import (
   "fmt"
   "os"
   "testing"
)

const (
   mech = `C:\Users\Steven\AppData\Local\mech\instagram.json`
   sidecar = "CT-cnxGhvvO"
   video = "CUWBw4TM6Np"
)

func TestData(t *testing.T) {
   f, err := os.Open(mech)
   if err != nil {
      t.Fatal(err)
   }
   defer f.Close()
   var auth Login
   if err := auth.Decode(f); err != nil {
      t.Fatal(err)
   }
   Verbose = true
   m, err := NewQuery(sidecar).Data(&auth)
   if err != nil {
      t.Fatal(err)
   }
   fmt.Printf("%+v\n", m)
}

func TestItem(t *testing.T) {
   f, err := os.Open(mech)
   if err != nil {
      t.Fatal(err)
   }
   defer f.Close()
   var auth Login
   if err := auth.Decode(f); err != nil {
      t.Fatal(err)
   }
   Verbose = true
   i, err := auth.Item(video)
   if err != nil {
      t.Fatal(err)
   }
   fmt.Printf("%+v\n", i)
}

func TestRead(t *testing.T) {
   f, err := os.Open(mech)
   if err != nil {
      t.Fatal(err)
   }
   defer f.Close()
   var auth Login
   if err := auth.Decode(f); err != nil {
      t.Fatal(err)
   }
   m, err := GraphQL(sidecar, &auth)
   if err != nil {
      t.Fatal(err)
   }
   fmt.Printf("%+v\n", m)
}

func TestWrite(t *testing.T) {
   pass, ok := os.LookupEnv("PASS")
   if ! ok {
      t.Fatal("PASS")
   }
   Verbose = true
   l, err := NewLogin("srpen6", pass)
   if err != nil {
      t.Fatal(err)
   }
   if err := l.Encode(os.Stdout); err != nil {
      t.Fatal(err)
   }
}
