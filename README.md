# zlog

## What is a log tool I really want

I want to a log tool which provide help for reading through others codes,
especially a big bounch of codes.

So the tool should like:
- It helps to show how the code runs
- Pretty showing key arguments
- Help to analyze data flow

And it should like below:

```
>>>> something is wrong                                         (app)[2018-07-09 18:07:57]
     - a=1
     - b=2
     - because the weather is not good
**** something is not working well                              (app/module1)[2018-07-09 18:07:57]
     - some one tole me the sun go  into cloud

---- This is normal output                                      (app/module2)[2018-07-09 18:07:57]
     {
       "a":1,
       "b":"abc",
     }
```

## API Example

The simplest way to use zlog is simply the package-level exported logger:

```go
package main

import (
  log "github.com/ssor/zlog"
)

func main() {
  log.WithFields(log.Fields{
    "animal": "walrus",
  }).Info("A walrus appears")
}
```

Note that it's completely api-compatible with the stdlib logger, so you can
replace your `log` imports everywhere with `log "github.com/ssor/zlog"`
and you'll now have the flexibility of Logrus. You can customize it all you
want:

```go
package main

import (
  "os"
  log "github.com/ssor/zlog"
)

func init() {
  // Output to stderr instead of stdout, could also be a file.
  log.SetOutput(os.Stderr)

  // Only log the warning severity or above.
  log.SetLevel(log.WarnLevel)
}

func main() {
  log.WithFields(log.Fields{
    "animal": "walrus",
    "size":   10,
  }).Info("A group of walrus emerges from the ocean")

  log.WithFields(log.Fields{
    "omg":    true,
    "number": 122,
  }).Warn("The group's number increased tremendously!")

  log.WithFields(log.Fields{
    "omg":    true,
    "number": 100,
  }).Fatal("The ice breaks!")

  // A common pattern is to re-use fields between logging statements by re-using
  contextLogger := log.WithFields(log.Fields{
    "common": "this is a common field",
    "other": "I also should be logged always",
  })

  contextLogger.Info("I'll be logged with common and other field")
  contextLogger.Info("Me too")
}
```


#### Fields

Logrus encourages careful, structured logging though logging fields instead of
long, unparseable error messages. For example, instead of: `log.Fatalf("Failed
to send event %s to topic %s with key %d")`, you should log the much more
discoverable:

```go
log.WithFields(log.Fields{
  "event": event,
  "topic": topic,
  "key": key,
}).Fatal("Failed to send event")
```

We've found this API forces you to think about logging in a way that produces
much more useful logging messages. We've been in countless situations where just
a single added field to a log statement that was already there would've saved us
hours. The `WithFields` call is optional.

