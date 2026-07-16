---
subject: scheduling deliveries
from: Test Author <joedoe@test.com>
to: to@test.com
schedule:
  after: 2026-12-30 00:00
  delay: 30m
  step: 24h
  fluctuate: 1h
  expire: 1000h
---

# Schedule Messages for Later Delivery

Queued messages can be delayed by providing `schedule` frontmatter instructions:

- **after:** time until the message can be sent.
  Must be in ISO format `YYYY-MM-DD hh:mm`.
- **delay:** time duration added to `after` or current time.
- **step:** time duration added for each recipient.
- **fluctuate:** normalized random time duration added for each
  mailing step, including the first message.
- **expire:** time duration, after all the messages have been sent,
  to keep the letter and message records in the database before they
  are deleted.

## Time Duration Units

- `m`: Minute
- `h`: Hour
- `d`: Day
- `w`: Week

Units can be combined together. For example, to delay the message
by eight weeks, four days, and seven hours, add to the front matter:

```yaml
schedule:
  delay: 8w4d7h
```
