import { watchEffect } from 'vue'
import { siteProfile } from '../data/site'

export interface PageMeta {
  /** Page title without the site suffix; omit on the home page. */
  title?: string
  description?: string
  path: string
  type?: 'website' | 'article'
  noindex?: boolean
  structuredData?: Record<string, unknown>
}

const HOME_TITLE = '零玖肆玖肆｜技术记录与工程思考'

export const siteBaseUrl = () => (import.meta.env.VITE_SITE_URL || window.location.origin).replace(/\/$/, '')

const upsertMeta = (attribute: 'name' | 'property', key: string, content: string) => {
  let element = document.head.querySelector<HTMLMetaElement>(`meta[${attribute}="${key}"]`)
  if (!element) {
    element = document.createElement('meta')
    element.setAttribute(attribute, key)
    document.head.appendChild(element)
  }
  element.content = content
}

export function usePageMeta(source: () => PageMeta) {
  watchEffect(() => {
    const meta = source()
    const title = meta.title ? `${meta.title}｜${siteProfile.name}` : HOME_TITLE
    const description = meta.description || siteProfile.description
    const url = `${siteBaseUrl()}${meta.path}`

    document.title = title
    upsertMeta('name', 'description', description)
    upsertMeta('name', 'robots', meta.noindex ? 'noindex,follow' : 'index,follow')
    upsertMeta('property', 'og:type', meta.type ?? 'website')
    upsertMeta('property', 'og:title', title)
    upsertMeta('property', 'og:description', description)
    upsertMeta('property', 'og:url', url)

    let canonical = document.head.querySelector<HTMLLinkElement>('link[rel="canonical"]')
    if (!canonical) {
      canonical = document.createElement('link')
      canonical.rel = 'canonical'
      document.head.appendChild(canonical)
    }
    canonical.href = url

    let structuredData = document.head.querySelector<HTMLScriptElement>('#blog-structured-data')
    if (!meta.structuredData) {
      structuredData?.remove()
      return
    }
    if (!structuredData) {
      structuredData = document.createElement('script')
      structuredData.id = 'blog-structured-data'
      structuredData.type = 'application/ld+json'
      document.head.appendChild(structuredData)
    }
    structuredData.textContent = JSON.stringify({ '@context': 'https://schema.org', ...meta.structuredData })
  })
}
