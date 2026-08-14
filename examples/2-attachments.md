---
subject: example message with attachments
from: Test Author <joedoe@test.com>
to: to@test.com
attachments:
  - ../internal/testdata/image/cat.jpg
  - ../internal/testdata/image/chamillion.jpg
  - ../internal/testdata/image/panda.jpg
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

Image files in templates are likewise included as inline attachments.

## Creative Commons Credits

- **cat.jpg:** Cindy vanHeerden.
- **panda.jpg:** Snow Chang.
- **camillion.jpg:** Regan Dsouza.
