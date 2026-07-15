---
subject: scheduling deliveries
from: Test Author <joedoe@test.com>
to: to@test.com
---

# Schedule Messages for Delivery



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
