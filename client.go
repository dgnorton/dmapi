package dmapi

import (
   "encoding/json"
   "fmt"
   "io/ioutil"
   "net/http"
   "strconv"
   "time"
)

type Entry struct {
   Id int `json:"id"`
   Url string `json:"url"`
   At string `json:"at"`
   Message string `json:"message"`
   Comments []Comment `json:"comments"`
   Likes []Like `json:"likes"`
   Geo Geo `json:"geo"`
   Location Location `json:"location"`
   User User `json:"user"`
   Workout Workout `json:"workout"`
}

func (e *Entry) Time() (time.Time, error) {
   return time.Parse(time.RFC3339, e.At)
}

type Comment struct {
   Body string `json:"body"`
   CreatedAt string `json:"created_at"`
   User User `json:"user"`
}

type Like struct {
}

type Location struct {
   Name string `json:"name"`
}

type User struct {
   Name string `json:"username"`
   DisplayName string `json:"display_name"`
   PhotoUrl string `json:"photo_url"`
   Url string `json:"usr"`
}

type Workout struct {
   Type string `json:"activity_type"`
   Distance Distance `json:"distance"`
   Felt string `json:"felt"`
   Dur int `json:"duration"`
   Title string `json:"title"`
}

func (w Workout) Pace() (time.Duration, error) {
   if w.Distance.Value <= 0.0 {
      return time.Duration(0.0), nil
   }
   secPerUnit := w.Duration().Seconds() / w.Distance.Value
   return time.ParseDuration(strconv.FormatFloat(secPerUnit, 'f', 6, 64) + "s")
}

func (w Workout) PaceStr() (string, error) {
   d, err := w.Pace()
   if err != nil {
      return "", err
   }
   return DurationStr(d), nil
}

func (w Workout) Duration() time.Duration {
   dur, _ := time.ParseDuration(strconv.Itoa(w.Dur) + "s")
   return dur
}

func DurationStr(d time.Duration) string {
   totSec := int(d.Seconds())
   h := totSec / 3600
   m := (totSec - (h * 3600)) / 60
   s := totSec - (h * 3600) - (m * 60)

   if h > 0 {
      return fmt.Sprintf("%d:%02d:%02d", h, m, s)
   }

   return fmt.Sprintf("%d:%02d", m, s)
}

func (w Workout) DurationStrColons() (string, error) {
   d, err := time.ParseDuration(strconv.Itoa(w.Dur) + "s")
   if err != nil {
      return "", err
   }
   return DurationStr(d), nil
}

type Distance struct {
   Value float64 `json:"value"`
   Units string `json:"units"`
}

type Geo struct {
   Type string `json:"type"`
   // longitude, latitude ... in that order
   Coordinates []string `json:"coordinates"`
}

type Entries struct {
   Entries []Entry `json:"entries"`
}

func (e Entries) TotalDistance() float64 {
   var d float64 = 0.0
   for _, entry := range e.Entries {
      // need to fix to deal with different units
      d += entry.Workout.Distance.Value
   }
   return d
}

func (e Entries) AvgPace() (time.Duration, error) {
   var avg int64 = 0
   var cnt int64 = 0

   for _, entry := range e.Entries {
      p, err := entry.Workout.Pace()
      if entry.Workout.Distance.Units == "kilometers" {
         p = mpkToMpm(p)
      }
      if err != nil {
         return time.Duration(0), err
      }

      avg += p.Nanoseconds()
      if p.Nanoseconds() > 0 {
         cnt++
      }
   }

   if cnt > 0 {
      avg /= cnt
   }

   return time.Duration(avg), nil
}

func (e Entries) AvgPaceStr() string {
   d, err := e.AvgPace()
   if err != nil {
      return ""
   }
   return DurationStr(d)
}

// Converts minutes/km to minutes/mi
func mpkToMpm(d time.Duration) time.Duration {
   return time.Duration(float64(d) * 1.609)
}

func EntriesByPage(user string, page int) (*Entries, error) {
   req := fmt.Sprintf("http://api.dailymile.com/people/%s/entries.json?page=%d", user, page)
   resp, err := http.Get(req)
   if err != nil {
      return nil, err
   }
   defer resp.Body.Close()
   body, err := ioutil.ReadAll(resp.Body)
   if err != nil {
      return nil, err
   }

   var entries Entries
   err = json.Unmarshal(body, &entries)

   return &entries, err
}

func EntriesSince(user string, unixTime int64) (*Entries, error) {
   req := fmt.Sprintf("http://api.dailymile.com/people/%s/entries.json?since=%d", user, unixTime)
   resp, err := http.Get(req)
   if err != nil {
      return nil, err
   }
   defer resp.Body.Close()
   body, err := ioutil.ReadAll(resp.Body)
   if err != nil {
      return nil, err
   }

   var entries Entries

   if len(body) > 0 {
      err = json.Unmarshal(body, &entries)
   }

   return &entries, err
}
