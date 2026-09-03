---
subject: example message with attachments
from: Test Author <joedoe@test.com>
domain: test.com
to: to@test.com
attachments:
  - ../internal/testdata/image/cat.jpg
  - name: "Chamillion"
    path: ../internal/testdata/image/chamillion.jpg
  - ../internal/testdata/image/panda.jpg
media_constraints:
  quality: 60%
  width: 240
  height: 120
---

# Attach Files to the Message

There are three ways to attach a file to a letter:

1. Frontmatter `attachments` list.
2. Inline Markdown content attachments.
3. Inline template attachments.

## Regular Attachments

List files under `attachments` frontmatter section.
Each can be named.

## Inline Attachments

Inline attachments are placed into message body, when linked like so:
[cat](../internal/media/testdata/cat.jpg "Cat Photo by Cindy vanHeerden").

Image files in HTML templates are likewise included as inline attachments.

Each inline attachment is assigned a unique content identifier
based on file hash and sender domain. Sender domain is taken from
`domain` frontmatter field or extracted out of `from` address.

## Media Constrains

Images are automatically compressed. Control the size
with `media_constraints` frontmatter section.

## Creative Commons Credits

- **cat.jpg:** Cindy vanHeerden.
- **panda.jpg:** Snow Chang.
- **camillion.jpg:** Regan Dsouza.
