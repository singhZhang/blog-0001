import { Marked } from 'marked'

/** @typedef {{ id: string, depth: number, text: string }} MarkdownHeading */

/** @param {string} text */
const toPlainText = (text) => text
  .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
  .replace(/[`*_~]/g, '')
  .trim()

/**
 * Output is unsanitized HTML; sanitize it before inserting into the DOM.
 * @param {string} source
 * @returns {{ html: string, headings: MarkdownHeading[] }}
 */
export function renderMarkdown(source) {
  /** @type {MarkdownHeading[]} */
  const headings = []
  const marked = new Marked({
    gfm: true,
    renderer: {
      heading({ tokens, depth, text }) {
        const content = this.parser.parseInline(tokens)
        // ASCII ids keep hash links stable regardless of heading language.
        const id = `section-${headings.length + 1}`
        headings.push({ id, depth, text: toPlainText(text) })
        return `<h${depth} id="${id}">${content}</h${depth}>\n`
      },
    },
  })

  return { html: /** @type {string} */ (marked.parse(source || '', { async: false })), headings }
}
