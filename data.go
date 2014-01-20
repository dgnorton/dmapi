package dmapi

import (
   "encoding/json"
   "errors"
   //"fmt"
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

func (e *Entries) Find(startDate, endDate, pattern string) (*Entries, error) {
   var start, end time.Time
   if startDate != "" {
      template, err := dateTemplate(startDate)
      if err != nil {
         return nil, err
      }
      start, err = time.Parse(template, startDate)
      if err != nil {
         return nil, err
      }
   }

   if startDate != "" && endDate != "" {
      template, err := dateTemplate(endDate)
      if err != nil {
         return nil, err
      }
      end, err = time.Parse(template, endDate)
      if err != nil {
         return nil, err
      }
   }

   if pattern == "*" || pattern == "" {
      pattern = ".*"
   }

   regex, err := regexp.Compile(pattern)
   if err != nil {
      return nil, err
   }

   var matches Entries
   for _, entry := range e.Entries {
      if startDate != "" {
         t, err := entry.Time()
         if err != nil {
            continue
         }
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
