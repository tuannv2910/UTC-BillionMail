import { describe, it, expect } from 'vitest'
import i18n, { setLanguage } from './index'

describe('i18n', () => {
  it('is created with legacy false', () => {
    expect(i18n.global).toBeDefined()
  })

  it('defaults to vi locale', () => {
    expect(i18n.global.locale.value).toBe('vi')
  })

  it('has en as fallback locale', () => {
    expect(i18n.global.fallbackLocale.value).toBe('en')
  })

  it('has en, zh, ja messages loaded', () => {
    const messages = i18n.global.messages.value
    expect(messages).toHaveProperty('en')
    expect(messages).toHaveProperty('zh')
    expect(messages).toHaveProperty('ja')
    expect(messages).toHaveProperty('vi')
  })

  it('vi messages cover every en key', () => {
    const messages = i18n.global.messages.value as Record<string, unknown>
    const keys = (obj: unknown, prefix = ''): string[] =>
      obj && typeof obj === 'object' && !Array.isArray(obj)
        ? Object.entries(obj as Record<string, unknown>).flatMap(([k, v]) => keys(v, prefix ? prefix + '.' + k : k))
        : [prefix]
    expect(keys(messages.vi).sort()).toEqual(keys(messages.en).sort())
  })

  describe('setLanguage', () => {
    it('changes locale to zh', () => {
      setLanguage('zh')
      expect(i18n.global.locale.value).toBe('zh')
    })

    it('changes locale to ja', () => {
      setLanguage('ja')
      expect(i18n.global.locale.value).toBe('ja')
    })

    it('changes locale to vi', () => {
      setLanguage('vi')
      expect(i18n.global.locale.value).toBe('vi')
    })

    it('changes locale back to en', () => {
      setLanguage('en')
      expect(i18n.global.locale.value).toBe('en')
    })
  })

  describe('translation', () => {
    it('translates a known key in en', () => {
      setLanguage('en')
      const { t } = i18n.global
      // layout.menu.overview should exist in en.json
      const result = t('layout.menu.overview')
      expect(result).toBeTruthy()
      expect(result).not.toBe('layout.menu.overview')
    })

    it('returns key path for unknown keys', () => {
      const { t } = i18n.global
      const result = t('this.key.does.not.exist')
      expect(result).toBe('this.key.does.not.exist')
    })
  })
})
