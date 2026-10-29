import { describe, expect, it } from 'vitest'
import { buildCcSwitchImportDeeplink } from '../ccswitchImport'

describe('Antigravity CC Switch endpoint', () => {
  it.each([
    ['https://api.example.com', 'https://api.example.com/antigravity'],
    ['https://api.example.com/', 'https://api.example.com/antigravity'],
    ['https://api.example.com///', 'https://api.example.com/antigravity'],
    ['https://api.example.com/sub2api/', 'https://api.example.com/sub2api/antigravity']
  ])('joins the platform path onto %s for both clients', (baseUrl, endpoint) => {
    for (const clientType of ['claude', 'gemini'] as const) {
      const url = new URL(buildCcSwitchImportDeeplink({
        baseUrl, clientType, platform: 'antigravity', providerName: 'Sub2API',
        apiKey: 'sk-test', usageScript: 'return true'
      }))
      expect(url.searchParams.get('endpoint')).toBe(endpoint)
      expect(url.searchParams.get('app')).toBe(clientType)
      expect(url.searchParams.get('homepage')).toBe(baseUrl)
      expect(url.searchParams.has('model')).toBe(false)
    }
  })
})

describe('Antigravity CC Switch Codex import', () => {
  const build = (codexModel?: string) => new URL(buildCcSwitchImportDeeplink({
    baseUrl: 'https://api.example.com/',
    clientType: 'codex',
    platform: 'antigravity',
    providerName: 'Sub2API',
    apiKey: 'sk-test',
    usageScript: 'return true',
    codexModel
  }))

  it('imports Codex against the root gateway endpoint, not /antigravity', () => {
    const url = build('gemini-3.8-flash')
    expect(url.searchParams.get('app')).toBe('codex')
    // Codex speaks /v1/responses on the root path; /antigravity would 404.
    expect(url.searchParams.get('endpoint')).toBe('https://api.example.com')
    expect(url.searchParams.get('model')).toBe('gemini-3.8-flash')
  })

  it('falls back to the antigravity default model when no model is supplied', () => {
    expect(build().searchParams.get('model')).toBe('claude-sonnet-4-6')
    expect(build('   ').searchParams.get('model')).toBe('claude-sonnet-4-6')
  })
})
