import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'mcp-setu',
  description: 'MCP Bridge for Ollama - Interactive tool-calling chats with local language models',
  base: '/mcp-setu/',
  cleanUrls: true,
  head: [],

  themeConfig: {    search: {
      provider: 'local',
    },

    nav: [
      { text: 'Guide', link: '/getting-started' },
      { text: 'CLI Reference', link: '/cli/' },
      { text: 'GitHub', link: 'https://github.com/manojshevate/mcp-setu' },
    ],

    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Introduction', link: '/' },
          { text: 'Quick Start', link: '/getting-started' },
          { text: 'Installation', link: '/installation' },
        ],
      },
      {
        text: 'Usage',
        items: [
          { text: 'Configuration', link: '/configuration' },
          { text: 'Examples', link: '/examples' },
          { text: 'Troubleshooting', link: '/troubleshooting' },
        ],
      },
      {
        text: 'Deep Dive',
        items: [
          { text: 'Concepts', link: '/concepts' },
          { text: 'CLI Reference', link: '/cli/' },
        ],
      },
      {
        text: 'Contributing',
        items: [
          { text: 'Development', link: '/development' },
        ],
      },
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/manojshevate/mcp-setu' },
    ],

    footer: {
      message: 'Released under MIT License',
      copyright: 'Copyright © 2026 mcp-setu contributors',
    },
  },

  markdown: {
    lineNumbers: true,
  },
})
