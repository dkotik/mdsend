---
subject: example message with attachments
from: Test Author <joedoe@test.com>
to: to@test.com
attachments:
  - ../internal/media/testdata/cat.jpg
  - ../internal/media/testdata/chamillion.jpg
  - ../internal/media/testdata/panda.jpg
media_constraints:
  quality: 60%
  width: 240
  height: 120
---

# Attach Files to the Message

List added files under `attachments` frontmatter section.
Images are automatically compressed. Control the size
with `media_constraints` frontmatter section.

## Inline Attachments

Inline attachments are placed into message body,
when linked like so:
[cat](../internal/media/testdata/cat.jpg "Cat Photo by Cindy vanHeerden").

## Creative Commons Credits

- **cat.jpg:** Cindy vanHeerden.
- **panda.jpg:** Snow Chang.
- **camillion.jpg:** Regan Dsouza.
