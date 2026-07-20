---
subject: extend one letter from letter
from: Test Author <joedoe@test.com>
to: extends@test.com
extends: 6-list.md
---

# Extend One Letter with Another

Merge the content and frontmatter fields into a letter from another
file by using the `extends` field.

```yaml
extends:
  # this letter will inherit the content
  # and configuration values of `template.md`
  - template.md
  # load a configuration file
  - config.yaml
```

Any Markdown content below the last horizontal rule is added as a
footer to the current letter. You may nest the extensions deep as
long as there are no circular references.
