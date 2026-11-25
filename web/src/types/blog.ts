export interface Article {
  id: string
  slug: string
  title: string
  excerpt: string
  publishedAt: string
  category: string
  readMinutes: number
  tags: string[]
  featuredCode?: string
  /** Markdown body. List endpoints may omit it; the detail endpoint should include it. */
  content?: string
}

export interface SiteProfile {
  name: string
  description: string
  email: string
}

export interface Category {
  name: string
  description: string
}

export interface AboutProfile {
  intro: string[]
  writing: { title: string; description: string }[]
  principles: string[]
  colophon: string[]
}
