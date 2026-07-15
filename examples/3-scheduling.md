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
- **delay:** time duration added to `after` or current time.
- **step:** time duration added for each recipient.
- **fluctuate:** normalized random time duration added for each mailing step, including the first message.
- **expire:** time duration, after all the messages have been sent, to keep the letter and message records in the database before they are deleted.

## Context

Template execution context contains the following fields:

- **Frontmatter:** the raw Markdown frontmatter as `map[string]any`.
- **Recipient:** the raw message target as `map[string]any`. It
  will contain `name` and `email` fields and any other fields
  fromt the contact source.
- **Content:** rendered Markdown body.
- **Schedule:**

## Reify Templates

Special `reify` function stores the output of a
