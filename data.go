package dmapi

import (
   "encoding/json"
   "errors"
   "io/ioutil"
   "regexp"
   "time"
)

func SaveEntries(file string, entries *Entries) error {
   bytes, err := json.Marshal(entries)
   if err != nil {
      return err
   }
   err = ioutil.WriteFile(file, bytes, 0600)
   return err
}

func LoadEntries(file string) (*Entries, error) {
   bytes, err := ioutil.ReadFile(file)
   if err != nil {
      return nil, err
   }
   var entries Entries
   err = json.Unmarshal(bytes, &entries)
   return &entries, err
}

func (e *Entries) Remove(id int) error {
   for i, entry := range e.Entries {
      if entry.Id == id || id == -1 {
         e.Entries = append(e.Entries[:i], e.Entries[i+1:]...)
         return nil
      }
   }
   return errors.New("id not found")
}

func (e *Entries) Find(startDate, endDate, pattern, workoutType string) (*Entries, error) {
   var start, end time.Time
   loc := time.Now().Location()

   if startDate != "" {
      template, err := dateTemplate(startDate)
      if err != nil {
         return nil, err
      }
      start, err = time.ParseInLocation(template, startDate, loc)
      if err != nil {
         return nil, err
      }
   }

   if startDate != "" && endDate != "" {
      template, err := dateTemplate(endDate)
      if err != nil {
         return nil, err
      }
      end, err = time.ParseInLocation(template, endDate, loc)
      if err != nil {
         return nil, err
      }
      end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
   }

   if pattern == "*" || pattern == "" {
      pattern = ".*"
   }

   regex, err := regexp.Compile(pattern)
   if err != nil {
      return nil, err
   }

   useType := false
   if (len(workoutType) > 0) {
      useType = true
   }

   var matches Entries
   for _, entry := range e.Entries {
      if entry.Workout.Type == "" {
         continue
      }
      if useType &&  entry.Workout.Type != workoutType {
         continue
      }
      if startDate != "" {
         t, err := entry.Time()
         if err != nil {
            continue
         }
         t = t.Local()
         if t.Before(start) {
            continue
         } else if endDate != "" && t.After(end) {
            continue
         }
      }
      if regex.MatchString(entry.Workout.Title) || regex.MatchString(entry.Message) {
         matches.Entries = append(matches.Entries, entry)
      }
   }
   return &matches, nil
}

type dateTemplateTest struct {
   Regex *regexp.Regexp
   Template string
}

var dateTemplates = []dateTemplateTest {
   { regexp.MustCompile(`^\d{2}/\d{1,2}/\d{1,2}$`), "06/1/2" },
   { regexp.MustCompile(`^\d{2]/\d{1,2}$`), "06/1" },
}

func dateTemplate(date string) (string, error) {
   for _, tst := range dateTemplates {
      if tst.Regex.MatchString(date) {
         return tst.Template, nil
      }
   }
   return "", errors.New("Invalid date format")
}
