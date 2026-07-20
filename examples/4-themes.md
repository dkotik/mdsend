---
subject: example using themes
from: Test Author <joedoe@test.com>
to: to@test.com
theme:
  font_size: 13px
  font_family: "'Courier New', verdana"
  color:
    text: orange
    heading: "#667799"
    link: aqua
    action: red
    blockquote: "#efefef"
    border: "green" 
    shadow: gray 
---

# Styling Your Letters

Markdown text can be stylized with `theme` frontmatter section
and with templates.

Templates are demonstrated in the next example. Default templates
rely on the same `theme` frontmatter values.

## Fonts

Use only standard font names. Electronic mail does not display custom fonts well. Specify `font_size` in pixels only.

## Colors

- **text:** the text color in paragraphs.
- **heading:** the text color of headings.
- **link:** the text color of anchors.
- **action:** the color of action button. A paragraph with one link, by itself, is rendered as an action button.
- **blockquote:** the background color of a block quote, table alt row background, inline code element.
- **border:** thematic break, borders of tables, code blocks, and blockquote margin.
