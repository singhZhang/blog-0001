import type { AboutProfile, SiteProfile } from '../types/blog'

export const siteProfile: SiteProfile = {
  name: '零玖肆玖肆',
  description: '记录工程实践、系统设计与产品思考，把解决问题的过程沉淀成可以再次使用的经验。',
  email: 'hello@09494.dev',
}

export const aboutProfile: AboutProfile = {
  intro: [
    '你好，这里是零玖肆玖肆。一个写给自己、也写给同路人的地方。',
    '白天写代码，晚上整理笔记。遇到的问题、读过的书、日常里冒出来的念头，都会慢慢沉淀在这里。比起"正确答案"，我更想记录得出答案的过程。',
  ],
  writing: [
    { title: '技术实践', description: '工作中真实踩过的坑，以及爬出来的路径。' },
    { title: '系统思考', description: '架构、接口与边界，关于"为什么这样设计"。' },
    { title: '日常随想', description: '读书、生活和一些暂时没有结论的想法。' },
  ],
  principles: [
    '只写自己真正做过、想过的事情。',
    '写清楚上下文，让半年后的自己也能读懂。',
    '错了就改，并在文中保留修改的痕迹。',
  ],
  colophon: [
    '站点使用 Vue 3、TypeScript 与 Vite 构建，首页景色会跟随你所在时区的时间变化。',
    '正文使用系统字体渲染，不追踪访客，也没有广告。',
  ],
}
